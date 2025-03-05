package tracestoredb

import (
	"context"
	"path"
	"testing"
	"time"

	"github.com/go-kit/log"
	"example.com/acme/tracestore/pkg/model"
	"example.com/acme/tracestore/pkg/tracestorepb"
	common_v1 "example.com/acme/tracestore/pkg/tracestorepb/common/v1"
	resource_v1 "example.com/acme/tracestore/pkg/tracestorepb/resource/v1"
	v1 "example.com/acme/tracestore/pkg/tracestorepb/trace/v1"
	"example.com/acme/tracestore/pkg/traceql"
	"example.com/acme/tracestore/pkg/util/test"
	"example.com/acme/tracestore/tracestoredb/backend"
	"example.com/acme/tracestore/tracestoredb/backend/local"
	"example.com/acme/tracestore/tracestoredb/encoding/common"
	"example.com/acme/tracestore/tracestoredb/encoding/vparquet4"
	"example.com/acme/tracestore/tracestoredb/wal"
	"github.com/stretchr/testify/require"
)

var queryRangeTestCases = []struct {
	req      *tracestorepb.QueryRangeRequest
	expected []*tracestorepb.TimeSeries
}{
	{
		req: &tracestorepb.QueryRangeRequest{
			Start: 1,
			End:   50 * uint64(time.Second),
			Step:  15 * uint64(time.Second),
			Query: "{ } | rate()",
		},
		expected: []*tracestorepb.TimeSeries{
			{
				PromLabels: `{__name__="rate"}`,
				Labels:     []common_v1.KeyValue{tracestorepb.MakeKeyValueString("__name__", "rate")},
				Samples: []tracestorepb.Sample{
					{TimestampMs: 0, Value: 14.0 / 15.0},     // First interval starts at 1, so it only has 14 spans
					{TimestampMs: 15_000, Value: 1.0},        // Spans every 1 second
					{TimestampMs: 30_000, Value: 1.0},        // Spans every 1 second
					{TimestampMs: 45_000, Value: 5.0 / 15.0}, // Interval [45,50) has 5 spans
					{TimestampMs: 60_000, Value: 0},          // I think this is a bug that we extend out an extra interval
				},
			},
		},
	},
	{
		req: &tracestorepb.QueryRangeRequest{
			Start: 1,
			End:   50 * uint64(time.Second),
			Step:  15 * uint64(time.Second),
			Query: `{ .service.name="even" } | rate()`,
		},
		expected: []*tracestorepb.TimeSeries{
			{
				PromLabels: `{__name__="rate"}`,
				Labels:     []common_v1.KeyValue{tracestorepb.MakeKeyValueString("__name__", "rate")},
				Samples: []tracestorepb.Sample{
					{TimestampMs: 0, Value: 7.0 / 15.0},      // Interval [ 1, 14], 7 spans at 2, 4, 6, 8, 10, 12, 14
					{TimestampMs: 15_000, Value: 7.0 / 15.0}, // Interval [15, 29], 7 spans at 16, 18, 20, 22, 24, 26, 28
					{TimestampMs: 30_000, Value: 8.0 / 15.0}, // Interval [30, 44], 8 spans at 30, 32, 34, 36, 38, 40, 42, 44
					{TimestampMs: 45_000, Value: 2.0 / 15.0}, // Interval [45, 50), 2 spans at 46, 48
					{TimestampMs: 60_000, Value: 0},          // I think this is a bug that we extend out an extra interval
				},
			},
		},
	},
}

