package querier

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/go-kit/log/level"
	httpgrpc_server "example.com/acme/kit/httpgrpc/server"
	"example.com/acme/kit/ring"
	ring_client "example.com/acme/kit/ring/client"
	"example.com/acme/kit/services"
	"example.com/acme/kit/user"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/atomic"
	"go.uber.org/multierr"

	generator_client "example.com/acme/tracestore/modules/generator/client"
	ingester_client "example.com/acme/tracestore/modules/ingester/client"
	"example.com/acme/tracestore/modules/overrides"
	"example.com/acme/tracestore/modules/querier/worker"
	"example.com/acme/tracestore/modules/storage"
	"example.com/acme/tracestore/pkg/api"
	"example.com/acme/tracestore/pkg/collector"
	"example.com/acme/tracestore/pkg/model/trace"
	"example.com/acme/tracestore/pkg/search"
	"example.com/acme/tracestore/pkg/tracestorepb"
	"example.com/acme/tracestore/pkg/traceql"
	"example.com/acme/tracestore/pkg/traceqlmetrics"
	"example.com/acme/tracestore/pkg/util"
	"example.com/acme/tracestore/pkg/util/log"
	"example.com/acme/tracestore/pkg/validation"
	"example.com/acme/tracestore/tracestoredb/backend"
	"example.com/acme/tracestore/tracestoredb/encoding/common"
)

var tracer = otel.Tracer("modules/querier")

var (
	metricIngesterClients = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "tracestore",
		Name:      "querier_ingester_clients",
		Help:      "The current number of ingester clients.",
	})
	metricMetricsGeneratorClients = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "tracestore",
		Name:      "querier_metrics_generator_clients",
		Help:      "The current number of generator clients.",
	})
)

type (
	forEachFn          func(ctx context.Context, client tracestorepb.QuerierClient) error
	forEachGeneratorFn func(ctx context.Context, client tracestorepb.MetricsGeneratorClient) error
	replicationSetFn   func(r ring.ReadRing) (ring.ReplicationSet, error)
)

// Querier handlers queries.
type Querier struct {
	services.Service

	cfg Config

	ingesterPools []*ring_client.Pool
	ingesterRings []ring.ReadRing

	generatorPool *ring_client.Pool
	generatorRing ring.ReadRing

	engine *traceql.Engine
	store  storage.Store
	limits overrides.Interface

	subservices        *services.Manager
	subservicesWatcher *services.FailureWatcher
}

// New makes a new Querier.
func New(
	cfg Config,
	ingesterClientConfig ingester_client.Config,
	ingesterRings []ring.ReadRing,
	generatorClientConfig generator_client.Config,
	generatorRing ring.ReadRing,
	store storage.Store,
	limits overrides.Interface,
) (*Querier, error) {
	var ingesterClientFactory ring_client.PoolAddrFunc = func(addr string) (ring_client.PoolClient, error) {
		return ingester_client.New(addr, ingesterClientConfig)
	}

	var generatorClientFactory ring_client.PoolAddrFunc = func(addr string) (ring_client.PoolClient, error) {
		return generator_client.New(addr, generatorClientConfig)
	}

	ingesterPools := make([]*ring_client.Pool, 0, len(ingesterRings))
	for i, ring := range ingesterRings {
		pool := ring_client.NewPool(fmt.Sprintf("querier_pool_%d", i),
			ingesterClientConfig.PoolConfig,
			ring_client.NewRingServiceDiscovery(ring),
			ingesterClientFactory,
			metricIngesterClients,
			log.Logger)
		ingesterPools = append(ingesterPools, pool)
	}

	q := &Querier{
		cfg:           cfg,
		ingesterRings: ingesterRings,
		ingesterPools: ingesterPools,
		generatorRing: generatorRing,
		generatorPool: ring_client.NewPool("querier_to_generator_pool",
			generatorClientConfig.PoolConfig,
			ring_client.NewRingServiceDiscovery(generatorRing),
			generatorClientFactory,
			metricMetricsGeneratorClients,
			log.Logger),
		engine: traceql.NewEngine(),
		store:  store,
		limits: limits,
	}

	q.Service = services.NewBasicService(q.starting, q.running, q.stopping)
	return q, nil
}

func (q *Querier) CreateAndRegisterWorker(handler http.Handler) error {
	q.cfg.Worker.MaxConcurrentRequests = q.cfg.MaxConcurrentQueries
	worker, err := worker.NewQuerierWorker(
		q.cfg.Worker,
		httpgrpc_server.NewServer(handler),
		log.Logger,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create frontend worker: %w", err)
	}

	subservices := []services.Service{worker, q.generatorPool}
	for _, pool := range q.ingesterPools {
		subservices = append(subservices, pool)
	}
	err = q.RegisterSubservices(subservices...)
	if err != nil {
		return fmt.Errorf("failed to register generator pool sub-service: %w", err)
	}

	return nil
}

func (q *Querier) RegisterSubservices(s ...services.Service) error {
	var err error
	q.subservices, err = services.NewManager(s...)
	q.subservicesWatcher = services.NewFailureWatcher()
	q.subservicesWatcher.WatchManager(q.subservices)
	return err
}

