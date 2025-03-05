package querier

import (
	"context"
	"sort"
	"testing"

	"example.com/acme/kit/user"
	generator_client "example.com/acme/tracestore/modules/generator/client"
	ingester_client "example.com/acme/tracestore/modules/ingester/client"
	"example.com/acme/tracestore/modules/overrides"
	"example.com/acme/tracestore/pkg/tracestorepb"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestVirtualTagsDoesntHitBackend(t *testing.T) {
	o, err := overrides.NewOverrides(overrides.Config{}, nil, prometheus.DefaultRegisterer)
	require.NoError(t, err)

	q, err := New(Config{}, ingester_client.Config{}, nil, generator_client.Config{}, nil, nil, o)
	require.NoError(t, err)

	ctx := user.InjectOrgID(context.Background(), "blerg")

	// duration should return nothing
	resp, err := q.SearchTagValuesV2(ctx, &tracestorepb.SearchTagValuesRequest{
		TagName: "duration",
	})
	require.NoError(t, err)
	require.Equal(t, &tracestorepb.SearchTagValuesV2Response{Metrics: &tracestorepb.MetadataMetrics{}}, resp)

	// traceDuration should return nothing
	resp, err = q.SearchTagValuesV2(ctx, &tracestorepb.SearchTagValuesRequest{
		TagName: "traceDuration",
	})
	require.NoError(t, err)
	require.Equal(t, &tracestorepb.SearchTagValuesV2Response{Metrics: &tracestorepb.MetadataMetrics{}}, resp)

	// status should return a static list
	resp, err = q.SearchTagValuesV2(ctx, &tracestorepb.SearchTagValuesRequest{
		TagName: "status",
	})
	require.NoError(t, err)
	sort.Slice(resp.TagValues, func(i, j int) bool { return resp.TagValues[i].Value < resp.TagValues[j].Value })
	require.Equal(t, &tracestorepb.SearchTagValuesV2Response{
		TagValues: []*tracestorepb.TagValue{
			{
				Type:  "keyword",
				Value: "error",
			},
			{
				Type:  "keyword",
				Value: "ok",
			},
			{
				Type:  "keyword",
				Value: "unset",
			},
		},
		Metrics: &tracestorepb.MetadataMetrics{},
	}, resp)

	// kind should return a static list
	resp, err = q.SearchTagValuesV2(ctx, &tracestorepb.SearchTagValuesRequest{
		TagName: "kind",
	})
	require.NoError(t, err)
	sort.Slice(resp.TagValues, func(i, j int) bool { return resp.TagValues[i].Value < resp.TagValues[j].Value })
	require.Equal(t, &tracestorepb.SearchTagValuesV2Response{
		TagValues: []*tracestorepb.TagValue{
			{
				Type:  "keyword",
				Value: "client",
			},
			{
				Type:  "keyword",
				Value: "consumer",
			},
			{
				Type:  "keyword",
				Value: "internal",
			},
			{
				Type:  "keyword",
				Value: "producer",
			},
			{
				Type:  "keyword",
				Value: "server",
			},
			{
				Type:  "keyword",
				Value: "unspecified",
			},
		},
		Metrics: &tracestorepb.MetadataMetrics{},
	}, resp)

	// this should error b/c it will attempt to hit the un-configured backend
	resp, err = q.SearchTagValuesV2(ctx, &tracestorepb.SearchTagValuesRequest{
		TagName: ".foo",
	})
	require.Error(t, err)
	require.Nil(t, resp)
}
