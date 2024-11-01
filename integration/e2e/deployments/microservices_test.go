package deployments

import (
	"fmt"
	"testing"
	"time"

	"example.com/acme/e2e"
	e2edb "example.com/acme/e2e/db"
	"example.com/acme/tracestore/integration/util"
	"example.com/acme/tracestore/pkg/httpclient"
	tracestoreUtil "example.com/acme/tracestore/pkg/util"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"
)

const configMicroservices = "config-microservices.tmpl.yaml"

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

			KVStoreConfig := tc.kvconfig("", 0)
			if kvstore != nil {
				KVStoreConfig = tc.kvconfig(kvstore.Name(), kvstore.HTTPPort())
			}

			// copy config template to shared directory and expand template variables
			tmplConfig := map[string]any{"KVStore": KVStoreConfig}
			_, err = util.CopyTemplateToSharedDir(s, configMicroservices, "config.yaml", tmplConfig)
			require.NoError(t, err)

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
			require.NoError(t, tracestoreDistributor.WaitSumMetricsWithOptions(e2e.Equals(3), []string{`tracestore_ring_members`}, e2e.WithLabelMatchers(matchers...), e2e.WaitMissingMetrics))

			// Get port for the Jaeger gRPC receiver endpoint
			c, err := util.NewJaegerGRPCClient(tracestoreDistributor.Endpoint(14250))
			require.NoError(t, err)
			require.NotNil(t, c)

			info := tracestoreUtil.NewTraceInfo(time.Now(), "")
			require.NoError(t, info.EmitAllBatches(c))

			expected, err := info.ConstructTraceFromEpoch()
			require.NoError(t, err)

			// test metrics
			require.NoError(t, tracestoreDistributor.WaitSumMetrics(e2e.Equals(util.SpanCount(expected)), "tracestore_distributor_spans_received_total"))

			// test echo
			util.AssertEcho(t, "http://"+tracestoreQueryFrontend.Endpoint(3200)+"/api/echo")

			apiClient := httpclient.New("http://"+tracestoreQueryFrontend.Endpoint(3200), "")

			// query an in-memory trace
			util.QueryAndAssertTrace(t, apiClient, info)

			// wait trace_idle_time and ensure trace is created in ingester
			require.NoError(t, tracestoreIngester1.WaitSumMetricsWithOptions(e2e.Less(3), []string{"tracestore_ingester_traces_created_total"}, e2e.WaitMissingMetrics))
			require.NoError(t, tracestoreIngester2.WaitSumMetricsWithOptions(e2e.Less(3), []string{"tracestore_ingester_traces_created_total"}, e2e.WaitMissingMetrics))
			require.NoError(t, tracestoreIngester3.WaitSumMetricsWithOptions(e2e.Less(3), []string{"tracestore_ingester_traces_created_total"}, e2e.WaitMissingMetrics))

			// flush trace to backend
			util.CallFlush(t, tracestoreIngester1)
			util.CallFlush(t, tracestoreIngester2)
			util.CallFlush(t, tracestoreIngester3)

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
			util.QueryAndAssertTrace(t, apiClient, info)

			// stop an ingester and confirm we can still write and query
			err = tracestoreIngester2.Kill()
			require.NoError(t, err)

			// sleep for heartbeat timeout
			time.Sleep(1 * time.Second)

			info = tracestoreUtil.NewTraceInfo(time.Now(), "")
			require.NoError(t, info.EmitAllBatches(c))

			// query by id
			util.QueryAndAssertTrace(t, apiClient, info)

			// wait trace_idle_time and ensure trace is created in ingester
			require.NoError(t, tracestoreIngester1.WaitSumMetricsWithOptions(e2e.Less(4), []string{"tracestore_ingester_traces_created_total"}, e2e.WaitMissingMetrics))
			require.NoError(t, tracestoreIngester3.WaitSumMetricsWithOptions(e2e.Less(4), []string{"tracestore_ingester_traces_created_total"}, e2e.WaitMissingMetrics))

			// flush trace to backend
			util.CallFlush(t, tracestoreIngester1)
			util.CallFlush(t, tracestoreIngester3)

			// search for trace
			util.SearchAndAssertTrace(t, apiClient, info)

			// stop another ingester and confirm things fail
			err = tracestoreIngester1.Kill()
			require.NoError(t, err)

			require.Error(t, info.EmitBatches(c))
		})
	}
}
