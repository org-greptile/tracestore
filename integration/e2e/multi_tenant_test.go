package e2e

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"example.com/acme/kit/user"
	"example.com/acme/e2e"
	"example.com/acme/tracestore/pkg/httpclient"
	"example.com/acme/tracestore/pkg/tracestorepb"
	tracestoreUtil "example.com/acme/tracestore/pkg/util"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"example.com/acme/tracestore/cmd/tracestore/app"
	util "example.com/acme/tracestore/integration"
	"example.com/acme/tracestore/integration/e2e/backend"
)

const (
	configMultiTenant = "config-multi-tenant-local.yaml"
)

type traceStringsMap struct {
	rKeys     []string
	rValues   []string
	spanNames []string
}

// TestMultiTenantSearch tests multi tenant query support
func TestMultiTenantSearch(t *testing.T) {
	testTenants := []struct {
		name       string
		tenant     string
		tenantSize int
	}{
		{
			name:       "single tenant",
			tenant:     "test",
			tenantSize: 1,
		},
		{
			name:       "wildcard tenant",
			tenant:     "*", // tenant id "*" is same as a tenant with name '*', no special handling...
			tenantSize: 1,
		},
		{
			name:       "two tenants",
			tenant:     "test|test2",
			tenantSize: 2,
		},
		{
			name:       "multiple tenants",
			tenant:     "test|test2|test3",
			tenantSize: 3,
		},
	}

	for _, tc := range testTenants {
		t.Run(tc.name, func(t *testing.T) {
			s, err := e2e.NewScenario("tracestore_e2e")
			require.NoError(t, err)
			defer s.Close()

			// set up the backend
			cfg := app.Config{}
			buff, err := os.ReadFile(configMultiTenant)
			require.NoError(t, err)
			err = yaml.UnmarshalStrict(buff, &cfg)
			require.NoError(t, err)
			_, err = backend.New(s, cfg)
			require.NoError(t, err)

			require.NoError(t, util.CopyFileToSharedDir(s, configMultiTenant, "config.yaml"))
			tracestore := util.NewTracestoreAllInOne()
			require.NoError(t, s.StartAndWaitReady(tracestore, newPrometheus()))

			// Get port for the Jaeger gRPC receiver endpoint
			c, err := util.NewJaegerGRPCClient(tracestore.Endpoint(14250))
			require.NoError(t, err)
			require.NotNil(t, c)

			var info *tracestoreUtil.TraceInfo
			var traceMap traceStringsMap

			tenants := strings.Split(tc.tenant, "|")
			require.Equal(t, tc.tenantSize, len(tenants))

			var expectedSpans float64
			// write traces for all tenants
			for _, tenant := range tenants {
				info = tracestoreUtil.NewTraceInfo(time.Now(), tenant)
				require.NoError(t, info.EmitAllBatches(c))

				trace, err := info.ConstructTraceFromEpoch()
				traceMap = getAttrsAndSpanNames(trace) // store it to assert tests

				require.NoError(t, err)
				expectedSpans = expectedSpans + spanCount(trace)
			}

			// assert that we have one trace and each tenant and correct number of spans received
			require.NoError(t, tracestore.WaitSumMetrics(e2e.Equals(float64(tc.tenantSize)), "tracestore_ingester_traces_created_total"))
			require.NoError(t, tracestore.WaitSumMetrics(e2e.Equals(expectedSpans), "tracestore_distributor_spans_received_total"))

			// Wait for the traces to be written to the WAL
			time.Sleep(time.Second * 3)

			// test echo
			assertEcho(t, "http://"+tracestore.Endpoint(3200)+"/api/echo")

			// client will have testcase tenant id
			apiClient := httpclient.New("http://"+tracestore.Endpoint(3200), tc.tenant)

			// check trace by id
			resp, err := apiClient.QueryTrace(info.HexID())
			require.NoError(t, err)
			respTm := getAttrsAndSpanNames(resp)

			assert.ElementsMatch(t, traceMap.rValues, respTm.rValues)
			assert.ElementsMatch(t, respTm.rKeys, traceMap.rKeys)
			assert.ElementsMatch(t, traceMap.spanNames, respTm.spanNames)

			// flush trace to backend
			callFlush(t, tracestore)

			// search and traceql search, note: SearchAndAssertTrace also calls SearchTagValues
			util.SearchAndAssertTrace(t, apiClient, info)
			util.SearchTraceQLAndAssertTrace(t, apiClient, info)

			// force clear completed block
			callFlush(t, tracestore)

			// wait for flush to complete for all tenants, each tenant will have one block
			require.NoError(t, tracestore.WaitSumMetrics(e2e.Equals(float64(tc.tenantSize)), "tracestore_ingester_blocks_flushed_total"))

			// call search tags endpoints, ensure no errors and results are not empty
			tagsResp, err := apiClient.SearchTags()
			require.NoError(t, err)
			require.NotEmpty(t, tagsResp.TagNames)

			tagsV2Resp, err := apiClient.SearchTagsV2()
			require.NoError(t, err)
			require.Equal(t, 3, len(tagsV2Resp.GetScopes()))
			for _, s := range tagsV2Resp.Scopes {
				require.NotEmpty(t, s.Tags)
			}

			tagsValuesResp, err := apiClient.SearchTagValues("vulture-0")
			require.NoError(t, err)
			require.NotEmpty(t, tagsValuesResp.TagValues)

			tagsValuesV2Resp, err := apiClient.SearchTagValuesV2("span.vulture-0", "{}")
			require.NoError(t, err)
			require.NotEmpty(t, tagsValuesV2Resp.TagValues)

			// check metrics for all routes
			routeTable := []struct {
				route    string
				reqCount int
			}{
				// query frontend routes
				{route: "api_search", reqCount: 2}, // called twice
				{route: "api_traces_traceid", reqCount: 1},
				{route: "api_search_tags", reqCount: 1},
				{route: "api_search_tag_tagname_values", reqCount: 2}, // called twice
				{route: "api_v2_search_tags", reqCount: 1},
				{route: "api_v2_search_tag_tagname_values", reqCount: 1},
				// Querier routes, we make one request for each tenant
				{route: "/tracestorepb.Querier/SearchRecent", reqCount: 2 * tc.tenantSize}, // called twice
				{route: "/tracestorepb.Querier/FindTraceByID", reqCount: tc.tenantSize},
				{route: "/tracestorepb.Querier/SearchTags", reqCount: tc.tenantSize},
				{route: "/tracestorepb.Querier/SearchTagsV2", reqCount: tc.tenantSize},
				{route: "/tracestorepb.Querier/SearchTagValues", reqCount: 2 * tc.tenantSize}, // called twice
				{route: "/tracestorepb.Querier/SearchTagValuesV2", reqCount: tc.tenantSize},
			}
			for _, rt := range routeTable {
				assertRequestCountMetric(t, tracestore, rt.route, rt.reqCount)
			}

			// test streaming search over grpc
			grpcCtx := user.InjectOrgID(context.Background(), tc.tenant)
			grpcCtx, err = user.InjectIntoGRPCRequest(grpcCtx)
			require.NoError(t, err)

			grpcClient, err := util.NewSearchGRPCClient(grpcCtx, tracestore.Endpoint(3200))
			require.NoError(t, err)

			time.Sleep(2 * time.Second) // ensure that blocklist poller has built the blocklist
			now := time.Now()
			util.SearchStreamAndAssertTrace(t, grpcCtx, grpcClient, info, now.Add(-10*time.Minute).Unix(), now.Add(10*time.Minute).Unix())
			assertRequestCountMetric(t, tracestore, "/tracestorepb.StreamingQuerier/Search", 1)

			// test unsupported endpoint
			_, msErr := apiClient.MetricsSummary("{}", "name", 0, 0)
			if tc.tenantSize > 1 {
				// error for multi-tenant request for unsupported endpoints
				require.Error(t, msErr)
			} else {
				require.NoError(t, msErr)
			}
		})
	}
}

