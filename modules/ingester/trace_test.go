package ingester

import (
	"context"
	"testing"

	"example.com/acme/tracestore/pkg/model"
	"example.com/acme/tracestore/pkg/tracestorepb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTraceStartEndTime(t *testing.T) {
	s := model.MustNewSegmentDecoder(model.CurrentEncoding)

	tr := newTrace(nil)

	// initial push
	buff, err := s.PrepareForWrite(&tracestorepb.Trace{}, 10, 20)
	require.NoError(t, err)
	err = tr.Push(context.Background(), "test", buff)
	require.NoError(t, err)

	assert.Equal(t, uint32(10), tr.start)
	assert.Equal(t, uint32(20), tr.end)

	// overwrite start
	buff, err = s.PrepareForWrite(&tracestorepb.Trace{}, 5, 15)
	require.NoError(t, err)
	err = tr.Push(context.Background(), "test", buff)
	require.NoError(t, err)

	assert.Equal(t, uint32(5), tr.start)
	assert.Equal(t, uint32(20), tr.end)

	// overwrite end
	buff, err = s.PrepareForWrite(&tracestorepb.Trace{}, 15, 25)
	require.NoError(t, err)
	err = tr.Push(context.Background(), "test", buff)
	require.NoError(t, err)

	assert.Equal(t, uint32(5), tr.start)
	assert.Equal(t, uint32(25), tr.end)
}
