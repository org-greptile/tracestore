package generator

import (
	"context"
	"net/http"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"go.opentelemetry.io/otel/attribute"

	"example.com/acme/tracestore/pkg/api"
	"example.com/acme/tracestore/pkg/tracestorepb"
)

func (g *Generator) SpanMetricsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithDeadline(r.Context(), time.Now().Add(g.cfg.QueryTimeout))
	defer cancel()

	ctx, span := tracer.Start(ctx, "Generator.SpanMetricsHandler")
	defer span.End()

	span.SetAttributes(attribute.String("requestURI", r.RequestURI))

	req, err := api.ParseSpanMetricsRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var resp *tracestorepb.SpanMetricsResponse
	resp, err = g.GetMetrics(ctx, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	marshaller := &jsonpb.Marshaler{}
	err = marshaller.Marshal(w, resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set(api.HeaderContentType, api.HeaderAcceptJSON)
}

func (g *Generator) QueryRangeHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithDeadline(r.Context(), time.Now().Add(g.cfg.QueryTimeout))
	defer cancel()

	ctx, span := tracer.Start(ctx, "Generator.QueryRangeHandler")
	defer span.End()

	span.SetAttributes(attribute.String("requestURI", r.RequestURI))

	req, err := api.ParseQueryRangeRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var resp *tracestorepb.QueryRangeResponse
	resp, err = g.QueryRange(ctx, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	marshaller := &jsonpb.Marshaler{}
	err = marshaller.Marshal(w, resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set(api.HeaderContentType, api.HeaderAcceptJSON)
}