func (q *Querier) starting(ctx context.Context) error {
	if q.subservices != nil {
		err := services.StartManagerAndAwaitHealthy(ctx, q.subservices)
		if err != nil {
			return fmt.Errorf("failed to start subservices: %w", err)
		}
	}

	return nil
}

func (q *Querier) running(ctx context.Context) error {
	if q.subservices != nil {
		select {
		case <-ctx.Done():
			return nil
		case err := <-q.subservicesWatcher.Chan():
			return fmt.Errorf("subservices failed: %w", err)
		}
	} else {
		<-ctx.Done()
	}
	return nil
}

func (q *Querier) stopping(_ error) error {
	if q.subservices != nil {
		return services.StopManagerAndAwaitStopped(context.Background(), q.subservices)
	}
	return nil
}

// FindTraceByID implements tracestorepb.Querier.
func (q *Querier) FindTraceByID(ctx context.Context, req *tracestorepb.TraceByIDRequest, timeStart int64, timeEnd int64) (*tracestorepb.TraceByIDResponse, error) {
	if !validation.ValidTraceID(req.TraceID) {
		return nil, errors.New("invalid trace id")
	}

	userID, err := user.ExtractOrgID(ctx)
	if err != nil {
		return nil, fmt.Errorf("error extracting org id in Querier.FindTraceByID: %w", err)
	}

	ctx, span := tracer.Start(ctx, "Querier.FindTraceByID")
	defer span.End()

	span.SetAttributes(attribute.String("queryMode", req.QueryMode))

	maxBytes := q.limits.MaxBytesPerTrace(userID)
	combiner := trace.NewCombiner(maxBytes, req.AllowPartialTrace)
	mc := collector.NewMetricsCollector()

	if req.QueryMode == QueryModeIngesters || req.QueryMode == QueryModeAll {
		var getRSFn replicationSetFn
		if q.cfg.QueryRelevantIngesters {
			traceKey := util.TokenFor(userID, req.TraceID)
			getRSFn = func(r ring.ReadRing) (ring.ReplicationSet, error) {
				return r.Get(traceKey, ring.Read, nil, nil, nil)
			}
		}
		var spanCountTotal, traceCountTotal atomic.Int64
		var found atomic.Bool

		// get responses from all ingesters in parallel
		span.AddEvent("searching ingesters")
		forEach := func(funcCtx context.Context, client tracestorepb.QuerierClient) error {
			resp, err := client.FindTraceByID(funcCtx, req)
			if err != nil {
				return err
			}
			t := resp.Trace
			if t != nil {
				// we found a trace, consume and count it
				spanCount, err := combiner.Consume(t)
				if err != nil {
					return err
				}
				spanCountTotal.Add(int64(spanCount))
				traceCountTotal.Inc()
				found.Store(true)
				if resp.Metrics != nil {
					mc.Add(resp.Metrics.InspectedBytes)
				}
			}
			return nil
		}
		err := q.forIngesterRings(ctx, userID, getRSFn, forEach)
		if err != nil {
			return nil, fmt.Errorf("error querying ingesters in Querier.FindTraceByID: %w", err)
		}
		span.AddEvent("done searching ingesters", oteltrace.WithAttributes(
			attribute.Bool("found", found.Load()),
			attribute.Int64("combinedSpans", spanCountTotal.Load()),
			attribute.Int64("combinedTraces", traceCountTotal.Load())))
	}

	if req.QueryMode == QueryModeBlocks || req.QueryMode == QueryModeAll {
		span.AddEvent("searching store", oteltrace.WithAttributes(
			attribute.Int64("timeStart", timeStart),
			attribute.Int64("timeEnd", timeEnd),
		))

		opts := common.DefaultSearchOptionsWithMaxBytes(maxBytes)
		opts.BlockReplicationFactor = backend.DefaultReplicationFactor
		partialTraces, blockErrs, err := q.store.Find(ctx, userID, req.TraceID, req.BlockStart, req.BlockEnd, timeStart, timeEnd, opts)
		if err != nil {
			retErr := fmt.Errorf("error querying store in Querier.FindTraceByID: %w", err)
			span.RecordError(retErr)
			return nil, retErr
		}

		if len(blockErrs) > 0 {
			return nil, multierr.Combine(blockErrs...)
		}

		span.AddEvent("done searching store", oteltrace.WithAttributes(
			attribute.Int("foundPartialTraces", len(partialTraces))))

		for _, partialTrace := range partialTraces {
			if partialTrace == nil {
				continue
			}
			_, err = combiner.Consume(partialTrace.Trace)
			if err != nil {
				return nil, err
			}
			if partialTrace.Metrics != nil {
				mc.Add(partialTrace.Metrics.InspectedBytes)
			}
		}
	}

	completeTrace, _ := combiner.Result()
	resp := &tracestorepb.TraceByIDResponse{
		Trace:   completeTrace,
		Metrics: &tracestorepb.TraceByIDMetrics{InspectedBytes: mc.TotalValue()},
	}

	if combiner.IsPartialTrace() {
		resp.Status = tracestorepb.TraceByIDResponse_PARTIAL
		resp.Message = fmt.Sprintf("Trace exceeds maximum size of %d bytes, a partial trace is returned", maxBytes)
	}

	return resp, nil
}

