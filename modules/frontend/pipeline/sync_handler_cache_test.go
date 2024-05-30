package pipeline

import (
	"bytes"
	"context"
	"testing"

	"github.com/go-kit/log"
	"github.com/gogo/protobuf/jsonpb"
	"example.com/acme/tracestore/pkg/cache"
	"example.com/acme/tracestore/pkg/tracestorepb"
	"example.com/acme/tracestore/pkg/util/test"
	"github.com/stretchr/testify/require"
)

func TestNilProvider(t *testing.T) {
	c := newFrontendCache(nil, cache.RoleFrontendSearch, log.NewNopLogger())
	require.Nil(t, c)
}

func TestCacheCaches(t *testing.T) {
	expected := &tracestorepb.SearchTagsResponse{
		TagNames: []string{"foo", "bar"},
	}

	// marshal mesage to bytes
	buf := bytes.NewBuffer([]byte{})
	err := (&jsonpb.Marshaler{}).Marshal(buf, expected)
	require.NoError(t, err)

	testKey := "key"
	testData := buf.Bytes()

	p := test.NewMockProvider()
	c := newFrontendCache(p, cache.RoleBloom, log.NewNopLogger())
	require.NotNil(t, c)

	// create response
	c.store(context.Background(), testKey, testData)

	actual := &tracestorepb.SearchTagsResponse{}
	found := c.fetch(testKey, actual)

	require.True(t, found)
	require.Equal(t, expected, actual)
}
