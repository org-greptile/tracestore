package deployments

import (
	"sync"
	"testing"
	"time"

	"example.com/acme/e2e"
	e2edb "example.com/acme/e2e/db"
	"example.com/acme/tracestore/integration/util"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"

	"example.com/acme/tracestore/pkg/httpclient"
	tracestoreUtil "example.com/acme/tracestore/pkg/util"
)

const (
	configHA = "config-scalable-single-binary.yaml"
)

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

	require.NoError(t, tracestore1.WaitSumMetricsWithOptions(e2e.Equals(3), []string{`tracestore_ring_members`}, e2e.WithLabelMatchers(matchers...), e2e.WaitMissingMetrics))
	require.NoError(t, tracestore2.WaitSumMetricsWithOptions(e2e.Equals(3), []string{`tracestore_ring_members`}, e2e.WithLabelMatchers(matchers...), e2e.WaitMissingMetrics))
	require.NoError(t, tracestore3.WaitSumMetricsWithOptions(e2e.Equals(3), []string{`tracestore_ring_members`}, e2e.WithLabelMatchers(matchers...), e2e.WaitMissingMetrics))

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
	require.NoError(t, tracestore1.WaitSumMetrics(e2e.Equals(util.SpanCount(expected)), "tracestore_distributor_spans_received_total"))

	// wait trace_idle_time and ensure trace is created in ingester
	time.Sleep(1 * time.Second)
	require.NoError(t, tracestore1.WaitSumMetricsWithOptions(e2e.Less(3), []string{"tracestore_ingester_traces_created_total"}, e2e.WaitMissingMetrics))

	for _, i := range []*e2e.HTTPService{tracestore1, tracestore2, tracestore3} {
		util.CallFlush(t, i)
		require.NoError(t, i.WaitSumMetrics(e2e.Equals(1), "tracestore_ingester_blocks_flushed_total"))
		util.CallIngesterRing(t, i)
		util.CallCompactorRing(t, i)
		util.CallStatus(t, i)
		util.CallBuildinfo(t, i)
	}

	apiClient1 := httpclient.New("http://"+tracestore1.Endpoint(3200), "")

	util.QueryAndAssertTrace(t, apiClient1, info)

	err = tracestore1.Kill()
	require.NoError(t, err)

	// Push to one of the instances that are still running.
	require.NoError(t, info.EmitBatches(c2))

	err = tracestore2.Kill()
	require.NoError(t, err)

	err = tracestore3.Kill()
	require.NoError(t, err)
}