// forIngesterRings runs f, in parallel, for given ingesters
func (q *Querier) forIngesterRings(ctx context.Context, userID string, getReplicationSet replicationSetFn, f forEachFn) error {
	if ctx.Err() != nil {
		_ = level.Debug(log.Logger).Log("forIngesterRings context error", "ctx.Err()", ctx.Err().Error())
		return ctx.Err()
	}

	// if we have no configured ingester rings this will fail silently. let's return an actual error instead
	if len(q.ingesterRings) == 0 {
		return errors.New("forIngesterRings: no ingester rings configured")
	}

	// if a nil replicationSetFn is passed, that means to just use a standard Read ring
	if getReplicationSet == nil {
		getReplicationSet = func(r ring.ReadRing) (ring.ReplicationSet, error) {
			return r.GetReplicationSetForOperation(ring.Read)
		}
	}

	var mtx sync.Mutex
	var wg sync.WaitGroup

	var responseErr error

	for i, ingesterRing := range q.ingesterRings {
		if q.cfg.ShuffleShardingIngestersEnabled {
			ingesterRing = ingesterRing.ShuffleShardWithLookback(
				userID,
				q.limits.IngestionTenantShardSize(userID),
				q.cfg.ShuffleShardingIngestersLookbackPeriod,
				time.Now(),
			)
		}

		replicationSet, err := getReplicationSet(ingesterRing)
		if err != nil {
			return fmt.Errorf("forIngesterRings: error getting replication set for ring (%d): %w", i, err)
		}
		pool := q.ingesterPools[i]

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := forOneIngesterRing(ctx, replicationSet, f, pool, q.cfg.ExtraQueryDelay)
			mtx.Lock()
			defer mtx.Unlock()

			if err != nil {
				responseErr = multierr.Combine(responseErr, err)
				return
			}
		}()
	}

	wg.Wait()

	if responseErr != nil {
		return responseErr
	}

	return nil
}

func forOneIngesterRing(ctx context.Context, replicationSet ring.ReplicationSet, f forEachFn, pool *ring_client.Pool, extraQueryDelay time.Duration) error {
	ctx, span := tracer.Start(ctx, "Querier.forOneIngesterRing")
	defer span.End()

	doFunc := func(funcCtx context.Context, ingester *ring.InstanceDesc) (interface{}, error) {
		if funcCtx.Err() != nil {
			_ = level.Warn(log.Logger).Log("funcCtx.Err()", funcCtx.Err().Error())
			return nil, funcCtx.Err()
		}

		client, err := pool.GetClientFor(ingester.Addr)
		if err != nil {
			return nil, fmt.Errorf("failed to get client for %s: %w", ingester.Addr, err)
		}

		err = f(funcCtx, client.(tracestorepb.QuerierClient))
		if err != nil {
			return nil, fmt.Errorf("failed to execute f() for %s: %w", ingester.Addr, err)
		}

		// we are returning the empty response here because response is collected by
		// the collector inside forEachFn
		return nil, nil
	}

	// ignore response because it's nil, and we are using a collector inside forEachFn to
	// collect the actual response. we need to return nil here and ignore it
	// because doFunc expects us to return a response
	_, err := replicationSet.Do(ctx, extraQueryDelay, doFunc)

	return err
}

// forGivenGenerators runs f, in parallel, for given generators
func (q *Querier) forGivenGenerators(ctx context.Context, replicationSet ring.ReplicationSet, f forEachGeneratorFn) error {
	if ctx.Err() != nil {
		_ = level.Debug(log.Logger).Log("foreGivenGenerators context error", "ctx.Err()", ctx.Err().Error())
		return ctx.Err()
	}

	ctx, span := tracer.Start(ctx, "Querier.forGivenGenerators")
	defer span.End()

	doFunc := func(funcCtx context.Context, generator *ring.InstanceDesc) (interface{}, error) {
		if funcCtx.Err() != nil {
			_ = level.Warn(log.Logger).Log("funcCtx.Err()", funcCtx.Err().Error())
			return nil, funcCtx.Err()
		}

		client, err := q.generatorPool.GetClientFor(generator.Addr)
		if err != nil {
			return nil, fmt.Errorf("failed to get client for %s: %w", generator.Addr, err)
		}

		err = f(funcCtx, client.(tracestorepb.MetricsGeneratorClient))
		if err != nil {
			return nil, fmt.Errorf("failed to execute f() for %s: %w", generator.Addr, err)
		}

		// we are returning the empty response here because response is collected by
		// the collector inside forEachGeneratorFn
		return nil, nil
	}

	// ignore response because it's nil, and we are using a collector inside forEachGeneratorFn to
	// collect the actual response. we need to return nil here and ignore it
	// because doFunc expects us to return a response
	_, err := replicationSet.Do(ctx, q.cfg.ExtraQueryDelay, doFunc)
	if err != nil {
		return fmt.Errorf("failed to get response from generators: %w", err)
	}

	return nil
}

func (q *Querier) SearchRecent(ctx context.Context, req *tracestorepb.SearchRequest) (*tracestorepb.SearchResponse, error) {
	userID, err := user.ExtractOrgID(ctx)
	if err != nil {
		return nil, fmt.Errorf("error extracting org id in Querier.Search: %w", err)
	}

	var results []*tracestorepb.SearchResponse
	mtx := sync.Mutex{}

	forEach := func(ctx context.Context, client tracestorepb.QuerierClient) error {
		resp, err := client.SearchRecent(ctx, req)
		if err != nil {
			return err
		}
		mtx.Lock()
		defer mtx.Unlock()
		results = append(results, resp)
		return nil
	}
	err = q.forIngesterRings(ctx, userID, nil, forEach)
	if err != nil {
		return nil, fmt.Errorf("error querying ingesters in Querier.Search: %w", err)
	}

	return q.postProcessIngesterSearchResults(req, results), nil
}

