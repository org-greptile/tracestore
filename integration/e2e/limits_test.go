package e2e

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"example.com/acme/kit/user"
	util2 "example.com/acme/tracestore/integration/util"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"example.com/acme/e2e"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"google.golang.org/genproto/googleapis/rpc/errdetails"

	"example.com/acme/tracestore/integration/util"
	"example.com/acme/tracestore/pkg/httpclient"
	"example.com/acme/tracestore/pkg/model/trace"
	"example.com/acme/tracestore/pkg/tracestorepb"
	tracestoreUtil "example.com/acme/tracestore/pkg/util"
	"example.com/acme/tracestore/pkg/util/test"

	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

const (
	configLimits             = "config-limits.yaml"
	configLimitsQuery        = "config-limits-query.yaml"
	configLimitsPartialError = "config-limits-partial-success.yaml"
	configLimits429          = "config-limits-429.yaml"
)

func TestLimits(t *testing.T) {
	s, err := e2e.NewScenario("tracestore_e2e")
	require.NoError(t, err)
	defer s.Close()

	require.NoError(t, util2.CopyFileToSharedDir(s, configLimits, "config.yaml"))
	tracestore := util2.NewTracestoreAllInOne()
	require.NoError(t, s.StartAndWaitReady(tracestore))

	// Get port for the otlp receiver endpoint
	c, err := util2.NewJaegerGRPCClient(tracestore.Endpoint(14250))
	require.NoError(t, err)
	require.NotNil(t, c)

	// should fail b/c the trace is too large. each batch should be ~70 bytes
	batch := util.MakeThriftBatchWithSpanCount(2)
	require.NoError(t, c.EmitBatch(context.Background(), batch), "max trace size")

	// push a trace
	require.NoError(t, c.EmitBatch(context.Background(), util.MakeThriftBatchWithSpanCount(1)))

	// should fail b/c this will be too many traces
	batch = util.MakeThriftBatch()
	require.NoError(t, c.EmitBatch(context.Background(), batch), "too many traces")

	// should fail b/c due to ingestion rate limit
	batch = util.MakeThriftBatchWithSpanCount(10)
	err = c.EmitBatch(context.Background(), batch)
	require.Error(t, err)

	// this error must have a retryinfo as expected in otel collector code: https://github.com/open-telemetry/opentelemetry-collector/blob/d7b49df5d9e922df6ce56ad4b64ee1c79f9dbdbe/exporter/otlpexporter/otlp.go#L172
	st, ok := status.FromError(err)
	require.True(t, ok)
	foundRetryInfo := false
	for _, detail := range st.Details() {
		if _, ok := detail.(*errdetails.RetryInfo); ok {
			foundRetryInfo = true
			break
		}
	}
	require.True(t, foundRetryInfo)

	// test limit metrics
	err = tracestore.WaitSumMetricsWithOptions(e2e.Equals(2),
		[]string{"tracestore_discarded_spans_total"},
		e2e.WithLabelMatchers(labels.MustNewMatcher(labels.MatchEqual, "reason", "trace_too_large")),
	)
	require.NoError(t, err)
	err = tracestore.WaitSumMetricsWithOptions(e2e.Equals(1),
		[]string{"tracestore_discarded_spans_total"},
		e2e.WithLabelMatchers(labels.MustNewMatcher(labels.MatchEqual, "reason", "live_traces_exceeded")),
	)
	require.NoError(t, err)
	err = tracestore.WaitSumMetricsWithOptions(e2e.Equals(10),
		[]string{"tracestore_discarded_spans_total"},
		e2e.WithLabelMatchers(labels.MustNewMatcher(labels.MatchEqual, "reason", "rate_limited")),
	)
	require.NoError(t, err)
}

func TestOTLPLimits(t *testing.T) {
	s, err := e2e.NewScenario("tracestore_e2e")
	require.NoError(t, err)
	defer s.Close()

	require.NoError(t, util2.CopyFileToSharedDir(s, configLimits, "config.yaml"))
	tracestore := util2.NewTracestoreAllInOne()
	require.NoError(t, s.StartAndWaitReady(tracestore))

	protoSpans := test.MakeProtoSpans(100)

	// gRPC
	grpcClient := otlptracegrpc.NewClient(
		otlptracegrpc.WithEndpoint(tracestore.Endpoint(4317)),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithRetry(otlptracegrpc.RetryConfig{Enabled: false}),
	)
	require.NoError(t, grpcClient.Start(context.Background()))

	grpcErr := grpcClient.UploadTraces(context.Background(), protoSpans)
	assert.Error(t, grpcErr)
	require.Equal(t, codes.ResourceExhausted, status.Code(grpcErr))

	// HTTP
	httpClient := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(tracestore.Endpoint(4318)),
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithRetry(otlptracehttp.RetryConfig{Enabled: false}),
	)
	require.NoError(t, httpClient.Start(context.Background()))

	httpErr := httpClient.UploadTraces(context.Background(), protoSpans)
	assert.Error(t, httpErr)
	require.Contains(t, httpErr.Error(), "retry-able request failure")
}

