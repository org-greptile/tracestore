package e2e

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"text/template"
	"time"

	thrift "github.com/jaegertracing/jaeger/thrift-gen/jaeger"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"example.com/acme/e2e"
	e2edb "example.com/acme/e2e/db"

	"example.com/acme/tracestore/cmd/tracestore/app"
	util "example.com/acme/tracestore/integration"
	"example.com/acme/tracestore/integration/e2e/backend"
	"example.com/acme/tracestore/pkg/model/trace"
	"example.com/acme/tracestore/pkg/tracestorepb"
	tracestoreUtil "example.com/acme/tracestore/pkg/util"
)

const (
	configMicroservices = "config-microservices.tmpl.yaml"
	configHA            = "config-scalable-single-binary.yaml"

	configAllInOneS3      = "config-all-in-one-s3.yaml"
	configAllInOneAzurite = "config-all-in-one-azurite.yaml"
	configAllInOneGCS     = "config-all-in-one-gcs.yaml"
)

func TestAllInOne(t *testing.T) {
	testBackends := []struct {
		name       string
		configFile string
	}{
		{
			name:       "s3",
			configFile: configAllInOneS3,
		},
		{
			name:       "azure",
			configFile: configAllInOneAzurite,
		},
		{
			name:       "gcs",
			configFile: configAllInOneGCS,
		},
	}

	for _, tc := range testBackends {
		t.Run(tc.name, func(t *testing.T) {
			s, err := e2e.NewScenario("tracestore_e2e")
			require.NoError(t, err)
			defer s.Close()

			// set up the backend
			cfg := app.Config{}
			buff, err := os.ReadFile(tc.configFile)
			require.NoError(t, err)
			err = yaml.UnmarshalStrict(buff, &cfg)
			require.NoError(t, err)
			_, err = backend.New(s, cfg)
			require.NoError(t, err)

			require.NoError(t, util.CopyFileToSharedDir(s, tc.configFile, "config.yaml"))
			tracestore := util.NewTracestoreAllInOne()
			require.NoError(t, s.StartAndWaitReady(tracestore))

			// Get port for the Jaeger gRPC receiver endpoint
			c, err := util.NewJaegerGRPCClient(tracestore.Endpoint(14250))
			require.NoError(t, err)
			require.NotNil(t, c)

			info := tracestoreUtil.NewTraceInfo(time.Now(), "")
			require.NoError(t, info.EmitAllBatches(c))

			expected, err := info.ConstructTraceFromEpoch()
			require.NoError(t, err)

			// test metrics
			require.NoError(t, tracestore.WaitSumMetrics(e2e.Equals(spanCount(expected)), "tracestore_distributor_spans_received_total"))

			// test echo
			assertEcho(t, "http://"+tracestore.Endpoint(3200)+"/api/echo")

			apiClient := tracestoreUtil.NewClient("http://"+tracestore.Endpoint(3200), "")

			// query an in-memory trace
			queryAndAssertTrace(t, apiClient, info)

			// wait trace_idle_time and ensure trace is created in ingester
			require.NoError(t, tracestore.WaitSumMetricsWithOptions(e2e.Less(3), []string{"tracestore_ingester_traces_created_total"}, e2e.WaitMissingMetrics))

			// flush trace to backend
			callFlush(t, tracestore)

			// search for trace in backend
			util.SearchAndAssertTrace(t, apiClient, info)
			util.SearchTraceQLAndAssertTrace(t, apiClient, info)

			// sleep
			time.Sleep(10 * time.Second)

			// force clear completed block
			callFlush(t, tracestore)

			// test metrics
			require.NoError(t, tracestore.WaitSumMetrics(e2e.Equals(1), "tracestore_ingester_blocks_flushed_total"))
			require.NoError(t, tracestore.WaitSumMetricsWithOptions(e2e.Equals(1), []string{"tracestoredb_blocklist_length"}, e2e.WaitMissingMetrics))
			require.NoError(t, tracestore.WaitSumMetrics(e2e.Equals(4), "tracestore_query_frontend_queries_total"))

			// query trace - should fetch from backend
			queryAndAssertTrace(t, apiClient, info)

			// search the backend. this works b/c we're passing a start/end AND setting query ingesters within min/max to 0
			now := time.Now()
			util.SearchAndAssertTraceBackend(t, apiClient, info, now.Add(-20*time.Minute).Unix(), now.Unix())
		})
	}
}

