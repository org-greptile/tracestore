package api

import (
	"net/http"

	"example.com/acme/tracestore/pkg/tracestorepb"
)

// IsSearchBlock returns true if the request appears to be for backend blocks. It is not exhaustive
// and only looks for blockID
func IsSearchBlock(r *http.Request) bool {
	q := r.URL.Query()

	return q.Get(urlParamBlockID) != ""
}

// IsTraceQLQuery returns true if the request contains a traceQL query.
func IsTraceQLQuery(r *tracestorepb.SearchRequest) bool {
	return len(r.Query) > 0
}
