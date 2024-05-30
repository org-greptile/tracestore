package traceql

import (
	"testing"

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
