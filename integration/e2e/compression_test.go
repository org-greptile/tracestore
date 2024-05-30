package e2e

import (
	"compress/gzip"
	"net/http"
	"testing"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"example.com/acme/e2e"
	"github.com/stretchr/testify/require"

	util "example.com/acme/tracestore/integration"
	"example.com/acme/tracestore/pkg/httpclient"
	"example.com/acme/tracestore/pkg/tracestorepb"
	tracestoreUtil "example.com/acme/tracestore/pkg/util"
)

const (
	configCompression = "config-all-in-one-local.yaml"
)

func TestCompression(t *testing.T) {
	s, err := e2e.NewScenario("tracestore_e2e")
	require.NoError(t, err)
	defer s.Close()

	require.NoError(t, util.CopyFileToSharedDir(s, configCompression, "config.yaml"))
	tracestore := util.NewTracestoreAllInOne()
	require.NoError(t, s.StartAndWaitReady(tracestore))

	// Get port for the Jaeger gRPC receiver endpoint
	c, err := util.NewJaegerGRPCClient(tracestore.Endpoint(14250))
	require.NoError(t, err)
	require.NotNil(t, c)

	info := tracestoreUtil.NewTraceInfo(time.Now(), "")
	require.NoError(t, info.EmitAllBatches(c))

	apiClient := httpclient.New("http://"+tracestore.Endpoint(3200), "")

	apiClientWithCompression := httpclient.NewWithCompression("http://"+tracestore.Endpoint(3200), "")

	queryAndAssertTrace(t, apiClient, info)
	queryAndAssertTraceCompression(t, apiClientWithCompression, info)
}

func queryAndAssertTraceCompression(t *testing.T, client *httpclient.Client, info *tracestoreUtil.TraceInfo) {
	// The received client will strip the header before we have a chance to inspect it, so just validate that the compressed client works as expected.
	result, err := client.QueryTrace(info.HexID())
	require.NoError(t, err)
	require.NotNil(t, result)

	expected, err := info.ConstructTraceFromEpoch()
	require.NoError(t, err)
	assertEqualTrace(t, result, expected)

	// Go's http.Client transparently requests gzip compression and automatically decompresses the
	// response, to disable this behaviour you have to explicitly set the Accept-Encoding header.

	// Make the call directly so we have a chance to inspect the response header and manually un-gzip it ourselves to confirm the content.
	request, err := http.NewRequest("GET", client.BaseURL+httpclient.QueryTraceEndpoint+"/"+info.HexID(), nil)
	require.NoError(t, err)
	request.Header.Add("Accept-Encoding", "gzip")

	res, err := client.Do(request)
	require.NoError(t, err)
	defer res.Body.Close()

	require.Equal(t, "gzip", res.Header.Get("Content-Encoding"))

	gzipReader, err := gzip.NewReader(res.Body)
	require.NoError(t, err)
	defer gzipReader.Close()

	m := &tracestorepb.Trace{}
	unmarshaller := &jsonpb.Unmarshaler{}
	err = unmarshaller.Unmarshal(gzipReader, m)
	require.NoError(t, err)
	assertEqualTrace(t, expected, m)
}
