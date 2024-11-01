package querier

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/log/level"
	"example.com/acme/kit/ring"
	"example.com/acme/kit/user"
	"example.com/acme/tracestore/pkg/tracestorepb"
	v1 "example.com/acme/tracestore/pkg/tracestorepb/common/v1"
	"example.com/acme/tracestore/pkg/traceql"
	"example.com/acme/tracestore/pkg/util/log"
	"example.com/acme/tracestore/tracestoredb/backend"
	"example.com/acme/tracestore/tracestoredb/encoding/common"
)

func (q *Querier) QueryRange(ctx context.Context, req *tracestorepb.QueryRangeRequest) (*tracestorepb.QueryRangeResponse, error) {
	if req.QueryMode == QueryModeRecent {
		return q.queryRangeRecent(ctx, req)
	}

	return q.queryBlock(ctx, req)
}

func (q *Querier) queryRangeRecent(ctx context.Context, req *tracestorepb.QueryRangeRequest) (*tracestorepb.QueryRangeResponse, error) {
	// // Get results from all generators
	replicationSet, err := q.generatorRing.GetReplicationSetForOperation(ring.Read)
	if err != nil {
		return nil, fmt.Errorf("error finding generators in Querier.queryRangeRecent: %w", err)
	}

	c, err := traceql.QueryRangeCombinerFor(req, traceql.AggregateModeSum, false)
	if err != nil {
		return nil, err
	}

	mtx := sync.Mutex{} // combiner doesn't lock, so take lock before calling Combine to make is safe
	forEach := func(ctx context.Context, client tracestorepb.MetricsGeneratorClient) error {
		resp, err := client.QueryRange(ctx, req)
		if err != nil {
			return err
		}
		mtx.Lock()
		defer mtx.Unlock()
		c.Combine(resp)
		return nil
	}
	err = q.forGivenGenerators(ctx, replicationSet, forEach)
	if err != nil {
		_ = level.Error(log.Logger).Log("error querying generators in Querier.queryRangeRecent", "err", err)
		return nil, fmt.Errorf("error querying generators in Querier.queryRangeRecent: %w", err)
	}

	return c.Response(), nil
}

func (q *Querier) queryBlock(ctx context.Context, req *tracestorepb.QueryRangeRequest) (*tracestorepb.QueryRangeResponse, error) {
	tenantID, err := user.ExtractOrgID(ctx)
	if err != nil {
		return nil, fmt.Errorf("error extracting org id in Querier.queryBlock: %w", err)
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
		Version:   req.Version,
		TenantID:  tenantID,
		StartTime: time.Unix(0, int64(req.Start)),
		EndTime:   time.Unix(0, int64(req.End)),
		Encoding:  enc,
		// IndexPageSize:    req.IndexPageSize,
		// TotalRecords:     req.TotalRecords,
		BlockID: blockID,
		// DataEncoding:     req.DataEncoding,
		Size_:            req.Size_,
		FooterSize:       req.FooterSize,
		DedicatedColumns: dc,
	}

	opts := common.DefaultSearchOptions()
	opts.StartPage = int(req.StartPage)
	opts.TotalPages = int(req.PagesToSearch)

	unsafe := q.limits.UnsafeQueryHints(tenantID)

	expr, err := traceql.Parse(req.Query)
	if err != nil {
		return nil, err
	}

	timeOverlapCutoff := q.cfg.Metrics.TimeOverlapCutoff
	if v, ok := expr.Hints.GetFloat(traceql.HintTimeOverlapCutoff, unsafe); ok && v >= 0 && v <= 1.0 {
		timeOverlapCutoff = v
	}

	eval, err := traceql.NewEngine().CompileMetricsQueryRange(req, int(req.Exemplars), timeOverlapCutoff, unsafe)
	if err != nil {
		return nil, err
	}

	f := traceql.NewSpansetFetcherWrapper(func(ctx context.Context, req traceql.FetchSpansRequest) (traceql.FetchSpansResponse, error) {
		return q.store.Fetch(ctx, meta, req, opts)
	})
	err = eval.Do(ctx, f, uint64(meta.StartTime.UnixNano()), uint64(meta.EndTime.UnixNano()))
	if err != nil {
		return nil, err
	}

	res := eval.Results()

	inspectedBytes, spansTotal, _ := eval.Metrics()

	return &tracestorepb.QueryRangeResponse{
		Series: queryRangeTraceQLToProto(res, req),
		Metrics: &tracestorepb.SearchMetrics{
			InspectedBytes: inspectedBytes,
			InspectedSpans: spansTotal,
		},
	}, nil
}

func queryRangeTraceQLToProto(set traceql.SeriesSet, req *tracestorepb.QueryRangeRequest) []*tracestorepb.TimeSeries {
	resp := make([]*tracestorepb.TimeSeries, 0, len(set))

	for promLabels, s := range set {
		labels := make([]v1.KeyValue, 0, len(s.Labels))
		for _, label := range s.Labels {
			labels = append(labels,
				v1.KeyValue{
					Key:   label.Name,
					Value: label.Value.AsAnyValue(),
				},
			)
		}

		intervals := traceql.IntervalCount(req.Start, req.End, req.Step)
		samples := make([]tracestorepb.Sample, 0, intervals)
		for i, value := range s.Values {

			ts := traceql.TimestampOf(uint64(i), req.Start, req.Step)

			samples = append(samples, tracestorepb.Sample{
				TimestampMs: time.Unix(0, int64(ts)).UnixMilli(),
				Value:       value,
			})
		}

		exemplars := make([]tracestorepb.Exemplar, 0, len(s.Exemplars))
		for _, e := range s.Exemplars {
			lbls := make([]v1.KeyValue, 0, len(e.Labels))
			for _, label := range e.Labels {
				lbls = append(lbls,
					v1.KeyValue{
						Key:   label.Name,
						Value: label.Value.AsAnyValue(),
					},
				)
			}
			exemplars = append(exemplars, tracestorepb.Exemplar{
				Labels:      lbls,
				TimestampMs: int64(e.TimestampMs),
				Value:       e.Value,
			})
		}

		ss := &tracestorepb.TimeSeries{
			PromLabels: promLabels,
			Labels:     labels,
			Samples:    samples,
			Exemplars:  exemplars,
		}

		resp = append(resp, ss)
	}

	return resp
}
