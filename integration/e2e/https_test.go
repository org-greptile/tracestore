package e2e

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"testing"
	"time"

	"example.com/acme/e2e"
	"example.com/acme/tracestore/cmd/tracestore/app"
	util "example.com/acme/tracestore/integration"
	"example.com/acme/tracestore/integration/e2e/backend"
	e2e_ca "example.com/acme/tracestore/integration/e2e/ca"
	"example.com/acme/tracestore/pkg/httpclient"
	tracestoreUtil "example.com/acme/tracestore/pkg/util"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/credentials"
	"gopkg.in/yaml.v2"
)

const (
	configHTTPS = "config-https.yaml"
)

func TestHTTPS(t *testing.T) {
	km := e2e_ca.SetupCertificates(t)

	s, err := e2e.NewScenario("tracestore_e2e")
	require.NoError(t, err)
	defer s.Close()

	// set up the backend
	cfg := app.Config{}
	buff, err := os.ReadFile(configHTTPS)
	require.NoError(t, err)
	err = yaml.UnmarshalStrict(buff, &cfg)
	require.NoError(t, err)
	_, err = backend.New(s, cfg)
	require.NoError(t, err)

	// copy in certs
	require.NoError(t, util.CopyFileToSharedDir(s, km.ServerCertFile, "tls.crt"))
	require.NoError(t, util.CopyFileToSharedDir(s, km.ServerKeyFile, "tls.key"))
	require.NoError(t, util.CopyFileToSharedDir(s, km.CaCertFile, "ca.crt"))

	require.NoError(t, util.CopyFileToSharedDir(s, configHTTPS, "config.yaml"))
	tracestore := util.NewTracestoreAllInOneWithReadinessProbe(e2e.NewHTTPReadinessProbe(3201, "/ready", 200, 299))
	require.NoError(t, s.StartAndWaitReady(tracestore))

	// Get port for the Jaeger gRPC receiver endpoint
	c, err := util.NewJaegerGRPCClient(tracestore.Endpoint(14250))
	require.NoError(t, err)
	require.NotNil(t, c)

	time.Sleep(10 * time.Second)

	info := tracestoreUtil.NewTraceInfo(time.Now(), "")
	require.NoError(t, info.EmitAllBatches(c))

	apiClient := httpclient.New("https://"+tracestore.Endpoint(3200), "")

	// trust bad certs
	defaultTransport := http.DefaultTransport.(*http.Transport).Clone()
	defaultTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	apiClient.WithTransport(defaultTransport)

	echoReq, err := http.NewRequest("GET", "https://"+tracestore.Endpoint(3200)+"/api/echo", nil)
	require.NoError(t, err)
	resp, err := apiClient.Do(echoReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// query an in-memory trace
	queryAndAssertTrace(t, apiClient, info)
	util.SearchAndAssertTrace(t, apiClient, info)
	util.SearchTraceQLAndAssertTrace(t, apiClient, info)

	creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
	grpcClient, err := util.NewSearchGRPCClientWithCredentials(context.Background(), tracestore.Endpoint(3200), creds)
	require.NoError(t, err)

	now := time.Now()
	util.SearchStreamAndAssertTrace(t, context.Background(), grpcClient, info, now.Add(-time.Hour).Unix(), now.Add(time.Hour).Unix())
}
