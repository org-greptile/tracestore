package integration

// Collection of utilities to share between our various load tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"example.com/acme/kit/backoff"
	"example.com/acme/e2e"
	jaeger_grpc "github.com/jaegertracing/jaeger/cmd/agent/app/reporter/grpc"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"example.com/acme/tracestore/pkg/tracestorepb"
	tracestoreUtil "example.com/acme/tracestore/pkg/util"
)

const (
	image = "tracestore:latest"
)

// GetExtraArgs returns the extra args to pass to the Docker command used to run Tracestore.
func GetExtraArgs() []string {
	// Get extra args from the TRACESTORE_EXTRA_ARGS env variable
	// falling back to an empty list
	if os.Getenv("TRACESTORE_EXTRA_ARGS") != "" {
		return strings.Fields(os.Getenv("TRACESTORE_EXTRA_ARGS"))
	}

	return nil
}

func buildArgsWithExtra(args []string) []string {
	extraArgs := GetExtraArgs()
	if len(extraArgs) > 0 {
		return append(extraArgs, args...)
	}

	return args
}

func NewTracestoreAllInOne() *e2e.HTTPService {
	args := []string{"-config.file=" + filepath.Join(e2e.ContainerSharedDir, "config.yaml")}
	args = buildArgsWithExtra(args)

	s := e2e.NewHTTPService(
		"tracestore",
		image,
		e2e.NewCommandWithoutEntrypoint("/tracestore", args...),
		e2e.NewHTTPReadinessProbe(3200, "/ready", 200, 299),
		3200,  // http all things
		14250, // jaeger grpc ingest
		9411,  // zipkin ingest (used by load)
		4317,  // otlp grpc
	)

	s.SetBackoff(TracestoreBackoff())

	return s
}

func NewTracestoreDistributor() *e2e.HTTPService {
	args := []string{"-config.file=" + filepath.Join(e2e.ContainerSharedDir, "config.yaml"), "-target=distributor"}
	args = buildArgsWithExtra(args)

	s := e2e.NewHTTPService(
		"distributor",
		image,
		e2e.NewCommandWithoutEntrypoint("/tracestore", args...),
		e2e.NewHTTPReadinessProbe(3200, "/ready", 200, 299),
		3200,
		14250,
	)

	s.SetBackoff(TracestoreBackoff())

	return s
}

func NewTracestoreIngester(replica int) *e2e.HTTPService {
	args := []string{"-config.file=" + filepath.Join(e2e.ContainerSharedDir, "config.yaml"), "-target=ingester"}
	args = buildArgsWithExtra(args)

	s := e2e.NewHTTPService(
		"ingester-"+strconv.Itoa(replica),
		image,
		e2e.NewCommandWithoutEntrypoint("/tracestore", args...),
		e2e.NewHTTPReadinessProbe(3200, "/ready", 200, 299),
		3200,
	)

	s.SetBackoff(TracestoreBackoff())

	return s
}

func NewTracestoreMetricsGenerator() *e2e.HTTPService {
	args := []string{"-config.file=" + filepath.Join(e2e.ContainerSharedDir, "config.yaml"), "-target=metrics-generator"}
	args = buildArgsWithExtra(args)

	s := e2e.NewHTTPService(
		"metrics-generator",
		image,
		e2e.NewCommandWithoutEntrypoint("/tracestore", args...),
		e2e.NewHTTPReadinessProbe(3200, "/ready", 200, 299),
		3200,
	)

	s.SetBackoff(TracestoreBackoff())

	return s
}

func NewTracestoreQueryFrontend() *e2e.HTTPService {
	args := []string{"-config.file=" + filepath.Join(e2e.ContainerSharedDir, "config.yaml"), "-target=query-frontend"}
	args = buildArgsWithExtra(args)

	s := e2e.NewHTTPService(
		"query-frontend",
		image,
		e2e.NewCommandWithoutEntrypoint("/tracestore", args...),
		e2e.NewHTTPReadinessProbe(3200, "/ready", 200, 299),
		3200,
	)

	s.SetBackoff(TracestoreBackoff())

	return s
}

func NewTracestoreQuerier() *e2e.HTTPService {
	args := []string{"-config.file=" + filepath.Join(e2e.ContainerSharedDir, "config.yaml"), "-target=querier"}
	args = buildArgsWithExtra(args)

	s := e2e.NewHTTPService(
		"querier",
		image,
		e2e.NewCommandWithoutEntrypoint("/tracestore", args...),
		e2e.NewHTTPReadinessProbe(3200, "/ready", 200, 299),
		3200,
	)

	s.SetBackoff(TracestoreBackoff())

	return s
}