func TestMicroservicesWithKVStores(t *testing.T) {
	testKVStores := []struct {
		name     string
		kvconfig func(hostname string, port int) string
	}{
		{
			name: "memberlist",
			kvconfig: func(string, int) string {
				return `
        store: memberlist`
			},
		},
		{
			name: "etcd",
			kvconfig: func(hostname string, port int) string {
				return fmt.Sprintf(`
        store: etcd
        etcd:
          endpoints:
            - http://%s:%d`, hostname, port)
			},
		},
		{
			name: "consul",
			kvconfig: func(hostname string, port int) string {
				return fmt.Sprintf(`
        store: consul
        consul:
          host: http://%s:%d`, hostname, port)
			},
		},
	}

	for _, tc := range testKVStores {
		t.Run(tc.name, func(t *testing.T) {
			s, err := e2e.NewScenario("tracestore_e2e")
			require.NoError(t, err)
			defer s.Close()

			// Set up KVStore
			var kvstore *e2e.HTTPService
			switch tc.name {
			case "etcd":
				kvstore = e2edb.NewETCD()
				require.NoError(t, s.StartAndWaitReady(kvstore))
			case "consul":
				kvstore = e2edb.NewConsul()
				require.NoError(t, s.StartAndWaitReady(kvstore))
			case "memberlist":
			default:
				t.Errorf("unknown KVStore %s", tc.name)
			}

			tmpl, err := template.New(filepath.Base(configMicroservices)).ParseFiles(configMicroservices)
			require.NoError(t, err)

			KVStoreConfig := tc.kvconfig("", 0)
			if kvstore != nil {
				KVStoreConfig = tc.kvconfig(kvstore.Name(), kvstore.HTTPPort())
			}

			var buf bytes.Buffer
			kvconfig := map[string]interface{}{
				"KVStore": KVStoreConfig,
			}
			require.NoError(t, tmpl.Execute(&buf, kvconfig))

			require.NoError(t, util.WriteFileToSharedDir(s, "config.yaml", buf.Bytes()))

			minio := e2edb.NewMinio(9000, "tracestore")
			require.NotNil(t, minio)
			require.NoError(t, s.StartAndWaitReady(minio))

			tracestoreIngester1 := util.NewTracestoreIngester(1)
			tracestoreIngester2 := util.NewTracestoreIngester(2)
			tracestoreIngester3 := util.NewTracestoreIngester(3)

			tracestoreDistributor := util.NewTracestoreDistributor()
			tracestoreQueryFrontend := util.NewTracestoreQueryFrontend()
			tracestoreQuerier := util.NewTracestoreQuerier()
			require.NoError(t, s.StartAndWaitReady(tracestoreIngester1, tracestoreIngester2, tracestoreIngester3, tracestoreDistributor, tracestoreQueryFrontend, tracestoreQuerier))

			// wait for active ingesters
			time.Sleep(1 * time.Second)
			matchers := []*labels.Matcher{
				{
					Type:  labels.MatchEqual,
					Name:  "name",
					Value: "ingester",
				},
				{
					Type:  labels.MatchEqual,
					Name:  "state",
					Value: "ACTIVE",
				},
			}
			require.NoError(t, tracestoreDistributor.WaitSumMetricsWithOptions(e2e.Equals(3), []string{`corestore_ring_members`}, e2e.WithLabelMatchers(matchers...), e2e.WaitMissingMetrics))

			// Get port for the Jaeger gRPC receiver endpoint
			c, err := util.NewJaegerGRPCClient(tracestoreDistributor.Endpoint(14250))
			require.NoError(t, err)
			require.NotNil(t, c)

			info := tracestoreUtil.NewTraceInfo(time.Now(), "")
			require.NoError(t, info.EmitAllBatches(c))

			expected, err := info.ConstructTraceFromEpoch()
			require.NoError(t, err)

			// test metrics
			require.NoError(t, tracestoreDistributor.WaitSumMetrics(e2e.Equals(spanCount(expected)), "tracestore_distributor_spans_received_total"))

			// test echo
			assertEcho(t, "http://"+tracestoreQueryFrontend.Endpoint(3200)+"/api/echo")

			apiClient := tracestoreUtil.NewClient("http://"+tracestoreQueryFrontend.Endpoint(3200), "")

			// query an in-memory trace
			queryAndAssertTrace(t, apiClient, info)

			// wait trace_idle_time and ensure trace is created in ingester
			require.NoError(t, tracestoreIngester1.WaitSumMetricsWithOptions(e2e.Less(3), []string{"tracestore_ingester_traces_created_total"}, e2e.WaitMissingMetrics))
			require.NoError(t, tracestoreIngester2.WaitSumMetricsWithOptions(e2e.Less(3), []string{"tracestore_ingester_traces_created_total"}, e2e.WaitMissingMetrics))
			require.NoError(t, tracestoreIngester3.WaitSumMetricsWithOptions(e2e.Less(3), []string{"tracestore_ingester_traces_created_total"}, e2e.WaitMissingMetrics))

			// flush trace to backend
			callFlush(t, tracestoreIngester1)
			callFlush(t, tracestoreIngester2)
			callFlush(t, tracestoreIngester3)

			// search for trace
			util.SearchAndAssertTrace(t, apiClient, info)
			util.SearchTraceQLAndAssertTrace(t, apiClient, info)

			// sleep for one maintenance cycle
			time.Sleep(5 * time.Second)

			// test metrics
			for _, i := range []*e2e.HTTPService{tracestoreIngester1, tracestoreIngester2, tracestoreIngester3} {
				require.NoError(t, i.WaitSumMetrics(e2e.Equals(1), "tracestore_ingester_blocks_flushed_total"))
			}
			require.NoError(t, tracestoreQuerier.WaitSumMetrics(e2e.Equals(3), "tracestoredb_blocklist_length"))
			require.NoError(t, tracestoreQueryFrontend.WaitSumMetrics(e2e.Equals(4), "tracestore_query_frontend_queries_total"))

			// query trace - should fetch from backend
			queryAndAssertTrace(t, apiClient, info)

			// stop an ingester and confirm we can still write and query
			err = tracestoreIngester2.Kill()
			require.NoError(t, err)

			// sleep for heartbeat timeout
			time.Sleep(1 * time.Second)

			info = tracestoreUtil.NewTraceInfo(time.Now(), "")
			require.NoError(t, info.EmitAllBatches(c))

			// query by id
			queryAndAssertTrace(t, apiClient, info)

			// wait trace_idle_time and ensure trace is created in ingester
			require.NoError(t, tracestoreIngester1.WaitSumMetricsWithOptions(e2e.Less(4), []string{"tracestore_ingester_traces_created_total"}, e2e.WaitMissingMetrics))
			require.NoError(t, tracestoreIngester3.WaitSumMetricsWithOptions(e2e.Less(4), []string{"tracestore_ingester_traces_created_total"}, e2e.WaitMissingMetrics))

			// flush trace to backend
			callFlush(t, tracestoreIngester1)
			callFlush(t, tracestoreIngester3)

			// search for trace
			util.SearchAndAssertTrace(t, apiClient, info)

			// stop another ingester and confirm things fail
			err = tracestoreIngester1.Kill()
			require.NoError(t, err)

			require.Error(t, info.EmitBatches(c))
		})
	}
}

