package ingest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"example.com/acme/e2e"
	"example.com/acme/tracestore/integration/util"
	"example.com/acme/tracestore/pkg/httpclient"
	tracestoreUtil "example.com/acme/tracestore/pkg/util"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"
)

func TestIngest(t *testing.T) {
	s, err := e2e.NewScenario("tracestore_e2e")
	require.NoError(t, err)
	defer s.Close()

	// copy config template to shared directory and expand template variables
	require.NoError(t, util.CopyFileToSharedDir(s, "config-kafka.yaml", "config.yaml"))

	kafka := NewKafka()
	require.NoError(t, s.StartAndWaitReady(kafka))

	tracestore := util.NewTracestoreAllInOne()
	require.NoError(t, s.StartAndWaitReady(tracestore))

	// Get port for the Jaeger gRPC receiver endpoint
	c, err := util.NewJaegerGRPCClient(tracestore.Endpoint(14250))
	require.NoError(t, err)
	require.NotNil(t, c)

	info := tracestoreUtil.NewTraceInfo(time.Now(), "")
	require.NoError(t, info.EmitAllBatches(c))

	time.Sleep(5 * time.Minute)

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
	require.NoError(t, tracestore.WaitSumMetrics(e2e.Equals(3), "tracestore_query_frontend_queries_total"))

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
}
