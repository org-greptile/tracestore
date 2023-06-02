package generator

import (
	"context"
	"net/http"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"example.com/acme/tracestore/pkg/api"
	"example.com/acme/tracestore/pkg/tracestorepb"
	"github.com/opentracing/opentracing-go"
)

func (g *Generator) SpanMetricsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithDeadline(r.Context(), time.Now().Add(g.cfg.QueryTimeout))
	defer cancel()

	span, ctx := opentracing.StartSpanFromContext(ctx, "Generator.SpanMetricsHandler")
	defer span.Finish()

	span.SetTag("requestURI", r.RequestURI)

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