func (q *Querier) SearchTagsBlocks(ctx context.Context, req *tracestorepb.SearchTagsBlockRequest) (*tracestorepb.SearchTagsResponse, error) {
	v2Response, err := q.internalTagsSearchBlockV2(ctx, req)
	if err != nil {
		return nil, err
	}

	distinctValues := collector.NewDistinctString(0, 0, 0)

	// flatten v2 response
	for _, s := range v2Response.Scopes {
		for _, t := range s.Tags {
			distinctValues.Collect(t)
			if distinctValues.Exceeded() {
				break // stop early
			}
		}
	}

	return &tracestorepb.SearchTagsResponse{
		TagNames: distinctValues.Strings(),
		Metrics:  v2Response.Metrics,
	}, nil
}

func (q *Querier) SearchTagValuesBlocks(ctx context.Context, req *tracestorepb.SearchTagValuesBlockRequest) (*tracestorepb.SearchTagValuesResponse, error) {
	return q.internalTagValuesSearchBlock(ctx, req)
}

func (q *Querier) SearchTagsBlocksV2(ctx context.Context, req *tracestorepb.SearchTagsBlockRequest) (*tracestorepb.SearchTagsV2Response, error) {
	return q.internalTagsSearchBlockV2(ctx, req)
}

func (q *Querier) SearchTagValuesBlocksV2(ctx context.Context, req *tracestorepb.SearchTagValuesBlockRequest) (*tracestorepb.SearchTagValuesV2Response, error) {
	return q.internalTagValuesSearchBlockV2(ctx, req)
}

func (q *Querier) SearchTags(ctx context.Context, req *tracestorepb.SearchTagsRequest) (*tracestorepb.SearchTagsResponse, error) {
	userID, err := user.ExtractOrgID(ctx)
	if err != nil {
		return nil, fmt.Errorf("error extracting org id in Querier.SearchTags: %w", err)
	}

	maxDataSize := q.limits.MaxBytesPerTagValuesQuery(userID)
	distinctValues := collector.NewDistinctString(maxDataSize, req.MaxTagsPerScope, req.StaleValuesThreshold)
	mc := collector.NewMetricsCollector()

	forEach := func(ctx context.Context, client tracestorepb.QuerierClient) error {
		resp, err := client.SearchTags(ctx, req)
		if err != nil {
			return err
		}
		// collect metrics first because we stop early with return
		if resp.Metrics != nil {
			mc.Add(resp.Metrics.InspectedBytes)
		}

		for _, tag := range resp.TagNames {
			distinctValues.Collect(tag)
			if distinctValues.Exceeded() {
				return nil // stop early
			}
		}
		return nil
	}
	err = q.forIngesterRings(ctx, userID, nil, forEach)
	if err != nil {
		return nil, fmt.Errorf("error querying ingesters in Querier.SearchTags: %w", err)
	}

	if distinctValues.Exceeded() {
		level.Warn(log.Logger).Log("msg", "size of tags in instance exceeded limit, reduce cardinality or size of tags", "userID", userID, "maxDataSize", maxDataSize, "size", distinctValues.Size())
	}

	return &tracestorepb.SearchTagsResponse{
		TagNames: distinctValues.Strings(),
		Metrics:  &tracestorepb.MetadataMetrics{InspectedBytes: mc.TotalValue()},
	}, nil
}

func (q *Querier) SearchTagsV2(ctx context.Context, req *tracestorepb.SearchTagsRequest) (*tracestorepb.SearchTagsV2Response, error) {
	orgID, err := user.ExtractOrgID(ctx)
	if err != nil {
		return nil, fmt.Errorf("error extracting org id in Querier.SearchTags: %w", err)
	}

	maxBytesPerTag := q.limits.MaxBytesPerTagValuesQuery(orgID)
	distinctValues := collector.NewScopedDistinctString(maxBytesPerTag, req.MaxTagsPerScope, req.StaleValuesThreshold)
	mc := collector.NewMetricsCollector()

	// Get results from all ingesters
	forEach := func(ctx context.Context, client tracestorepb.QuerierClient) error {
		resp, err := client.SearchTagsV2(ctx, req)
		if err != nil {
			return err
		}
		// collect metrics first because we stop early with return
		if resp.Metrics != nil {
			mc.Add(resp.Metrics.InspectedBytes)
		}

		for _, res := range resp.Scopes {
			for _, tag := range res.Tags {
				if distinctValues.Collect(res.Name, tag) {
					return nil
				}
			}
		}
		return nil
	}

	err = q.forIngesterRings(ctx, orgID, nil, forEach)
	if err != nil {
		return nil, fmt.Errorf("error querying ingesters in Querier.SearchTags: %w", err)
	}

	if distinctValues.Exceeded() {
		level.Warn(log.Logger).Log("msg", "Search tags exceeded limit, reduce cardinality or size of tags", "orgID", orgID, "stopReason", distinctValues.StopReason())
	}

	collected := distinctValues.Strings()
	resp := &tracestorepb.SearchTagsV2Response{
		Scopes:  make([]*tracestorepb.SearchTagsV2Scope, 0, len(collected)),
		Metrics: &tracestorepb.MetadataMetrics{InspectedBytes: mc.TotalValue()}, // send metrics with response
	}
	for scope, vals := range collected {
		resp.Scopes = append(resp.Scopes, &tracestorepb.SearchTagsV2Scope{
			Name: scope,
			Tags: vals,
		})
	}

	return resp, nil
}

