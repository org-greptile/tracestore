package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	util2 "example.com/acme/tracestore/integration/util"
	v2 "example.com/acme/tracestore/tracestoredb/encoding/v2"

	"example.com/acme/e2e"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"example.com/acme/tracestore/cmd/tracestore/app"
	"example.com/acme/tracestore/integration/e2e/backend"
	"example.com/acme/tracestore/integration/util"
	"example.com/acme/tracestore/pkg/httpclient"
	tracestoreUtil "example.com/acme/tracestore/pkg/util"
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
			config, err := util2.CopyTemplateToSharedDir(s, configAllEncodings, "config.yaml", tmplConfig)
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

			tracestore := util2.NewTracestoreAllInOne()
			require.NoError(t, s.StartAndWaitReady(tracestore))

			// Get port for the Jaeger gRPC receiver endpoint
			c, err := util2.NewJaegerGRPCClient(tracestore.Endpoint(14250))
			require.NoError(t, err)
			require.NotNil(t, c)

			info := tracestoreUtil.NewTraceInfo(time.Now(), "")
			require.NoError(t, info.EmitAllBatches(c))

			expected, err := info.ConstructTraceFromEpoch()
			require.NoError(t, err)

			// test metrics
			require.NoError(t, tracestore.WaitSumMetrics(e2e.Equals(util.SpanCount(expected)), "tracestore_distributor_spans_received_total"))

			// test echo
			util.AssertEcho(t, "http://"+tracestore.Endpoint(3200)+"/api/echo")

			apiClient := httpclient.New("http://"+tracestore.Endpoint(3200), "")

			// query an in-memory trace
			util.QueryAndAssertTrace(t, apiClient, info)

			// wait trace_idle_time and ensure trace is created in ingester
			require.NoError(t, tracestore.WaitSumMetricsWithOptions(e2e.Less(3), []string{"tracestore_ingester_traces_created_total"}, e2e.WaitMissingMetrics))

			// flush trace to backend
			util.CallFlush(t, tracestore)

			// v2 does not support querying and must be skipped
			if enc.Version() != v2.VersionString {
				// search for trace in backend multiple times with different attributes to make sure
				// we search with different scopes and with attributes from dedicated columns
				for i := 0; i < repeatedSearchCount; i++ {
					util2.SearchAndAssertTrace(t, apiClient, info)
					util2.SearchTraceQLAndAssertTrace(t, apiClient, info)
				}
			}

			// sleep
			time.Sleep(10 * time.Second)

			// force clear completed block
			util.CallFlush(t, tracestore)

			// test metrics
			require.NoError(t, tracestore.WaitSumMetrics(e2e.Equals(1), "tracestore_ingester_blocks_flushed_total"))
			require.NoError(t, tracestore.WaitSumMetricsWithOptions(e2e.Equals(1), []string{"tracestoredb_blocklist_length"}, e2e.WaitMissingMetrics))
			if enc.Version() != v2.VersionString {
				require.NoError(t, tracestore.WaitSumMetrics(e2e.Greater(15), "tracestore_query_frontend_queries_total"))
			}

			// query trace - should fetch from backend
			util.QueryAndAssertTrace(t, apiClient, info)

			// create grpc client used for streaming
			grpcClient, err := util2.NewSearchGRPCClient(context.Background(), tracestore.Endpoint(3200))
			require.NoError(t, err)

			if enc.Version() == v2.VersionString {
				return // v2 does not support querying and must be skipped
			}

			// search for trace in backend multiple times with different attributes to make sure
			// we search with different scopes and with attributes from dedicated columns
			now := time.Now()
			for i := 0; i < repeatedSearchCount; i++ {
				// search the backend. this works b/c we're passing a start/end AND setting query ingesters within min/max to 0
				util2.SearchAndAssertTraceBackend(t, apiClient, info, now.Add(-20*time.Minute).Unix(), now.Unix())
				// find the trace with streaming. using the http server b/c that's what Acme will do
				util2.SearchStreamAndAssertTrace(t, context.Background(), grpcClient, info, now.Add(-20*time.Minute).Unix(), now.Unix())
			}
		})
	}
}