func TestScalableSingleBinary(t *testing.T) {
	s, err := e2e.NewScenario("tracestore_e2e")
	require.NoError(t, err)
	defer s.Close()

	minio := e2edb.NewMinio(9000, "tracestore")
	require.NotNil(t, minio)
	require.NoError(t, s.StartAndWaitReady(minio))

	// copy configuration file over to shared dir
	require.NoError(t, util.CopyFileToSharedDir(s, configHA, "config.yaml"))

	// start three scalable single binary tracestores in parallel
	var wg sync.WaitGroup
	var tracestore1, tracestore2, tracestore3 *e2e.HTTPService
	wg.Add(3)
	go func() {
		tracestore1 = util.NewTracestoreScalableSingleBinary(1)
		wg.Done()
	}()
	go func() {
		tracestore2 = util.NewTracestoreScalableSingleBinary(2)
		wg.Done()
	}()
	go func() {
		tracestore3 = util.NewTracestoreScalableSingleBinary(3)
		wg.Done()
	}()
	wg.Wait()
	require.NoError(t, s.StartAndWaitReady(tracestore1, tracestore2, tracestore3))

	// wait for 2 active ingesters
	time.Sleep(1 * time.Second)
	matchers := []*labels.Matcher{
		{
			Type:  labels.MatchEqual,
			Name:  "name",
			Value: "ingester",
		},
		{
			Type:  labels.MatchEqual,
			Name:  "state",
			Value: "ACTIVE",
		},
	}

	t.Logf("tracestore1.Endpoint(): %+v", tracestore1.Endpoint(3200))

	require.NoError(t, tracestore1.WaitSumMetricsWithOptions(e2e.Equals(3), []string{`corestore_ring_members`}, e2e.WithLabelMatchers(matchers...), e2e.WaitMissingMetrics))
	require.NoError(t, tracestore2.WaitSumMetricsWithOptions(e2e.Equals(3), []string{`corestore_ring_members`}, e2e.WithLabelMatchers(matchers...), e2e.WaitMissingMetrics))
	require.NoError(t, tracestore3.WaitSumMetricsWithOptions(e2e.Equals(3), []string{`corestore_ring_members`}, e2e.WithLabelMatchers(matchers...), e2e.WaitMissingMetrics))

	c1, err := util.NewJaegerGRPCClient(tracestore1.Endpoint(14250))
	require.NoError(t, err)
	require.NotNil(t, c1)

	c2, err := util.NewJaegerGRPCClient(tracestore2.Endpoint(14250))
	require.NoError(t, err)
	require.NotNil(t, c2)

	c3, err := util.NewJaegerGRPCClient(tracestore3.Endpoint(14250))
	require.NoError(t, err)
	require.NotNil(t, c3)

	info := tracestoreUtil.NewTraceInfo(time.Unix(1632169410, 0), "")
	require.NoError(t, info.EmitBatches(c1))

	expected, err := info.ConstructTraceFromEpoch()
	require.NoError(t, err)

	// test metrics
	require.NoError(t, tracestore1.WaitSumMetrics(e2e.Equals(spanCount(expected)), "tracestore_distributor_spans_received_total"))

	// wait trace_idle_time and ensure trace is created in ingester
	time.Sleep(1 * time.Second)
	require.NoError(t, tracestore1.WaitSumMetricsWithOptions(e2e.Less(3), []string{"tracestore_ingester_traces_created_total"}, e2e.WaitMissingMetrics))

	for _, i := range []*e2e.HTTPService{tracestore1, tracestore2, tracestore3} {
		callFlush(t, i)
		require.NoError(t, i.WaitSumMetrics(e2e.Equals(1), "tracestore_ingester_blocks_flushed_total"))
		callIngesterRing(t, i)
		callCompactorRing(t, i)
		callStatus(t, i)
	}

	apiClient1 := tracestoreUtil.NewClient("http://"+tracestore1.Endpoint(3200), "")

	queryAndAssertTrace(t, apiClient1, info)

	err = tracestore1.Kill()
	require.NoError(t, err)

	// Push to one of the instances that are still running.
	require.NoError(t, info.EmitBatches(c2))

	err = tracestore2.Kill()
	require.NoError(t, err)

	err = tracestore3.Kill()
	require.NoError(t, err)
}

