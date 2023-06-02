package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"example.com/acme/tracestore/pkg/tracestorepb"
	"example.com/acme/tracestore/pkg/traceql"
	"github.com/pkg/errors"
)

const (
	muxVarTagName = "tagName"

	ParamScopeIntrinsic = "intrinsic"
)

// ParseSearchTagValuesRequest handles parsing of requests from /api/search/tags/{tagName}/values and /api/v2/search/tags/{tagName}/values
func ParseSearchTagValuesRequest(r *http.Request) (*tracestorepb.SearchTagValuesRequest, error) {
	return parseSearchTagValuesRequest(r, false)
}

func ParseSearchTagValuesRequestV2(r *http.Request) (*tracestorepb.SearchTagValuesRequest, error) {
	return parseSearchTagValuesRequest(r, true)
}

func parseSearchTagValuesRequest(r *http.Request, enforceTraceQL bool) (*tracestorepb.SearchTagValuesRequest, error) {
	vars := mux.Vars(r)
	tagName, ok := vars[muxVarTagName]
	if !ok {
		return nil, errors.New("please provide a tagName")
	}

	if tagName == "" {
		return nil, errors.New("please provide a non-empty tagName")
	}

	if enforceTraceQL {
		_, err := traceql.ParseIdentifier(tagName)
		if err != nil {
			return nil, errors.Wrap(err, "please provide a valid tagName")
		}
	}

	query, _ := extractQueryParam(r, urlParamQuery)

	req := &tracestorepb.SearchTagValuesRequest{
		TagName: tagName,
		Query:   query,
	}

	return req, nil
}

func ParseSearchTagsRequest(r *http.Request) (*tracestorepb.SearchTagsRequest, error) {
	scope, _ := extractQueryParam(r, urlParamScope)

	attScope := traceql.AttributeScopeFromString(scope)
	if attScope == traceql.AttributeScopeUnknown && scope != ParamScopeIntrinsic {
		return nil, fmt.Errorf("invalid scope: %s", scope)
	}

	return &tracestorepb.SearchTagsRequest{
		Scope: scope,
	}, nil
}