func (q *Querier) SearchTagValues(ctx context.Context, req *tracestorepb.SearchTagValuesRequest) (*tracestorepb.SearchTagValuesResponse, error) {
	userID, err := user.ExtractOrgID(ctx)
	if err != nil {
		return nil, fmt.Errorf("error extracting org id in Querier.SearchTagValues: %w", err)
	}

	maxDataSize := q.limits.MaxBytesPerTagValuesQuery(userID)
	distinctValues := collector.NewDistinctString(maxDataSize, req.MaxTagValues, req.StaleValueThreshold)
	mc := collector.NewMetricsCollector()

	// Virtual tags values. Get these first.
	for _, v := range search.GetVirtualTagValues(req.TagName) {
		// virtual tags are small so no need to stop early here
		distinctValues.Collect(v)
	}

	forEach := func(ctx context.Context, client tracestorepb.QuerierClient) error {
		resp, err := client.SearchTagValues(ctx, req)
		if err != nil {
			return err
		}
		// add metrics first because we stop early with return
		if resp.Metrics != nil {
			mc.Add(resp.Metrics.InspectedBytes)
		}

		for _, res := range resp.TagValues {
			distinctValues.Collect(res)
			if distinctValues.Exceeded() {
				return nil
			}
		}
		return nil
	}

	err = q.forIngesterRings(ctx, userID, nil, forEach)
	if err != nil {
		return nil, fmt.Errorf("error querying ingesters in Querier.SearchTagValues: %w", err)
	}

	if distinctValues.Exceeded() {
		level.Warn(log.Logger).Log("msg", "Search of tag values exceeded limit, reduce cardinality or size of tags", "tag", req.TagName, "orgID", userID, "stopReason", distinctValues.StopReason())
	}

	return &tracestorepb.SearchTagValuesResponse{
		TagValues: distinctValues.Strings(),
		Metrics:   &tracestorepb.MetadataMetrics{InspectedBytes: mc.TotalValue()},
	}, nil
}

func (q *Querier) SearchTagValuesV2(ctx context.Context, req *tracestorepb.SearchTagValuesRequest) (*tracestorepb.SearchTagValuesV2Response, error) {
	userID, err := user.ExtractOrgID(ctx)
	if err != nil {
		return nil, fmt.Errorf("error extracting org id in Querier.SearchTagValues: %w", err)
	}

	maxDataSize := q.limits.MaxBytesPerTagValuesQuery(userID)
	distinctValues := collector.NewDistinctValue(maxDataSize, req.MaxTagValues, req.StaleValueThreshold, func(v tracestorepb.TagValue) int { return len(v.Type) + len(v.Value) })
	mc := collector.NewMetricsCollector()

	// Virtual tags values. Get these first.
	virtualVals := search.GetVirtualTagValuesV2(req.TagName)
	for _, v := range virtualVals {
		// no need to stop early here, virtual tags are small
		distinctValues.Collect(v)
	}

	// with v2 search we can confidently bail if GetVirtualTagValuesV2 gives us any hits. this doesn't work
	// in v1 search b/c intrinsic tags like "status" are conflated with attributes named "status"
	if virtualVals != nil {
		// no data was read to collect virtual tags so 0 bytesRead
		return valuesToV2Response(distinctValues, 0), nil
	}

	forEach := func(ctx context.Context, client tracestorepb.QuerierClient) error {
		// combine metrics as we get results from ingesters
		resp, err := client.SearchTagValuesV2(ctx, req)
		if err != nil {
			return err
		}
		// collect metrics first, we stop early with return
		if resp.Metrics != nil {
			mc.Add(resp.Metrics.InspectedBytes)
		}

		for _, res := range resp.TagValues {
			distinctValues.Collect(*res)
			if distinctValues.Exceeded() {
				return nil // stop early
			}
		}
		return nil
	}
	err = q.forIngesterRings(ctx, userID, nil, forEach)
	if err != nil {
		return nil, fmt.Errorf("error querying ingesters in Querier.SearchTagValues: %w", err)
	}

	if distinctValues.Exceeded() {
		_ = level.Warn(log.Logger).Log("msg", "Search of tag values exceeded limit, reduce cardinality or size of tags", "tag", req.TagName, "orgID", userID, "stopReason", distinctValues.StopReason())
	}

	return valuesToV2Response(distinctValues, mc.TotalValue()), nil
}