func TestOTLPLimitsVanillaClient(t *testing.T) {
	s, err := e2e.NewScenario("tracestore_e2e")
	require.NoError(t, err)
	defer s.Close()

	require.NoError(t, util2.CopyFileToSharedDir(s, configLimits, "config.yaml"))
	tracestore := util2.NewTracestoreAllInOne()
	require.NoError(t, s.StartAndWaitReady(tracestore))

	trace := test.MakeTrace(10, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	testCases := []struct {
		name    string
		payload func() []byte
		headers map[string]string
	}{
		// TODO There is an issue when sending the payload in json format. The server returns a 200 instead of a 429.
		// {
		// 	"JSON format",
		// 	func() []byte {
		// 		b := &bytes.Buffer{}
		// 		err := (&jsonpb.Marshaler{}).Marshal(b, trace)
		// 		require.NoError(t, err)
		// 		return b.Bytes()
		// 	},
		// 	map[string]string{
		// 		"Content-Type": "application/json",
		// 	},
		// },
		{
			"Proto format",
			func() []byte {
				b, err := trace.Marshal()
				require.NoError(t, err)
				return b
			},
			map[string]string{
				"Content-Type": "application/x-protobuf",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "http://"+tracestore.Endpoint(4318)+"/v1/traces", bytes.NewReader(tc.payload()))
			require.NoError(t, err)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()
			bodyBytes, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			fmt.Println(string(bodyBytes))

			assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
		})
	}
}

func TestQueryLimits(t *testing.T) {
	s, err := e2e.NewScenario("tracestore_e2e")
	require.NoError(t, err)
	defer s.Close()

	require.NoError(t, util2.CopyFileToSharedDir(s, configLimitsQuery, "config.yaml"))
	tracestore := util2.NewTracestoreAllInOne()
	require.NoError(t, s.StartAndWaitReady(tracestore))

	// Get port for the otlp receiver endpoint
	c, err := util2.NewJaegerGRPCClient(tracestore.Endpoint(14250))
	require.NoError(t, err)
	require.NotNil(t, c)

	// make a trace with 10 spans and push them one at a time, flush in between each one to force different blocks
	batch := util.MakeThriftBatchWithSpanCount(5)
	allSpans := batch.Spans
	for i := range batch.Spans {
		batch.Spans = allSpans[i : i+1]
		require.NoError(t, c.EmitBatch(context.Background(), batch))
		util.CallFlush(t, tracestore)
		// this push along with the double flush is required to forget the too large trace
		require.NoError(t, c.EmitBatch(context.Background(), util.MakeThriftBatchWithSpanCount(1)))
		util.CallFlush(t, tracestore)
		time.Sleep(2 * time.Second) // trace idle and flush time are both 1ms
	}

	// calc trace id
	traceID := [16]byte{}
	binary.BigEndian.PutUint64(traceID[:8], uint64(batch.Spans[0].TraceIdHigh))
	binary.BigEndian.PutUint64(traceID[8:], uint64(batch.Spans[0].TraceIdLow))

	// now try to query it back. this should fail b/c the trace is too large
	client := httpclient.New("http://"+tracestore.Endpoint(3200), tracestoreUtil.FakeTenantID)
	querierClient := httpclient.New("http://"+tracestore.Endpoint(3200)+"/querier", tracestoreUtil.FakeTenantID)

	_, err = client.QueryTrace(tracestoreUtil.TraceIDToHexString(traceID[:]))
	require.ErrorContains(t, err, trace.ErrTraceTooLarge.Error())
	require.ErrorContains(t, err, "failed with response: 422") // confirm frontend returns 422

	_, err = querierClient.QueryTrace(tracestoreUtil.TraceIDToHexString(traceID[:]))
	require.ErrorContains(t, err, trace.ErrTraceTooLarge.Error())
	require.ErrorContains(t, err, "failed with response: 422")

	// complete block timeout  is 10 seconds
	time.Sleep(15 * time.Second)
	_, err = client.QueryTrace(tracestoreUtil.TraceIDToHexString(traceID[:]))
	require.ErrorContains(t, err, trace.ErrTraceTooLarge.Error())
	require.ErrorContains(t, err, "failed with response: 422") // confirm frontend returns 422

	_, err = querierClient.QueryTrace(tracestoreUtil.TraceIDToHexString(traceID[:]))
	require.ErrorContains(t, err, trace.ErrTraceTooLarge.Error())
	require.ErrorContains(t, err, "failed with response: 422") // confirm querier returns 422
}