func TestTracestoreDBQueryRange(t *testing.T) {
	var (
		tempDir      = t.TempDir()
		blockVersion = vparquet4.VersionString
	)

	dc := backend.DedicatedColumns{
		{Scope: "resource", Name: "res-dedicated.01", Type: "string"},
		{Scope: "resource", Name: "res-dedicated.02", Type: "string"},
		{Scope: "span", Name: "span-dedicated.01", Type: "string"},
		{Scope: "span", Name: "span-dedicated.02", Type: "string"},
	}
	r, w, c, err := New(&Config{
		Backend: backend.Local,
		Local: &local.Config{
			Path: path.Join(tempDir, "traces"),
		},
		Block: &common.BlockConfig{
			IndexDownsampleBytes: 17,
			BloomFP:              .01,
			BloomShardSizeBytes:  100_000,
			Version:              blockVersion,
			IndexPageSizeBytes:   1000,
			RowGroupSizeBytes:    10000,
			DedicatedColumns:     dc,
		},
		WAL: &wal.Config{
			Filepath:       path.Join(tempDir, "wal"),
			IngestionSlack: time.Since(time.Time{}),
		},
		Search: &SearchConfig{
			ChunkSizeBytes:  1_000_000,
			ReadBufferCount: 8, ReadBufferSizeBytes: 4 * 1024 * 1024,
		},
		BlocklistPoll: 0,
	}, nil, log.NewNopLogger())
	require.NoError(t, err)

	err = c.EnableCompaction(context.Background(), &CompactorConfig{
		ChunkSizeBytes:          10,
		MaxCompactionRange:      time.Hour,
		BlockRetention:          0,
		CompactedBlockRetention: 0,
	}, &mockSharder{}, &mockOverrides{})
	require.NoError(t, err)

	ctx := context.Background()
	r.EnablePolling(ctx, &mockJobSharder{})

	// Write to wal
	wal := w.WAL()

	meta := &backend.BlockMeta{BlockID: backend.NewUUID(), TenantID: testTenantID, DedicatedColumns: dc}
	head, err := wal.NewBlock(meta, model.CurrentEncoding)
	require.NoError(t, err)
	dec := model.MustNewSegmentDecoder(model.CurrentEncoding)

	totalSpans := 100
	for i := 1; i <= totalSpans; i++ {
		tid := test.ValidTraceID(nil)

		sp := test.MakeSpan(tid)

		// Start time is i seconds
		sp.StartTimeUnixNano = uint64(i * int(time.Second))

		// Duration is i seconds
		sp.EndTimeUnixNano = sp.StartTimeUnixNano + uint64(i*int(time.Second))

		// Service name
		var svcName string
		if i%2 == 0 {
			svcName = "even"
		} else {
			svcName = "odd"
		}

		tr := &tracestorepb.Trace{
			ResourceSpans: []*v1.ResourceSpans{
				{
					Resource: &resource_v1.Resource{
						Attributes: []*common_v1.KeyValue{tracestorepb.MakeKeyValueStringPtr("service.name", svcName)},
					},
					ScopeSpans: []*v1.ScopeSpans{
						{
							Spans: []*v1.Span{
								sp,
							},
						},
					},
				},
			},
		}

		b1, err := dec.PrepareForWrite(tr, 0, 0)
		require.NoError(t, err)

		b2, err := dec.ToObject([][]byte{b1})
		require.NoError(t, err)
		err = head.Append(tid, b2, 0, 0, true)
		require.NoError(t, err)
	}

	// Complete block
	block, err := w.CompleteBlock(context.Background(), head)
	require.NoError(t, err)

	f := traceql.NewSpansetFetcherWrapper(func(ctx context.Context, req traceql.FetchSpansRequest) (traceql.FetchSpansResponse, error) {
		return block.Fetch(ctx, req, common.DefaultSearchOptions())
	})

	for _, tc := range queryRangeTestCases {

		eval, err := traceql.NewEngine().CompileMetricsQueryRange(tc.req, 0, 0, false)
		require.NoError(t, err)

		err = eval.Do(ctx, f, 0, 0)
		require.NoError(t, err)

		actual := eval.Results().ToProto(tc.req)

		require.Equal(t, tc.expected, actual, "Query: %v", tc.req.Query)
	}
}