func (q *Querier) SpanMetricsSummary(
	ctx context.Context,
	req *tracestorepb.SpanMetricsSummaryRequest,
) (*tracestorepb.SpanMetricsSummaryResponse, error) {
	// userID, err := user.ExtractOrgID(ctx)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "error extracting org id in Querier.SpanMetricsSummary")
	// }

	// limit := q.limits.MaxBytesPerTagValuesQuery(userID)

	genReq := &tracestorepb.SpanMetricsRequest{
		Query:   req.Query,
		GroupBy: req.GroupBy,
		Start:   req.Start,
		End:     req.End,
		Limit:   0,
	}

	// Get results from all generators
	replicationSet, err := q.generatorRing.GetReplicationSetForOperation(ring.Read)
	if err != nil {
		return nil, fmt.Errorf("error finding generators in Querier.SpanMetricsSummary: %w", err)
	}

	var results []*tracestorepb.SpanMetricsResponse
	mtx := sync.Mutex{}

	forEach := func(ctx context.Context, client tracestorepb.MetricsGeneratorClient) error {
		resp, err := client.GetMetrics(ctx, genReq)
		if err != nil {
			return err
		}
		// collect the results from the generators in the pool
		mtx.Lock()
		defer mtx.Unlock()
		results = append(results, resp)
		return nil
	}
	err = q.forGivenGenerators(ctx, replicationSet, forEach)
	if err != nil {
		return nil, fmt.Errorf("error querying generators in Querier.SpanMetricsSummary: %w", err)
	}

	// Combine the results
	yyy := make(map[traceqlmetrics.MetricKeys]*traceqlmetrics.LatencyHistogram)
	xxx := make(map[traceqlmetrics.MetricKeys]*tracestorepb.SpanMetricsSummary)

	var h *traceqlmetrics.LatencyHistogram
	var s traceqlmetrics.MetricSeries
	for _, r := range results {
		for _, m := range r.Metrics {
			s = protoToMetricSeries(m.Series)
			k := s.MetricKeys()

			if _, ok := xxx[k]; !ok {
				xxx[k] = &tracestorepb.SpanMetricsSummary{Series: m.Series}
			}

			xxx[k].ErrorSpanCount += m.Errors

			var b [64]int
			for _, l := range m.GetLatencyHistogram() {
				// Reconstitude the bucket
				b[l.Bucket] += int(l.Count)
				// Add to the total
				xxx[k].SpanCount += l.Count
			}

			// Combine the histogram
			h = traceqlmetrics.New(b)
			if _, ok := yyy[k]; !ok {
				yyy[k] = h
			} else {
				yyy[k].Combine(*h)
			}
		}
	}

	for s, h := range yyy {
		xxx[s].P50 = h.Percentile(0.5)
		xxx[s].P90 = h.Percentile(0.9)
		xxx[s].P95 = h.Percentile(0.95)
		xxx[s].P99 = h.Percentile(0.99)
	}

	resp := &tracestorepb.SpanMetricsSummaryResponse{}
	for _, x := range xxx {
		resp.Summaries = append(resp.Summaries, x)
	}

	return resp, nil
}

func valuesToV2Response(distinctValues *collector.DistinctValue[tracestorepb.TagValue], bytesRead uint64) *tracestorepb.SearchTagValuesV2Response {
	resp := &tracestorepb.SearchTagValuesV2Response{
		Metrics: &tracestorepb.MetadataMetrics{InspectedBytes: bytesRead},
	}
	for _, v := range distinctValues.Values() {
		v2 := v
		resp.TagValues = append(resp.TagValues, &v2)
	}
	return resp
}

// SearchBlock searches the specified subset of the block for the passed tags.
func (q *Querier) SearchBlock(ctx context.Context, req *tracestorepb.SearchBlockRequest) (*tracestorepb.SearchResponse, error) {
	tenantID, err := user.ExtractOrgID(ctx)
	if err != nil {
		return nil, fmt.Errorf("error extracting org id in Querier.BackendSearch: %w", err)
	}

	blockID, err := backend.ParseUUID(req.BlockID)
	if err != nil {
		return nil, err
	}

	enc, err := backend.ParseEncoding(req.Encoding)
	if err != nil {
		return nil, err
	}

	dc, err := backend.DedicatedColumnsFromTracestorepb(req.DedicatedColumns)
	if err != nil {
		return nil, err
	}

	meta := &backend.BlockMeta{
		Version:          req.Version,
		TenantID:         tenantID,
		Encoding:         enc,
		Size_:            req.Size_,
		IndexPageSize:    req.IndexPageSize,
		TotalRecords:     req.TotalRecords,
		BlockID:          blockID,
		DataEncoding:     req.DataEncoding,
		FooterSize:       req.FooterSize,
		DedicatedColumns: dc,
	}

	opts := common.DefaultSearchOptions()
	opts.StartPage = int(req.StartPage)
	opts.TotalPages = int(req.PagesToSearch)
	opts.MaxBytes = q.limits.MaxBytesPerTrace(tenantID)

	if api.IsTraceQLQuery(req.SearchReq) {
		fetcher := traceql.NewSpansetFetcherWrapper(func(ctx context.Context, req traceql.FetchSpansRequest) (traceql.FetchSpansResponse, error) {
			return q.store.Fetch(ctx, meta, req, opts)
		})

		return q.engine.ExecuteSearch(ctx, req.SearchReq, fetcher)
	}

	return q.store.Search(ctx, meta, req.SearchReq, opts)
}