func TestLimitsPartialSuccess(t *testing.T) {
	s, err := e2e.NewScenario("tracestore_e2e")
	require.NoError(t, err)
	defer s.Close()
	require.NoError(t, util2.CopyFileToSharedDir(s, configLimitsPartialError, "config.yaml"))
	tracestore := util2.NewTracestoreAllInOne()
	require.NoError(t, s.StartAndWaitReady(tracestore))

	// otel grpc exporter
	exporter, err := util2.NewOtelGRPCExporter(tracestore.Endpoint(4317))
	require.NoError(t, err)

	err = exporter.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)

	// make request
	traceIDs := make([][]byte, 6)
	for index := range traceIDs {
		traceID := make([]byte, 16)
		_, err = crand.Read(traceID)
		require.NoError(t, err)
		traceIDs[index] = traceID
	}

	// 3 traces with trace_too_large and 3 with no error
	spanCountsByTrace := []int{1, 4, 1, 5, 6, 1}
	req := test.MakeReqWithMultipleTraceWithSpanCount(spanCountsByTrace, traceIDs)

	b, err := req.Marshal()
	require.NoError(t, err)

	// unmarshal into otlp proto
	traces, err := (&ptrace.ProtoUnmarshaler{}).UnmarshalTraces(b)
	require.NoError(t, err)
	require.NotNil(t, traces)

	ctx := user.InjectOrgID(context.Background(), tracestoreUtil.FakeTenantID)
	ctx, err = user.InjectIntoGRPCRequest(ctx)
	require.NoError(t, err)

	// send traces to tracestore
	// partial success = no error
	err = exporter.ConsumeTraces(ctx, traces)
	require.NoError(t, err)

	// shutdown to ensure traces are flushed
	require.NoError(t, exporter.Shutdown(context.Background()))

	// query for the one trace that didn't trigger an error
	client := httpclient.New("http://"+tracestore.Endpoint(3200), tracestoreUtil.FakeTenantID)
	for i, count := range spanCountsByTrace {
		if count == 1 {
			result, err := client.QueryTrace(tracestoreUtil.TraceIDToHexString(traceIDs[i]))
			require.NoError(t, err)
			assert.Equal(t, 1, len(result.ResourceSpans))
		}
	}

	// test metrics
	// 3 traces with trace_too_large each with 4+5+6 spans
	err = tracestore.WaitSumMetricsWithOptions(e2e.Equals(15),
		[]string{"tracestore_discarded_spans_total"},
		e2e.WithLabelMatchers(labels.MustNewMatcher(labels.MatchEqual, "reason", "trace_too_large")),
	)
	require.NoError(t, err)

	// this metric should never exist
	err = tracestore.WaitSumMetricsWithOptions(e2e.Equals(0),
		[]string{"tracestore_discarded_spans_total"},
		e2e.WithLabelMatchers(labels.MustNewMatcher(labels.MatchEqual, "reason", "unknown_error")),
	)
	require.NoError(t, err)
}

func TestQueryRateLimits(t *testing.T) {
	s, err := e2e.NewScenario("tracestore_e2e")
	require.NoError(t, err)
	defer s.Close()

	require.NoError(t, util2.CopyFileToSharedDir(s, configLimits429, "config.yaml"))
	tracestore := util2.NewTracestoreAllInOne()
	require.NoError(t, s.StartAndWaitReady(tracestore))

	// Get port for the otlp receiver endpoint
	c, err := util2.NewJaegerGRPCClient(tracestore.Endpoint(14250))
	require.NoError(t, err)
	require.NotNil(t, c)

	// make a trace with 10 spans and push them one at a time, flush in between each one to force different blocks
	batch := util.MakeThriftBatchWithSpanCount(5)
	allSpans := batch.Spans
	for i := range batch.Spans {
		batch.Spans = allSpans[i : i+1]
		require.NoError(t, c.EmitBatch(context.Background(), batch))
		util.CallFlush(t, tracestore)
		time.Sleep(2 * time.Second) // trace idle and flush time are both 1ms
	}
	// now try to query it back. this should fail b/c the frontend queue doesn't have room
	client := httpclient.New("http://"+tracestore.Endpoint(3200), tracestoreUtil.FakeTenantID)

	// 429 HTTP Trace ID Lookup
	traceID := []byte{0x01, 0x02}
	_, err = client.QueryTrace(tracestoreUtil.TraceIDToHexString(traceID))
	require.ErrorContains(t, err, "job queue full")
	require.ErrorContains(t, err, "failed with response: 429")

	start := time.Now().Add(-1 * time.Hour).Unix()
	end := time.Now().Add(1 * time.Hour).Unix()

	// 429 HTTP Search
	_, err = client.SearchTraceQLWithRange("{}", start, end)
	require.ErrorContains(t, err, "job queue full")
	require.ErrorContains(t, err, "failed with response: 429")

	// 429 GRPC Search
	grpcClient, err := util2.NewSearchGRPCClient(context.Background(), tracestore.Endpoint(3200))
	require.NoError(t, err)

	resp, err := grpcClient.Search(context.Background(), &tracestorepb.SearchRequest{
		Query: "{}",
		Start: uint32(start),
		End:   uint32(end),
	})
	require.NoError(t, err)

	// loop until we get io.EOF or an error
	for {
		_, err = resp.Recv()
		if err != nil {
			break
		}
	}
	require.ErrorContains(t, err, "job queue full")
	require.ErrorContains(t, err, "code = ResourceExhausted")
}
