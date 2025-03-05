package deployments

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"example.com/acme/e2e"
	"example.com/acme/tracestore/cmd/tracestore/app"
	"example.com/acme/tracestore/integration/e2e/backend"
	"example.com/acme/tracestore/integration/util"
	"example.com/acme/tracestore/pkg/httpclient"
	tracestoreUtil "example.com/acme/tracestore/pkg/util"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

const (
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

	storageBackendTestPermutations := []struct {
		name   string
		prefix string
	}{
		{
			name: "no-prefix",
		},
		{
			name:   "prefix",
			prefix: "a/b/c/",
		},
	}

	for _, tc := range testBackends {
		for _, pc := range storageBackendTestPermutations {
			t.Run(tc.name+"-"+pc.name, func(t *testing.T) {
				s, err := e2e.NewScenario("tracestore_e2e")
				require.NoError(t, err)
				defer s.Close()

				// copy config template to shared directory and expand template variables
				tmplConfig := map[string]any{"Prefix": pc.prefix}
				configFile, err := util.CopyTemplateToSharedDir(s, tc.configFile, "config.yaml", tmplConfig)
				require.NoError(t, err)

				// set up the backend
				cfg := app.Config{}
				buff, err := os.ReadFile(configFile)
				require.NoError(t, err)
				err = yaml.UnmarshalStrict(buff, &cfg)
				require.NoError(t, err)
				_, err = backend.New(s, cfg)
				require.NoError(t, err)

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
				require.NoError(t, tracestore.WaitSumMetrics(e2e.Equals(util.SpanCount(expected)), "tracestore_distributor_spans_received_total"))

				// test echo
				// nolint:goconst
				util.AssertEcho(t, "http://"+tracestore.Endpoint(3200)+"/api/echo")

				apiClient := httpclient.New("http://"+tracestore.Endpoint(3200), "")

				// query an in-memory trace
				util.QueryAndAssertTrace(t, apiClient, info)

				// wait trace_idle_time and ensure trace is created in ingester
				require.NoError(t, tracestore.WaitSumMetricsWithOptions(e2e.Less(3), []string{"tracestore_ingester_traces_created_total"}, e2e.WaitMissingMetrics))

				// flush trace to backend
				util.CallFlush(t, tracestore)

				// search for trace in backend
				util.SearchAndAssertTrace(t, apiClient, info)
				util.SearchTraceQLAndAssertTrace(t, apiClient, info)

				// sleep
				time.Sleep(10 * time.Second)

				// force clear completed block
				util.CallFlush(t, tracestore)

				fmt.Println(tracestore.Endpoint(3200))
				// test metrics
				require.NoError(t, tracestore.WaitSumMetrics(e2e.Equals(1), "tracestore_ingester_blocks_flushed_total"))
				require.NoError(t, tracestore.WaitSumMetricsWithOptions(e2e.Equals(1), []string{"tracestoredb_blocklist_length"}, e2e.WaitMissingMetrics))
				require.NoError(t, tracestore.WaitSumMetrics(e2e.Equals(4), "tracestore_query_frontend_queries_total"))

				matchers := []*labels.Matcher{
					{
						Type:  labels.MatchEqual,
						Name:  "receiver",
						Value: "tracestore/jaeger_receiver",
					},
					{
						Type:  labels.MatchEqual,
						Name:  "transport",
						Value: "grpc",
					},
				}

				require.NoError(t, tracestore.WaitSumMetricsWithOptions(e2e.Greater(1), []string{"tracestore_receiver_accepted_spans"}, e2e.WithLabelMatchers(matchers...)))
				require.NoError(t, tracestore.WaitSumMetricsWithOptions(e2e.Equals(0), []string{"tracestore_receiver_refused_spans"}, e2e.WithLabelMatchers(matchers...)))

				// query trace - should fetch from backend
				util.QueryAndAssertTrace(t, apiClient, info)

				// search the backend. this works b/c we're passing a start/end AND setting query ingesters within min/max to 0
				now := time.Now()
				util.SearchAndAssertTraceBackend(t, apiClient, info, now.Add(-20*time.Minute).Unix(), now.Unix())

				util.SearchAndAsserTagsBackend(t, apiClient, now.Add(-20*time.Minute).Unix(), now.Unix())

				// find the trace with streaming. using the http server b/c that's what Acme will do
				grpcClient, err := util.NewSearchGRPCClient(context.Background(), tracestore.Endpoint(3200))
				require.NoError(t, err)

				util.SearchStreamAndAssertTrace(t, context.Background(), grpcClient, info, now.Add(-20*time.Minute).Unix(), now.Unix())
			})
		}
	}
}

func TestShutdownDelay(t *testing.T) {
	s, err := e2e.NewScenario("tracestore_e2e")
	require.NoError(t, err)
	defer s.Close()

	// copy config template to shared directory and expand template variables
	tmplConfig := map[string]any{"Prefix": ""}
	configFile, err := util.CopyTemplateToSharedDir(s, configAllInOneS3, "config.yaml", tmplConfig)
	require.NoError(t, err)

	// set up the backend
	cfg := app.Config{}
	buff, err := os.ReadFile(configFile)
	require.NoError(t, err)
	err = yaml.UnmarshalStrict(buff, &cfg)
	require.NoError(t, err)
	_, err = backend.New(s, cfg)
	require.NoError(t, err)

	tracestore := util.NewTracestoreAllInOne("-shutdown-delay=5s")

	// this line tests confirms that the readiness flag is up
	require.NoError(t, s.StartAndWaitReady(tracestore))

	// if we're here the readiness flag is up. now call kill and check the readiness flag is down
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		for i := 0; i < 10; i++ {
			res, err := e2e.DoGet("http://" + tracestore.Endpoint(3200) + "/ready")
			require.NoError(t, err)
			res.Body.Close()

			if res.StatusCode == http.StatusServiceUnavailable {
				// found it!
				return
			}
			time.Sleep(time.Second)
		}

		require.Fail(t, "readiness flag never went down")
	}()

	// call stop and allow the code above to test for a unavailable readiness flag
	_ = tracestore.Stop()

	wg.Wait()
}
