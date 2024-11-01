package pipeline

import (
	"fmt"
	"net/url"

	"example.com/acme/tracestore/modules/frontend/combiner"
	"example.com/acme/tracestore/pkg/traceql"
)

type queryValidatorWare struct {
	next AsyncRoundTripper[combiner.PipelineResponse]
}

func NewQueryValidatorWare() AsyncMiddleware[combiner.PipelineResponse] {
	return AsyncMiddlewareFunc[combiner.PipelineResponse](func(next AsyncRoundTripper[combiner.PipelineResponse]) AsyncRoundTripper[combiner.PipelineResponse] {
		return &queryValidatorWare{
			next: next,
		}
	})
}

func (c queryValidatorWare) RoundTrip(req Request) (Responses[combiner.PipelineResponse], error) {
	query := req.HTTPRequest().URL.Query()
	err := c.validateTraceQLQuery(query)
	if err != nil {
		return NewBadRequest(err), nil
	}
	return c.next.RoundTrip(req)
}

func (c queryValidatorWare) validateTraceQLQuery(queryParams url.Values) error {
	var traceQLQuery string
	if queryParams.Has("q") {
		traceQLQuery = queryParams.Get("q")
	}
	if queryParams.Has("query") {
		traceQLQuery = queryParams.Get("query")
	}
	if traceQLQuery != "" {
		expr, err := traceql.Parse(traceQLQuery)
		if err == nil {
			err = traceql.Validate(expr)
		}
		if err != nil {
			return fmt.Errorf("invalid TraceQL query: %w", err)
		}
	}
	return nil
}
