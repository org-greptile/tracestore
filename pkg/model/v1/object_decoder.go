package v1

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"example.com/acme/tracestore/pkg/model/decoder"
	"example.com/acme/tracestore/pkg/model/trace"
	"example.com/acme/tracestore/pkg/tracestorepb"
)

const Encoding = "v1"

type ObjectDecoder struct{}

var staticDecoder = &ObjectDecoder{}

func NewObjectDecoder() *ObjectDecoder {
	return staticDecoder
}

func (d *ObjectDecoder) PrepareForRead(obj []byte) (*tracestorepb.Trace, error) {
	trace := &tracestorepb.Trace{}
	traceBytes := &tracestorepb.TraceBytes{}
	err := proto.Unmarshal(obj, traceBytes)
	if err != nil {
		return nil, err
	}

	for _, bytes := range traceBytes.Traces {
		innerTrace := &tracestorepb.Trace{}
		err = proto.Unmarshal(bytes, innerTrace)
		if err != nil {
			return nil, err
		}

		trace.Batches = append(trace.Batches, innerTrace.Batches...)
	}
	return trace, err
}

func (d *ObjectDecoder) Combine(objs ...[]byte) ([]byte, error) {
	c := trace.NewCombiner(0)
	for i, obj := range objs {
		t, err := staticDecoder.PrepareForRead(obj)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling trace: %w", err)
		}

		_, err = c.ConsumeWithFinal(t, i == len(obj)-1)
		if err != nil {
			return nil, fmt.Errorf("error combining trace: %w", err)
		}
	}
	combinedTrace, _ := c.Result()

	combinedBytes, err := d.Marshal(combinedTrace)
	if err != nil {
		return nil, fmt.Errorf("error marshaling combinedBytes: %w", err)
	}

	return combinedBytes, nil
}

func (d *ObjectDecoder) FastRange([]byte) (uint32, uint32, error) {
	return 0, 0, decoder.ErrUnsupported
}

func (d *ObjectDecoder) Marshal(t *tracestorepb.Trace) ([]byte, error) {
	traceBytes := &tracestorepb.TraceBytes{}
	bytes, err := proto.Marshal(t)
	if err != nil {
		return nil, err
	}

	traceBytes.Traces = append(traceBytes.Traces, bytes)

	return proto.Marshal(traceBytes)
}