func makeThriftBatch() *thrift.Batch {
	return makeThriftBatchWithSpanCount(1)
}

func makeThriftBatchWithSpanCount(n int) *thrift.Batch {
	var spans []*thrift.Span

	traceIDLow := rand.Int63()
	traceIDHigh := rand.Int63()
	tagValue := "y"
	for i := 0; i < n; i++ {
		spans = append(spans, &thrift.Span{
			TraceIdLow:    traceIDLow,
			TraceIdHigh:   traceIDHigh,
			SpanId:        rand.Int63(),
			ParentSpanId:  0,
			OperationName: "my operation",
			References:    nil,
			Flags:         0,
			StartTime:     time.Now().Unix(),
			Duration:      1,
			Tags: []*thrift.Tag{
				{
					Key:  "x",
					VStr: &tagValue,
				},
			},
			Logs: nil,
		})
	}

	return &thrift.Batch{Spans: spans}
}

func callFlush(t *testing.T, ingester *e2e.HTTPService) {
	fmt.Printf("Calling /flush on %s\n", ingester.Name())
	res, err := e2e.DoGet("http://" + ingester.Endpoint(3200) + "/flush")
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, res.StatusCode)
}

func callIngesterRing(t *testing.T, svc *e2e.HTTPService) {
	endpoint := "/ingester/ring"
	fmt.Printf("Calling %s on %s\n", endpoint, svc.Name())
	res, err := e2e.DoGet("http://" + svc.Endpoint(3200) + endpoint)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
}

func callCompactorRing(t *testing.T, svc *e2e.HTTPService) {
	endpoint := "/compactor/ring"
	fmt.Printf("Calling %s on %s\n", endpoint, svc.Name())
	res, err := e2e.DoGet("http://" + svc.Endpoint(3200) + endpoint)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
}

func callStatus(t *testing.T, svc *e2e.HTTPService) {
	endpoint := "/status/endpoints"
	fmt.Printf("Calling %s on %s\n", endpoint, svc.Name())
	res, err := e2e.DoGet("http://" + svc.Endpoint(3200) + endpoint)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
}

func assertEcho(t *testing.T, url string) {
	res, err := e2e.DoGet(url)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
	defer res.Body.Close()
}

func queryAndAssertTrace(t *testing.T, client *tracestoreUtil.Client, info *tracestoreUtil.TraceInfo) {
	resp, err := client.QueryTrace(info.HexID())
	require.NoError(t, err)

	expected, err := info.ConstructTraceFromEpoch()
	require.NoError(t, err)

	require.True(t, equalTraces(resp, expected))
}

func equalTraces(a, b *tracestorepb.Trace) bool {
	trace.SortTrace(a)
	trace.SortTrace(b)

	return reflect.DeepEqual(a, b)
}

func spanCount(a *tracestorepb.Trace) float64 {
	count := 0
	for _, batch := range a.Batches {
		for _, spans := range batch.ScopeSpans {
			count += len(spans.Spans)
		}
	}

	return float64(count)
}
