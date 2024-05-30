package v1

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"example.com/acme/tracestore/pkg/model/decoder"
	"example.com/acme/tracestore/pkg/model/trace"
	"example.com/acme/tracestore/pkg/tracestorepb"
)

type SegmentDecoder struct{}

var segmentDecoder = &SegmentDecoder{}

// NewSegmentDecoder() returns a v1 segment decoder.
func NewSegmentDecoder() *SegmentDecoder {
	return segmentDecoder
}

func (d *SegmentDecoder) PrepareForWrite(trace *tracestorepb.Trace, _, _ uint32) ([]byte, error) {
	// v1 encoding doesn't support start/end
	return proto.Marshal(trace)
}

func (d *SegmentDecoder) PrepareForRead(segments [][]byte) (*tracestorepb.Trace, error) {
	// each slice is a marshalled tracestorepb.Trace, unmarshal and combine
	combiner := trace.NewCombiner(0)
	for i, s := range segments {
		t := &tracestorepb.Trace{}
		err := proto.Unmarshal(s, t)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling trace: %w", err)
		}

		_, err = combiner.ConsumeWithFinal(t, i == len(segments)-1)
		if err != nil {
			return nil, fmt.Errorf("error combining trace: %w", err)
		}
	}

	combinedTrace, _ := combiner.Result()

	return combinedTrace, nil
}

func (d *SegmentDecoder) ToObject(segments [][]byte) ([]byte, error) {
	// wrap byte slices in a tracestorepb.TraceBytes and marshal
	wrapper := &tracestorepb.TraceBytes{
		Traces: append([][]byte(nil), segments...),
	}
	return proto.Marshal(wrapper)
}

func (d *SegmentDecoder) FastRange([]byte) (uint32, uint32, error) {
	return 0, 0, decoder.ErrUnsupported
}