func (q *Querier) internalTagsSearchBlockV2(ctx context.Context, req *tracestorepb.SearchTagsBlockRequest) (*tracestorepb.SearchTagsV2Response, error) {
	// For the intrinsic scope there is nothing to do in the querier,
	// these are always added by the frontend.
	if req.SearchReq.Scope == api.ParamScopeIntrinsic {
		return &tracestorepb.SearchTagsV2Response{}, nil
	}

	tenantID, err := user.ExtractOrgID(ctx)
	if err != nil {
		return nil, fmt.Errorf("error extracting org id in Querier.BackendSearch: %w", err)
	}

	blockID, err := backend.ParseUUID(req.BlockID)
	if err != nil {
		return nil, err
	}

	enc, err := backend.ParseEncoding(req.Encoding)
	if err != nil {
		return nil, err
	}

	dc, err := backend.DedicatedColumnsFromTracestorepb(req.DedicatedColumns)
	if err != nil {
		return nil, err
	}

	meta := &backend.BlockMeta{
		Version:          req.Version,
		TenantID:         tenantID,
		Encoding:         enc,
		Size_:            req.Size_,
		IndexPageSize:    req.IndexPageSize,
		TotalRecords:     req.TotalRecords,
		BlockID:          blockID,
		DataEncoding:     req.DataEncoding,
		FooterSize:       req.FooterSize,
		DedicatedColumns: dc,
	}

	opts := common.DefaultSearchOptions()
	opts.StartPage = int(req.StartPage)
	opts.TotalPages = int(req.PagesToSearch)

	query := traceql.ExtractMatchers(req.SearchReq.Query)
	if traceql.IsEmptyQuery(query) {
		return q.store.SearchTags(ctx, meta, req, opts)
	}

	valueCollector := collector.NewScopedDistinctString(q.limits.MaxBytesPerTagValuesQuery(tenantID), req.MaxTagsPerScope, req.StaleValueThreshold)
	mc := collector.NewMetricsCollector()

	fetcher := traceql.NewTagNamesFetcherWrapper(func(ctx context.Context, req traceql.FetchTagsRequest, cb traceql.FetchTagsCallback) error {
		return q.store.FetchTagNames(ctx, meta, req, cb, mc.Add, common.DefaultSearchOptions())
	})

	scope := traceql.AttributeScopeFromString(req.SearchReq.Scope)
	if scope == traceql.AttributeScopeUnknown {
		return nil, fmt.Errorf("unknown scope: %s", req.SearchReq.Scope)
	}

	err = q.engine.ExecuteTagNames(ctx, scope, query, func(tag string, scope traceql.AttributeScope) bool {
		return valueCollector.Collect(scope.String(), tag)
	}, fetcher)
	if err != nil {
		return nil, err
	}

	if valueCollector.Exceeded() {
		level.Warn(log.Logger).Log("msg", "Search tags exceeded limit, reduce cardinality or size of tags", "orgID", tenantID, "stopReason", valueCollector.StopReason())
	}

	scopedVals := valueCollector.Strings()
	resp := &tracestorepb.SearchTagsV2Response{
		Scopes:  make([]*tracestorepb.SearchTagsV2Scope, 0, len(scopedVals)),
		Metrics: &tracestorepb.MetadataMetrics{InspectedBytes: mc.TotalValue()}, // send metrics with response
	}
	for scope, vals := range scopedVals {
		resp.Scopes = append(resp.Scopes, &tracestorepb.SearchTagsV2Scope{
			Name: scope,
			Tags: vals,
		})
	}

	return resp, nil
}

func (q *Querier) internalTagValuesSearchBlock(ctx context.Context, req *tracestorepb.SearchTagValuesBlockRequest) (*tracestorepb.SearchTagValuesResponse, error) {
	tenantID, err := user.ExtractOrgID(ctx)
	if err != nil {
		return &tracestorepb.SearchTagValuesResponse{}, fmt.Errorf("error extracting org id in Querier.BackendSearch: %w", err)
	}

	blockID, err := backend.ParseUUID(req.BlockID)
	if err != nil {
		return &tracestorepb.SearchTagValuesResponse{}, err
	}

	enc, err := backend.ParseEncoding(req.Encoding)
	if err != nil {
		return &tracestorepb.SearchTagValuesResponse{}, err
	}

	dc, err := backend.DedicatedColumnsFromTracestorepb(req.DedicatedColumns)
	if err != nil {
		return &tracestorepb.SearchTagValuesResponse{}, err
	}

	meta := &backend.BlockMeta{
		Version:          req.Version,
		TenantID:         tenantID,
		Encoding:         enc,
		Size_:            req.Size_,
		IndexPageSize:    req.IndexPageSize,
		TotalRecords:     req.TotalRecords,
		BlockID:          blockID,
		DataEncoding:     req.DataEncoding,
		FooterSize:       req.FooterSize,
		DedicatedColumns: dc,
	}

	opts := common.DefaultSearchOptions()
	opts.StartPage = int(req.StartPage)
	opts.TotalPages = int(req.PagesToSearch)

	resp, err := q.store.SearchTagValues(ctx, meta, req, opts)
	if err != nil {
		return &tracestorepb.SearchTagValuesResponse{}, err
	}

	return resp, nil
}

