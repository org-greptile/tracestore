package trace

import (
	"testing"

	"example.com/acme/tracestore/pkg/tracestorepb"
	v1 "example.com/acme/tracestore/pkg/tracestorepb/trace/v1"
	"github.com/stretchr/testify/assert"
)

func TestSortTrace(t *testing.T) {
	tests := []struct {
		input    *tracestorepb.Trace
		expected *tracestorepb.Trace
	}{
		{
			input:    &tracestorepb.Trace{},
			expected: &tracestorepb.Trace{},
		},

		{
			input: &tracestorepb.Trace{
				ResourceSpans: []*v1.ResourceSpans{
					{
						ScopeSpans: []*v1.ScopeSpans{
							{
								Spans: []*v1.Span{
									{
										StartTimeUnixNano: 2,
									},
								},
							},
						},
					},
					{
						ScopeSpans: []*v1.ScopeSpans{
							{
								Spans: []*v1.Span{
									{
										StartTimeUnixNano: 1,
									},
								},
							},
						},
					},
				},
			},
			expected: &tracestorepb.Trace{
				ResourceSpans: []*v1.ResourceSpans{
					{
						ScopeSpans: []*v1.ScopeSpans{
							{
								Spans: []*v1.Span{
									{
										StartTimeUnixNano: 1,
									},
								},
							},
						},
					},
					{
						ScopeSpans: []*v1.ScopeSpans{
							{
								Spans: []*v1.Span{
									{
										StartTimeUnixNano: 2,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		SortTrace(tt.input)

		assert.Equal(t, tt.expected, tt.input)
	}
}
