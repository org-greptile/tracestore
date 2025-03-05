package main

import (
	"context"
	"net/http"
	"sync"

	userconfigurableoverrides "example.com/acme/tracestore/modules/overrides/userconfigurable/client"
	thrift "github.com/jaegertracing/jaeger/thrift-gen/jaeger"
	"github.com/jaegertracing/jaeger/thrift-gen/zipkincore"

	"example.com/acme/tracestore/pkg/tracestorepb"
)

type MockReporter struct {
	err            error
	batchesEmitted []*thrift.Batch
	// We need the lock to control concurrent accesses to batchesEmitted
	m sync.Mutex
}

func (r *MockReporter) EmitZipkinBatch(_ context.Context, _ []*zipkincore.Span) error {
	return r.err
}

func (r *MockReporter) EmitBatch(_ context.Context, b *thrift.Batch) error {
	if r.err == nil {
		r.m.Lock()
		defer r.m.Unlock()
		r.batchesEmitted = append(r.batchesEmitted, b)
	}

	return r.err
}

func (r *MockReporter) GetEmittedBatches() []*thrift.Batch {
	r.m.Lock()
	defer r.m.Unlock()
	return r.batchesEmitted
}

type MockHTTPClient struct {
	err            error
	resp           http.Response
	traceResp      *tracestorepb.Trace
	requestsCount  int
	searchResponse []*tracestorepb.TraceSearchMetadata
	searchesCount  int
	// We need the lock to control concurrent accesses to shared variables in the tests
	m sync.Mutex
}

//nolint:all
func (m *MockHTTPClient) DeleteOverrides(version string) error {
	panic("unimplemented")
}

//nolint:all
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return &m.resp, m.err
}

//nolint:all
func (m *MockHTTPClient) GetOverrides() (*userconfigurableoverrides.Limits, string, error) {
	panic("unimplemented")
}

//nolint:all
func (m *MockHTTPClient) MetricsSummary(query string, groupBy string, start int64, end int64) (*tracestorepb.SpanMetricsSummaryResponse, error) {
	panic("unimplemented")
}

//nolint:all
func (m *MockHTTPClient) PatchOverrides(limits *userconfigurableoverrides.Limits) (*userconfigurableoverrides.Limits, string, error) {
	panic("unimplemented")
}

//nolint:all
func (m *MockHTTPClient) QueryTrace(id string) (*tracestorepb.Trace, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.m.Lock()
	defer m.m.Unlock()
	m.requestsCount++
	return m.traceResp, m.err
}

//nolint:all
func (m *MockHTTPClient) QueryTraceWithRange(id string, start int64, end int64) (*tracestorepb.Trace, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.m.Lock()
	defer m.m.Unlock()
	m.requestsCount++
	return m.traceResp, m.err
}

func (m *MockHTTPClient) GetRequestsCount() int {
	m.m.Lock()
	defer m.m.Unlock()
	return m.requestsCount
}

//nolint:all
func (m *MockHTTPClient) Search(tags string) (*tracestorepb.SearchResponse, error) {
	panic("unimplemented")
}

//nolint:all
func (m *MockHTTPClient) SearchTagValues(key string) (*tracestorepb.SearchTagValuesResponse, error) {
	panic("unimplemented")
}

//nolint:all
func (m *MockHTTPClient) SearchTagValuesV2(key string, query string) (*tracestorepb.SearchTagValuesV2Response, error) {
	panic("unimplemented")
}

//nolint:all
func (m *MockHTTPClient) SearchTagValuesV2WithRange(tag string, start int64, end int64) (*tracestorepb.SearchTagValuesV2Response, error) {
	panic("unimplemented")
}

//nolint:all
func (m *MockHTTPClient) SearchTags() (*tracestorepb.SearchTagsResponse, error) {
	panic("unimplemented")
}

//nolint:all
func (m *MockHTTPClient) SearchTagsV2() (*tracestorepb.SearchTagsV2Response, error) {
	panic("unimplemented")
}

//nolint:all
func (m *MockHTTPClient) SearchTagsV2WithRange(start int64, end int64) (*tracestorepb.SearchTagsV2Response, error) {
	panic("unimplemented")
}

//nolint:all
func (m *MockHTTPClient) SearchTagsWithRange(start int64, end int64) (*tracestorepb.SearchTagsResponse, error) {
	panic("unimplemented")
}

//nolint:all
func (m *MockHTTPClient) SearchTraceQL(query string) (*tracestorepb.SearchResponse, error) {
	panic("unimplemented")
}

//nolint:all
func (m *MockHTTPClient) SearchTraceQLWithRangeAndLimit(query string, start int64, end int64, limit int64, spss int64) (*tracestorepb.SearchResponse, error) {
	panic("unimplemented")
}

//nolint:all
func (m *MockHTTPClient) SearchTraceQLWithRange(query string, start int64, end int64) (*tracestorepb.SearchResponse, error) {
	if m.err != nil {
		return nil, m.err
	}

	m.m.Lock()
	defer m.m.Unlock()
	traceQlSearchResponse := &tracestorepb.SearchResponse{
		Traces: m.searchResponse,
	}
	m.searchesCount++
	return traceQlSearchResponse, m.err
}

//nolint:all
func (m *MockHTTPClient) SearchWithRange(tags string, start int64, end int64) (*tracestorepb.SearchResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.m.Lock()
	defer m.m.Unlock()
	traceQlSearchResponse := &tracestorepb.SearchResponse{
		Traces: m.searchResponse,
	}

	m.searchesCount++
	return traceQlSearchResponse, m.err
}

func (m *MockHTTPClient) GetSearchesCount() int {
	m.m.Lock()
	defer m.m.Unlock()
	return m.searchesCount
}

//nolint:all
func (m *MockHTTPClient) SetOverrides(limits *userconfigurableoverrides.Limits, version string) (string, error) {
	panic("unimplemented")
}

//nolint:all
func (m *MockHTTPClient) WithTransport(t http.RoundTripper) {
	panic("unimplemented")
}