func (q *Querier) internalTagValuesSearchBlockV2(ctx context.Context, req *tracestorepb.SearchTagValuesBlockRequest) (*tracestorepb.SearchTagValuesV2Response, error) {
	tenantID, err := user.ExtractOrgID(ctx)
	if err != nil {
		return &tracestorepb.SearchTagValuesV2Response{}, fmt.Errorf("error extracting org id in Querier.BackendSearch: %w", err)
	}

	blockID, err := backend.ParseUUID(req.BlockID)
	if err != nil {
		return &tracestorepb.SearchTagValuesV2Response{}, err
	}

	enc, err := backend.ParseEncoding(req.Encoding)
	if err != nil {
		return &tracestorepb.SearchTagValuesV2Response{}, err
	}

	dc, err := backend.DedicatedColumnsFromTracestorepb(req.DedicatedColumns)
	if err != nil {
		return &tracestorepb.SearchTagValuesV2Response{}, err
	}

	meta := &backend.BlockMeta{
		Version:          req.Version,
		TenantID:         tenantID,
		Encoding:         enc,
		Size_:            req.Size_,
		IndexPageSize:    req.IndexPageSize,
		TotalRecords:     req.TotalRecords,
		BlockID:          blockID,
		DataEncoding:     req.DataEncoding,
		FooterSize:       req.FooterSize,
		DedicatedColumns: dc,
	}

	opts := common.DefaultSearchOptions()
	opts.StartPage = int(req.StartPage)
	opts.TotalPages = int(req.PagesToSearch)

	query := traceql.ExtractMatchers(req.SearchReq.Query)
	if traceql.IsEmptyQuery(query) {
		return q.store.SearchTagValuesV2(ctx, meta, req.SearchReq, opts)
	}

	tag, err := traceql.ParseIdentifier(req.SearchReq.TagName)
	if err != nil {
		return nil, err
	}

	valueCollector := collector.NewDistinctValue(q.limits.MaxBytesPerTagValuesQuery(tenantID),
		req.SearchReq.MaxTagValues, req.SearchReq.StaleValueThreshold,
		func(v tracestorepb.TagValue) int { return len(v.Type) + len(v.Value) })

	mc := collector.NewMetricsCollector()

	fetcher := traceql.NewTagValuesFetcherWrapper(func(ctx context.Context, req traceql.FetchTagValuesRequest, cb traceql.FetchTagValuesCallback) error {
		return q.store.FetchTagValues(ctx, meta, req, cb, mc.Add, opts)
	})

	err = q.engine.ExecuteTagValues(ctx, tag, query, traceql.MakeCollectTagValueFunc(valueCollector.Collect), fetcher)
	if err != nil {
		return nil, err
	}

	if valueCollector.Exceeded() {
		level.Warn(log.Logger).Log("msg", "Search tags exceeded limit, reduce cardinality or size of tags", "orgID", tenantID, "stopReason", valueCollector.StopReason())
	}

	return valuesToV2Response(valueCollector, mc.TotalValue()), nil
}

func (q *Querier) postProcessIngesterSearchResults(req *tracestorepb.SearchRequest, results []*tracestorepb.SearchResponse) *tracestorepb.SearchResponse {
	response := &tracestorepb.SearchResponse{
		Metrics: &tracestorepb.SearchMetrics{},
	}

	traces := map[string]*tracestorepb.TraceSearchMetadata{}

	for _, sr := range results {
		for _, t := range sr.Traces {
			// Just simply take first result for each trace
			if _, ok := traces[t.TraceID]; !ok {
				traces[t.TraceID] = t
			}
		}
		if sr.Metrics != nil {
			response.Metrics.InspectedBytes += sr.Metrics.InspectedBytes
			response.Metrics.InspectedTraces += sr.Metrics.InspectedTraces
		}
	}

	for _, t := range traces {
		response.Traces = append(response.Traces, t)
	}

	// Sort and limit results
	sort.Slice(response.Traces, func(i, j int) bool {
		return response.Traces[i].StartTimeUnixNano > response.Traces[j].StartTimeUnixNano
	})
	if req.Limit != 0 && int(req.Limit) < len(response.Traces) {
		response.Traces = response.Traces[:req.Limit]
	}

	return response
}

func protoToMetricSeries(proto []*tracestorepb.KeyValue) traceqlmetrics.MetricSeries {
	r := traceqlmetrics.MetricSeries{}
	for i := range proto {
		r[i] = protoToTraceQLStatic(proto[i])
	}
	return r
}

func protoToTraceQLStatic(kv *tracestorepb.KeyValue) traceqlmetrics.KeyValue {
	var val traceql.Static

	switch traceql.StaticType(kv.Value.Type) {
	case traceql.TypeInt:
		val = traceql.NewStaticInt(int(kv.Value.N))
	case traceql.TypeFloat:
		val = traceql.NewStaticFloat(kv.Value.F)
	case traceql.TypeString:
		val = traceql.NewStaticString(kv.Value.S)
	case traceql.TypeBoolean:
		val = traceql.NewStaticBool(kv.Value.B)
	case traceql.TypeDuration:
		val = traceql.NewStaticDuration(time.Duration(kv.Value.D))
	case traceql.TypeStatus:
		val = traceql.NewStaticStatus(traceql.Status(kv.Value.Status))
	case traceql.TypeKind:
		val = traceql.NewStaticKind(traceql.Kind(kv.Value.Kind))
	default:
		val = traceql.NewStaticNil()
	}

	return traceqlmetrics.KeyValue{
		Key:   kv.Key,
		Value: val,
	}
}
