package traceql

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"example.com/acme/tracestore/pkg/tracestorepb"
	v1 "example.com/acme/tracestore/pkg/tracestorepb/common/v1"
	"github.com/stretchr/testify/require"
)

func TestCombineResults(t *testing.T) {
	tcs := []struct {
		name     string
		existing *tracestorepb.TraceSearchMetadata
		new      *tracestorepb.TraceSearchMetadata
		expected *tracestorepb.TraceSearchMetadata
	}{
		{
			name: "overwrite nothing",
			existing: &tracestorepb.TraceSearchMetadata{
				SpanSet:  &tracestorepb.SpanSet{},
				SpanSets: []*tracestorepb.SpanSet{},
			},
			new: &tracestorepb.TraceSearchMetadata{
				TraceID:           "trace-1",
				RootServiceName:   "service-1",
				RootTraceName:     "root-trace-1",
				StartTimeUnixNano: 123,
				DurationMs:        100,
				SpanSets:          []*tracestorepb.SpanSet{},
			},
			expected: &tracestorepb.TraceSearchMetadata{
				TraceID:           "trace-1",
				RootServiceName:   "service-1",
				RootTraceName:     "root-trace-1",
				StartTimeUnixNano: 123,
				DurationMs:        100,
				SpanSets:          []*tracestorepb.SpanSet{},
			},
		},
		{
			name: "mixed copying in fields",
			existing: &tracestorepb.TraceSearchMetadata{
				TraceID:           "existing-trace",
				RootServiceName:   "existing-service",
				RootTraceName:     "existing-root-trace",
				StartTimeUnixNano: 100,
				DurationMs:        200,
				SpanSets:          []*tracestorepb.SpanSet{},
			},
			new: &tracestorepb.TraceSearchMetadata{
				TraceID:           "new-trace",
				RootServiceName:   "new-service",
				RootTraceName:     "new-root-trace",
				StartTimeUnixNano: 150,
				DurationMs:        300,
				SpanSets:          []*tracestorepb.SpanSet{},
			},
			expected: &tracestorepb.TraceSearchMetadata{
				TraceID:           "existing-trace",
				RootServiceName:   "existing-service",
				RootTraceName:     "existing-root-trace",
				StartTimeUnixNano: 100,
				DurationMs:        300,
				SpanSets:          []*tracestorepb.SpanSet{},
			},
		},
		{
			name: "copy in spansets",
			existing: &tracestorepb.TraceSearchMetadata{
				SpanSet:  &tracestorepb.SpanSet{},
				SpanSets: []*tracestorepb.SpanSet{},
			},
			new: &tracestorepb.TraceSearchMetadata{
				SpanSets: []*tracestorepb.SpanSet{
					{
						Matched:    3,
						Spans:      []*tracestorepb.Span{{SpanID: "span-1"}},
						Attributes: []*v1.KeyValue{{Key: "avg(test)", Value: &v1.AnyValue{Value: &v1.AnyValue_DoubleValue{DoubleValue: 1}}}},
					},
				},
			},
			expected: &tracestorepb.TraceSearchMetadata{
				SpanSets: []*tracestorepb.SpanSet{
					{
						Matched:    3,
						Spans:      []*tracestorepb.Span{{SpanID: "span-1"}},
						Attributes: []*v1.KeyValue{{Key: "avg(test)", Value: &v1.AnyValue{Value: &v1.AnyValue_DoubleValue{DoubleValue: 1}}}},
					},
				},
			},
		},
		{
			name: "take higher matches",
			existing: &tracestorepb.TraceSearchMetadata{
				SpanSet: &tracestorepb.SpanSet{},
				SpanSets: []*tracestorepb.SpanSet{
					{
						Matched:    3,
						Spans:      []*tracestorepb.Span{{SpanID: "span-1"}},
						Attributes: []*v1.KeyValue{{Key: "avg(test)", Value: &v1.AnyValue{Value: &v1.AnyValue_DoubleValue{DoubleValue: 1}}}},
					},
				},
			},
			new: &tracestorepb.TraceSearchMetadata{
				SpanSets: []*tracestorepb.SpanSet{
					{
						Matched:    5,
						Spans:      []*tracestorepb.Span{{SpanID: "span-2"}},
						Attributes: []*v1.KeyValue{{Key: "avg(test)", Value: &v1.AnyValue{Value: &v1.AnyValue_DoubleValue{DoubleValue: 3}}}},
					},
				},
			},
			expected: &tracestorepb.TraceSearchMetadata{
				SpanSets: []*tracestorepb.SpanSet{
					{
						Matched:    5,
						Spans:      []*tracestorepb.Span{{SpanID: "span-2"}},
						Attributes: []*v1.KeyValue{{Key: "avg(test)", Value: &v1.AnyValue{Value: &v1.AnyValue_DoubleValue{DoubleValue: 3}}}},
					},
				},
			},
		},
		{
			name: "keep higher matches",
			existing: &tracestorepb.TraceSearchMetadata{
				SpanSet: &tracestorepb.SpanSet{},
				SpanSets: []*tracestorepb.SpanSet{
					{
						Matched:    7,
						Spans:      []*tracestorepb.Span{{SpanID: "span-1"}},
						Attributes: []*v1.KeyValue{{Key: "avg(test)", Value: &v1.AnyValue{Value: &v1.AnyValue_DoubleValue{DoubleValue: 1}}}},
					},
				},
			},
			new: &tracestorepb.TraceSearchMetadata{
				SpanSets: []*tracestorepb.SpanSet{
					{
						Matched:    5,
						Spans:      []*tracestorepb.Span{{SpanID: "span-2"}},
						Attributes: []*v1.KeyValue{{Key: "avg(test)", Value: &v1.AnyValue{Value: &v1.AnyValue_DoubleValue{DoubleValue: 3}}}},
					},
				},
			},
			expected: &tracestorepb.TraceSearchMetadata{
				SpanSets: []*tracestorepb.SpanSet{
					{
						Matched:    7,
						Spans:      []*tracestorepb.Span{{SpanID: "span-1"}},
						Attributes: []*v1.KeyValue{{Key: "avg(test)", Value: &v1.AnyValue{Value: &v1.AnyValue_DoubleValue{DoubleValue: 1}}}},
					},
				},
			},
		},
		{
			name: "respect by()",
			existing: &tracestorepb.TraceSearchMetadata{
				SpanSet: &tracestorepb.SpanSet{},
				SpanSets: []*tracestorepb.SpanSet{
					{
						Matched:    7,
						Spans:      []*tracestorepb.Span{{SpanID: "span-1"}},
						Attributes: []*v1.KeyValue{{Key: "by(name)", Value: &v1.AnyValue{Value: &v1.AnyValue_StringValue{StringValue: "a"}}}},
					},
					{
						Matched:    3,
						Spans:      []*tracestorepb.Span{{SpanID: "span-1"}},
						Attributes: []*v1.KeyValue{{Key: "by(duration)", Value: &v1.AnyValue{Value: &v1.AnyValue_DoubleValue{DoubleValue: 1.1}}}},
					},
				},
			},
			new: &tracestorepb.TraceSearchMetadata{
				SpanSets: []*tracestorepb.SpanSet{
					{
						Matched:    5,
						Spans:      []*tracestorepb.Span{{SpanID: "span-2"}},
						Attributes: []*v1.KeyValue{{Key: "by(name)", Value: &v1.AnyValue{Value: &v1.AnyValue_StringValue{StringValue: "a"}}}},
					},
				},
			},
			expected: &tracestorepb.TraceSearchMetadata{
				SpanSets: []*tracestorepb.SpanSet{
					{
						Matched:    7,
						Spans:      []*tracestorepb.Span{{SpanID: "span-1"}},
						Attributes: []*v1.KeyValue{{Key: "by(name)", Value: &v1.AnyValue{Value: &v1.AnyValue_StringValue{StringValue: "a"}}}},
					},
					{
						Matched:    3,
						Spans:      []*tracestorepb.Span{{SpanID: "span-1"}},
						Attributes: []*v1.KeyValue{{Key: "by(duration)", Value: &v1.AnyValue{Value: &v1.AnyValue_DoubleValue{DoubleValue: 1.1}}}},
					},
				},
			},
		},
		{
			name: "merge ServiceStats",
			existing: &tracestorepb.TraceSearchMetadata{
				ServiceStats: map[string]*tracestorepb.ServiceStats{
					"service1": {
						SpanCount:  5,
						ErrorCount: 1,
					},
				},
			},
			new: &tracestorepb.TraceSearchMetadata{
				ServiceStats: map[string]*tracestorepb.ServiceStats{
					"service1": {
						SpanCount:  3,
						ErrorCount: 2,
					},
				},
			},
			expected: &tracestorepb.TraceSearchMetadata{
				ServiceStats: map[string]*tracestorepb.ServiceStats{
					"service1": {
						SpanCount:  5,
						ErrorCount: 2,
					},
				},
			},
		},
		{
			name:     "existing ServiceStats is nil doesn't panic",
			existing: &tracestorepb.TraceSearchMetadata{},
			new: &tracestorepb.TraceSearchMetadata{
				ServiceStats: map[string]*tracestorepb.ServiceStats{
					"service1": {
						SpanCount:  3,
						ErrorCount: 2,
					},
				},
			},
			expected: &tracestorepb.TraceSearchMetadata{
				ServiceStats: map[string]*tracestorepb.ServiceStats{
					"service1": {
						SpanCount:  3,
						ErrorCount: 2,
					},
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			combineSearchResults(tc.existing, tc.new)

			// confirm that the SpanSet on tc.existing is contained in the slice of SpanSets
			// then nil out. the actual spanset chosen is based on map iteration order
			found := len(tc.existing.SpanSets) == 0
			for _, ss := range tc.existing.SpanSets {
				if ss == tc.existing.SpanSet {
					found = true
					break
				}
			}
			require.True(t, found)
			tc.expected.SpanSet = nil
			tc.existing.SpanSet = nil

			require.Equal(t, tc.expected, tc.existing)
		})
	}
}

// nolint:govet
func TestQueryRangeCombinerDiffs(t *testing.T) {
	start := uint64(100 * time.Millisecond)
	end := uint64(150 * time.Millisecond)
	step := uint64(10 * time.Millisecond)

	tcs := []struct {
		resp, expectedResponse, expectedDiff *tracestorepb.QueryRangeResponse
	}{
		// push nothing get nothing
		{
			resp: &tracestorepb.QueryRangeResponse{},
			expectedResponse: &tracestorepb.QueryRangeResponse{
				Series: []*tracestorepb.TimeSeries{},
			},
		},
		// push 3 data points, get them back
		{
			resp: &tracestorepb.QueryRangeResponse{
				Series: []*tracestorepb.TimeSeries{
					timeSeries("foo", "1", []tracestorepb.Sample{{100, 1}, {110, 2}, {120, 3}}),
				},
			},
			expectedResponse: &tracestorepb.QueryRangeResponse{
				Series: []*tracestorepb.TimeSeries{
					timeSeries("foo", "1", []tracestorepb.Sample{{100, 1}, {110, 2}, {120, 3}, {130, 0}, {140, 0}, {150, 0}}),
				},
			},
			expectedDiff: &tracestorepb.QueryRangeResponse{
				Series: []*tracestorepb.TimeSeries{
					timeSeries("foo", "1", []tracestorepb.Sample{{100, 1}, {110, 2}, {120, 3}}),
				},
			},
		},
		// push 2 data points, check aggregation
		{
			resp: &tracestorepb.QueryRangeResponse{
				Series: []*tracestorepb.TimeSeries{
					timeSeries("foo", "1", []tracestorepb.Sample{{120, 1}, {130, 2}, {150, 3}}),
				},
			},
			expectedResponse: &tracestorepb.QueryRangeResponse{
				Series: []*tracestorepb.TimeSeries{
					timeSeries("foo", "1", []tracestorepb.Sample{{100, 1}, {110, 2}, {120, 4}, {130, 2}, {140, 0}, {150, 3}}),
				},
			},
		},
		// push different series
		{
			resp: &tracestorepb.QueryRangeResponse{
				Series: []*tracestorepb.TimeSeries{
					timeSeries("bar", "1", []tracestorepb.Sample{{100, 1}, {110, 2}, {120, 3}}),
				},
			},
			expectedResponse: &tracestorepb.QueryRangeResponse{
				Series: []*tracestorepb.TimeSeries{
					timeSeries("foo", "1", []tracestorepb.Sample{{100, 1}, {110, 2}, {120, 4}, {130, 2}, {140, 0}, {150, 3}}),
					timeSeries("bar", "1", []tracestorepb.Sample{{100, 1}, {110, 2}, {120, 3}, {130, 0}, {140, 0}, {150, 0}}),
				},
			},
			// includes last 2 pushes
			expectedDiff: &tracestorepb.QueryRangeResponse{
				Series: []*tracestorepb.TimeSeries{
					timeSeries("foo", "1", []tracestorepb.Sample{{120, 4}, {130, 2}, {140, 0}, {150, 3}}),
					timeSeries("bar", "1", []tracestorepb.Sample{{100, 1}, {110, 2}, {120, 3}}),
				},
			},
		},
		// push different series by label value
		{
			resp: &tracestorepb.QueryRangeResponse{
				Series: []*tracestorepb.TimeSeries{
					timeSeries("foo", "2", []tracestorepb.Sample{{100, 1}, {110, 2}, {120, 3}}),
				},
			},
			expectedResponse: &tracestorepb.QueryRangeResponse{
				Series: []*tracestorepb.TimeSeries{
					timeSeries("foo", "1", []tracestorepb.Sample{{100, 1}, {110, 2}, {120, 4}, {130, 2}, {140, 0}, {150, 3}}),
					timeSeries("bar", "1", []tracestorepb.Sample{{100, 1}, {110, 2}, {120, 3}, {130, 0}, {140, 0}, {150, 0}}),
					timeSeries("foo", "2", []tracestorepb.Sample{{100, 1}, {110, 2}, {120, 3}, {130, 0}, {140, 0}, {150, 0}}),
				},
			},
			// includes last 2 pushes
			expectedDiff: &tracestorepb.QueryRangeResponse{
				Series: []*tracestorepb.TimeSeries{
					timeSeries("foo", "2", []tracestorepb.Sample{{100, 1}, {110, 2}, {120, 3}}),
				},
			},
		},
	}

	req := &tracestorepb.QueryRangeRequest{
		Start: start,
		End:   end,
		Step:  step,
		Query: "{} | rate()", // simple aggregate
	}
	combiner, err := QueryRangeCombinerFor(req, AggregateModeFinal, true)
	require.NoError(t, err)

	for i, tc := range tcs {
		t.Run(fmt.Sprintf("step %d", i), func(t *testing.T) {
			combiner.Combine(tc.resp)

			resp := combiner.Response()
			resp.Metrics = nil // we want to ignore metrics for this test, just nil them out
			metricsEqual(t, tc.expectedResponse, resp)

			if tc.expectedDiff != nil {
				// call diff and get expected
				diff := combiner.Diff()
				diff.Metrics = nil
				metricsEqual(t, tc.expectedDiff, diff)

				// call diff again and get nothing!
				diff = combiner.Diff()
				diff.Metrics = nil
				require.Equal(t, &tracestorepb.QueryRangeResponse{
					Series: []*tracestorepb.TimeSeries{},
				}, diff)
			}
		})
	}
}

func metricsEqual(t *testing.T, a, b *tracestorepb.QueryRangeResponse) {
	t.Helper()

	slices.SortFunc(a.Series, func(a, b *tracestorepb.TimeSeries) int {
		return strings.Compare(a.PromLabels, b.PromLabels)
	})
	slices.SortFunc(b.Series, func(a, b *tracestorepb.TimeSeries) int {
		return strings.Compare(a.PromLabels, b.PromLabels)
	})

	require.Equal(t, a, b)
}

func timeSeries(name, val string, samples []tracestorepb.Sample) *tracestorepb.TimeSeries {
	lbls := Labels{
		{
			Name:  name,
			Value: NewStaticString(val),
		},
	}

	return &tracestorepb.TimeSeries{
		Labels:     []v1.KeyValue{{Key: name, Value: &v1.AnyValue{Value: &v1.AnyValue_StringValue{StringValue: val}}}},
		Samples:    samples,
		PromLabels: lbls.String(),
	}
}
