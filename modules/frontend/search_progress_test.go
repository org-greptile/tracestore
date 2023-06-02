package frontend

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"example.com/acme/tracestore/pkg/tracestorepb"
	"example.com/acme/tracestore/pkg/traceql"
	"github.com/stretchr/testify/assert"
)

func TestSearchProgressShouldQuit(t *testing.T) {
	ctx := context.Background()

	// brand-new response should not quit
	sr := newSearchProgress(ctx, 10, 0, 0, 0)
	assert.False(t, sr.shouldQuit())

	// errored response should quit
	sr = newSearchProgress(ctx, 10, 0, 0, 0)
	sr.setError(errors.New("blerg"))
	assert.True(t, sr.shouldQuit())

	// happy status code should not quit
	sr = newSearchProgress(ctx, 10, 0, 0, 0)
	sr.setStatus(200, "")
	assert.False(t, sr.shouldQuit())

	// sad status code should quit
	sr = newSearchProgress(ctx, 10, 0, 0, 0)
	sr.setStatus(400, "")
	assert.True(t, sr.shouldQuit())

	sr = newSearchProgress(ctx, 10, 0, 0, 0)
	sr.setStatus(500, "")
	assert.True(t, sr.shouldQuit())

	// cancelled context should quit
	cancellableContext, cancel := context.WithCancel(ctx)
	sr = newSearchProgress(cancellableContext, 10, 0, 0, 0)
	cancel()
	assert.True(t, sr.shouldQuit())

	// limit reached should quit
	sr = newSearchProgress(ctx, 2, 0, 0, 0)
	sr.addResponse(&tracestorepb.SearchResponse{
		Traces: []*tracestorepb.TraceSearchMetadata{
			{
				TraceID: "something",
			},
		},
		Metrics: &tracestorepb.SearchMetrics{},
	})
	assert.False(t, sr.shouldQuit())
	sr.addResponse(&tracestorepb.SearchResponse{
		Traces: []*tracestorepb.TraceSearchMetadata{
			{
				TraceID: "something",
			},
			{
				TraceID: "something",
			},
		},
		Metrics: &tracestorepb.SearchMetrics{},
	})
	assert.False(t, sr.shouldQuit())
	sr.addResponse(&tracestorepb.SearchResponse{
		Traces: []*tracestorepb.TraceSearchMetadata{
			{
				TraceID: "otherthing",
			},
			{
				TraceID: "thingthatsdifferent",
			},
		},
		Metrics: &tracestorepb.SearchMetrics{},
	})
	assert.True(t, sr.shouldQuit())
}

func TestSearchProgressCombineResults(t *testing.T) {
	start := time.Date(1, 2, 3, 4, 5, 6, 7, time.UTC)
	traceID := "traceID"

	sr := newSearchProgress(context.Background(), 10, 0, 0, 0)
	sr.addResponse(&tracestorepb.SearchResponse{
		Traces: []*tracestorepb.TraceSearchMetadata{
			{
				TraceID:           traceID,
				StartTimeUnixNano: uint64(start.Add(time.Second).UnixNano()),
				DurationMs:        uint32(time.Second.Milliseconds()),
			}, // 1 second after start and shorter duration
			{
				TraceID:           traceID,
				StartTimeUnixNano: uint64(start.UnixNano()),
				DurationMs:        uint32(time.Hour.Milliseconds()),
			}, // earliest start time and longer duration
			{
				TraceID:           traceID,
				StartTimeUnixNano: uint64(start.Add(time.Hour).UnixNano()),
				DurationMs:        uint32(time.Millisecond.Milliseconds()),
			}, // 1 hour after start and shorter duration
		},
		Metrics: &tracestorepb.SearchMetrics{},
	})

	expected := &shardedSearchResults{
		response: &tracestorepb.SearchResponse{
			Traces: []*tracestorepb.TraceSearchMetadata{
				{TraceID: traceID, StartTimeUnixNano: uint64(start.UnixNano()), DurationMs: uint32(time.Hour.Milliseconds())},
			},
			Metrics: &tracestorepb.SearchMetrics{
				CompletedJobs: 1,
			},
		},
		finishedRequests: 1,
		statusCode:       200,
	}

	assert.Equal(t, expected, sr.result())

}

func TestInstanceDoesNotRace(t *testing.T) {

	end := make(chan struct{})
	concurrent := func(f func()) {
		for {
			select {
			case <-end:
				return
			default:
				f()
			}
		}
	}

	traceID := "1234"

	progress := newSearchProgress(context.Background(), 10, 0, 0, 0)
	i := 0
	go concurrent(func() {
		i++
		resp := &tracestorepb.SearchResponse{
			Traces: []*tracestorepb.TraceSearchMetadata{
				{
					TraceID:           traceID,
					StartTimeUnixNano: math.MaxUint64 - uint64(i),
					DurationMs:        uint32(i),
					SpanSets: []*tracestorepb.SpanSet{{
						Matched: uint32(i),
					}},
				},
			},
			Metrics: &tracestorepb.SearchMetrics{
				InspectedTraces: 1,
				InspectedBytes:  1,
				TotalBlocks:     1,
				TotalJobs:       1,
				CompletedJobs:   1,
			},
		}
		progress.addResponse(resp)
	})

	combiner := traceql.NewMetadataCombiner()
	go concurrent(func() {
		res := progress.result()
		if len(res.response.Traces) > 0 {
			// by using a combiner we are testing and changing the entire response
			combiner.AddMetadata(res.response.Traces[0])
		}
	})

	time.Sleep(100 * time.Millisecond)
	close(end)
	// Wait for go funcs to quit before
	// exiting and cleaning up
	time.Sleep(2 * time.Second)
}