func assertRequestCountMetric(t *testing.T, s *e2e.HTTPService, route string, reqCount int) {
	fmt.Printf("==== %s, assertRequestCountMetric route: %v, rt.reqCount: %v \n", t.Name(), route, reqCount)

	err := s.WaitSumMetricsWithOptions(e2e.Equals(float64(reqCount)),
		[]string{"tracestore_request_duration_seconds"},
		e2e.WithLabelMatchers(labels.MustNewMatcher(labels.MatchEqual, "route", route)),
		e2e.WithMetricCount, // get count from histogram metric
	)
	require.NoError(t, err)
}

// getAttrsAndSpanNames returns trace attrs and span names
func getAttrsAndSpanNames(trace *tracestorepb.Trace) traceStringsMap {
	rAttrsKeys := tracestoreUtil.NewDistinctStringCollector(0)
	rAttrsValues := tracestoreUtil.NewDistinctStringCollector(0)
	spanNames := tracestoreUtil.NewDistinctStringCollector(0)

	for _, b := range trace.Batches {
		for _, ss := range b.ScopeSpans {
			for _, s := range ss.Spans {
				if s.Name != "" {
					spanNames.Collect(s.Name)
				}
			}
		}
		for _, a := range b.Resource.Attributes {
			if a.Key != "" {
				rAttrsKeys.Collect(a.Key)
			}
			if a.Value.GetStringValue() != "" {
				rAttrsValues.Collect(a.Value.GetStringValue())
			}
		}
	}

	return traceStringsMap{
		rKeys:     rAttrsKeys.Strings(),
		rValues:   rAttrsValues.Strings(),
		spanNames: spanNames.Strings(),
	}
}
