package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	v2 "example.com/acme/tracestore/tracestoredb/encoding/v2"

	"example.com/acme/e2e"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"example.com/acme/tracestore/cmd/tracestore/app"
	"example.com/acme/tracestore/integration"
	"example.com/acme/tracestore/integration/e2e/backend"
	"example.com/acme/tracestore/pkg/httpclient"
	"example.com/acme/tracestore/pkg/util"
	"example.com/acme/tracestore/tracestoredb/encoding"
)

const (
	configAllEncodings = "./config-encodings.tmpl.yaml"
)

func TestEncodings(t *testing.T) {
	const repeatedSearchCount = 10

	for _, enc := range encoding.AllEncodings() {
		t.Run(enc.Version(), func(t *testing.T) {
			s, err := e2e.NewScenario("tracestore_e2e")
			require.NoError(t, err)
			defer s.Close()

			// copy config template to shared directory and expand template variables
			tmplConfig := map[string]any{"Version": enc.Version()}
			config, err := integration.CopyTemplateToSharedDir(s, configAllEncodings, "config.yaml", tmplConfig)
			require.NoError(t, err)

			// load final config
			var cfg app.Config
			buff, err := os.ReadFile(config)
			require.NoError(t, err)
			err = yaml.UnmarshalStrict(buff, &cfg)
			require.NoError(t, err)

			// set up the backend
			_, err = backend.New(s, cfg)
			require.NoError(t, err)

			tracestore := integration.NewTracestoreAllInOne()
			require.NoError(t, s.StartAndWaitReady(tracestore))

			// Get port for the Jaeger gRPC receiver endpoint
			c, err := integration.NewJaegerGRPCClient(tracestore.Endpoint(14250))
			require.NoError(t, err)
			require.NotNil(t, c)

			info := util.NewTraceInfo(time.Now(), "")
			require.NoError(t, info.EmitAllBatches(c))

			expected, err := info.ConstructTraceFromEpoch()
			require.NoError(t, err)

			// test metrics
			require.NoError(t, tracestore.WaitSumMetrics(e2e.Equals(spanCount(expected)), "tracestore_distributor_spans_received_total"))

			// test echo
			assertEcho(t, "http://"+tracestore.Endpoint(3200)+"/api/echo")

			apiClient := httpclient.New("http://"+tracestore.Endpoint(3200), "")

			// query an in-memory trace
			queryAndAssertTrace(t, apiClient, info)

			// wait trace_idle_time and ensure trace is created in ingester
			require.NoError(t, tracestore.WaitSumMetricsWithOptions(e2e.Less(3), []string{"tracestore_ingester_traces_created_total"}, e2e.WaitMissingMetrics))

			// flush trace to backend
			callFlush(t, tracestore)

			// v2 does not support querying and must be skipped
			if enc.Version() != v2.VersionString {
				// search for trace in backend multiple times with different attributes to make sure
				// we search with different scopes and with attributes from dedicated columns
				for i := 0; i < repeatedSearchCount; i++ {
					integration.SearchAndAssertTrace(t, apiClient, info)
					integration.SearchTraceQLAndAssertTrace(t, apiClient, info)
				}
			}

			// sleep
			time.Sleep(10 * time.Second)

			// force clear completed block
			callFlush(t, tracestore)

			// test metrics
			require.NoError(t, tracestore.WaitSumMetrics(e2e.Equals(1), "tracestore_ingester_blocks_flushed_total"))
			require.NoError(t, tracestore.WaitSumMetricsWithOptions(e2e.Equals(1), []string{"tracestoredb_blocklist_length"}, e2e.WaitMissingMetrics))
			if enc.Version() != v2.VersionString {
				require.NoError(t, tracestore.WaitSumMetrics(e2e.Greater(15), "tracestore_query_frontend_queries_total"))
			}

			// query trace - should fetch from backend
			queryAndAssertTrace(t, apiClient, info)

			// create grpc client used for streaming
			grpcClient, err := integration.NewSearchGRPCClient(context.Background(), tracestore.Endpoint(3200))
			require.NoError(t, err)

			if enc.Version() == v2.VersionString {
				return // v2 does not support querying and must be skipped
			}

			// search for trace in backend multiple times with different attributes to make sure
			// we search with different scopes and with attributes from dedicated columns
			now := time.Now()
			for i := 0; i < repeatedSearchCount; i++ {
				// search the backend. this works b/c we're passing a start/end AND setting query ingesters within min/max to 0
				integration.SearchAndAssertTraceBackend(t, apiClient, info, now.Add(-20*time.Minute).Unix(), now.Unix())
				// find the trace with streaming. using the http server b/c that's what Acme will do
				integration.SearchStreamAndAssertTrace(t, context.Background(), grpcClient, info, now.Add(-20*time.Minute).Unix(), now.Unix())
			}
		})
	}
}