func NewTracestoreScalableSingleBinary(replica int) *e2e.HTTPService {
	args := []string{"-config.file=" + filepath.Join(e2e.ContainerSharedDir, "config.yaml"), "-target=scalable-single-binary", "-querier.frontend-address=tracestore-" + strconv.Itoa(replica) + ":9095"}
	args = buildArgsWithExtra(args)

	s := e2e.NewHTTPService(
		"tracestore-"+strconv.Itoa(replica),
		image,
		e2e.NewCommandWithoutEntrypoint("/tracestore", args...),
		e2e.NewHTTPReadinessProbe(3200, "/ready", 200, 299),
		3200,  // http all things
		14250, // jaeger grpc ingest
		// 9411,  // zipkin ingest (used by load)
	)

	s.SetBackoff(TracestoreBackoff())

	return s
}

func WriteFileToSharedDir(s *e2e.Scenario, dst string, content []byte) error {
	dst = filepath.Join(s.SharedDir(), dst)

	// Ensure the entire path of directories exist.
	if err := os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
		return err
	}

	return os.WriteFile(
		dst,
		content,
		os.ModePerm)
}

func CopyFileToSharedDir(s *e2e.Scenario, src, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return errors.Wrapf(err, "unable to read local file %s", src)
	}

	return WriteFileToSharedDir(s, dst, content)
}

func TracestoreBackoff() backoff.Config {
	return backoff.Config{
		MinBackoff: 500 * time.Millisecond,
		MaxBackoff: time.Second,
		MaxRetries: 300, // Sometimes the CI is slow ¯\_(ツ)_/¯
	}
}

func NewJaegerGRPCClient(endpoint string) (*jaeger_grpc.Reporter, error) {
	// new jaeger grpc exporter
	conn, err := grpc.Dial(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	return jaeger_grpc.NewReporter(conn, nil, logger), err
}

func SearchAndAssertTrace(t *testing.T, client *tracestoreUtil.Client, info *tracestoreUtil.TraceInfo) {
	expected, err := info.ConstructTraceFromEpoch()
	require.NoError(t, err)

	attr := tracestoreUtil.RandomAttrFromTrace(expected)

	// NOTE: SearchTags doesn't include live traces anymore
	// so don't check SearchTags

	// verify attribute value is present in tag values
	tagValuesResp, err := client.SearchTagValues(attr.Key)
	require.NoError(t, err)
	require.Contains(t, tagValuesResp.TagValues, attr.GetValue().GetStringValue())

	// verify trace can be found using attribute
	resp, err := client.Search(attr.GetKey() + "=" + attr.GetValue().GetStringValue())
	require.NoError(t, err)

	hasHex := func(hexId string, resp *tracestorepb.SearchResponse) bool {
		for _, s := range resp.Traces {
			equal, err := tracestoreUtil.EqualHexStringTraceIDs(s.TraceID, hexId)
			require.NoError(t, err)
			if equal {
				return true
			}
		}

		return false
	}

	require.True(t, hasHex(info.HexID(), resp))
}

func SearchTraceQLAndAssertTrace(t *testing.T, client *tracestoreUtil.Client, info *tracestoreUtil.TraceInfo) {
	expected, err := info.ConstructTraceFromEpoch()
	require.NoError(t, err)

	attr := tracestoreUtil.RandomAttrFromTrace(expected)
	query := fmt.Sprintf(`{ .%s = "%s"}`, attr.GetKey(), attr.GetValue().GetStringValue())

	resp, err := client.SearchTraceQL(query)
	require.NoError(t, err)

	hasHex := func(hexId string, resp *tracestorepb.SearchResponse) bool {
		for _, s := range resp.Traces {
			equal, err := tracestoreUtil.EqualHexStringTraceIDs(s.TraceID, hexId)
			require.NoError(t, err)
			if equal {
				return true
			}
		}

		return false
	}

	require.True(t, hasHex(info.HexID(), resp))
}

// by passing a time range and using a query_ingesters_until/backend_after of 0 we can force the queriers
// to look in the backend blocks
func SearchAndAssertTraceBackend(t *testing.T, client *tracestoreUtil.Client, info *tracestoreUtil.TraceInfo, start int64, end int64) {
	expected, err := info.ConstructTraceFromEpoch()
	require.NoError(t, err)

	attr := tracestoreUtil.RandomAttrFromTrace(expected)

	// verify trace can be found using attribute and time range
	resp, err := client.SearchWithRange(attr.GetKey()+"="+attr.GetValue().GetStringValue(), start, end)
	require.NoError(t, err)

	hasHex := func(hexId string, resp *tracestorepb.SearchResponse) bool {
		for _, s := range resp.Traces {
			equal, err := tracestoreUtil.EqualHexStringTraceIDs(s.TraceID, hexId)
			require.NoError(t, err)
			if equal {
				return true
			}
		}

		return false
	}

	require.True(t, hasHex(info.HexID(), resp))
}
