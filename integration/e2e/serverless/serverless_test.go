package serverless

import (
	"testing"
	"time"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"

	"example.com/acme/e2e"
	e2e_db "example.com/acme/e2e/db"

	util "example.com/acme/tracestore/integration"
	tracestoreUtil "example.com/acme/tracestore/pkg/util"
)

const (
	configServerlessGCR    = "config-serverless-gcr.yaml"
	configServerlessLambda = "config-serverless-lambda.yaml"
)

func TestServerless(t *testing.T) {

	testClouds := []struct {
		name       string
		serverless *e2e.HTTPService
		config     string
	}{
		{
			name:       "gcr",
			serverless: newTracestoreServerlessGCR(),
			config:     configServerlessGCR,
		},
		{
			name:       "lambda",
			serverless: newTracestoreServerlessLambda(),
			config:     configServerlessLambda,
		},
	}

	for _, tc := range testClouds {
		t.Run(tc.name, func(t *testing.T) {
			s, err := e2e.NewScenario("tracestore_e2e")
			require.NoError(t, err)
			defer s.Close()

			minio := e2e_db.NewMinio(9000, "tracestore")
			require.NotNil(t, minio)
			require.NoError(t, s.StartAndWaitReady(minio))

			require.NoError(t, util.CopyFileToSharedDir(s, tc.config, "config.yaml"))
			tracestoreIngester1 := util.NewTracestoreIngester(1)
			tracestoreIngester2 := util.NewTracestoreIngester(2)
			tracestoreIngester3 := util.NewTracestoreIngester(3)
			tracestoreDistributor := util.NewTracestoreDistributor()
			tracestoreQueryFrontend := util.NewTracestoreQueryFrontend()
			tracestoreQuerier := util.NewTracestoreQuerier()
			tracestoreServerless := tc.serverless
			require.NoError(t, s.StartAndWaitReady(tracestoreIngester1, tracestoreIngester2, tracestoreIngester3, tracestoreDistributor, tracestoreQueryFrontend, tracestoreQuerier, tracestoreServerless))

			// wait for 2 active ingesters
			time.Sleep(1 * time.Second)
			matchers := []*labels.Matcher{
				{
					Type:  labels.MatchEqual,
					Name:  "name",
					Value: "ingester",
				},
				{
					Type:  labels.MatchEqual,
					Name:  "state",
					Value: "ACTIVE",
				},
			}
			require.NoError(t, tracestoreDistributor.WaitSumMetricsWithOptions(e2e.Equals(3), []string{`tracestore_ring_members`}, e2e.WithLabelMatchers(matchers...), e2e.WaitMissingMetrics))

			// Get port for the Jaeger gRPC receiver endpoint
			c, err := util.NewJaegerGRPCClient(tracestoreDistributor.Endpoint(14250))
			require.NoError(t, err)
			require.NotNil(t, c)

			info := tracestoreUtil.NewTraceInfo(time.Now(), "")
			require.NoError(t, info.EmitAllBatches(c))

			// wait trace_idle_time and ensure trace is created in ingester
			require.NoError(t, tracestoreIngester1.WaitSumMetricsWithOptions(e2e.Less(3), []string{"tracestore_ingester_traces_created_total"}, e2e.WaitMissingMetrics))
			require.NoError(t, tracestoreIngester2.WaitSumMetricsWithOptions(e2e.Less(3), []string{"tracestore_ingester_traces_created_total"}, e2e.WaitMissingMetrics))
			require.NoError(t, tracestoreIngester3.WaitSumMetricsWithOptions(e2e.Less(3), []string{"tracestore_ingester_traces_created_total"}, e2e.WaitMissingMetrics))

			features := []*labels.Matcher{
				{
					Type:  labels.MatchEqual,
					Name:  "feature",
					Value: "search_external_endpoints",
				},
			}
			require.NoError(t, tracestoreDistributor.WaitSumMetricsWithOptions(e2e.Equals(1), []string{`tracestore_feature_enabled`}, e2e.WithLabelMatchers(features...), e2e.WaitMissingMetrics))

			apiClient := tracestoreUtil.NewClient("http://"+tracestoreQueryFrontend.Endpoint(3200), "")

			// flush trace to backend
			res, err := e2e.DoGet("http://" + tracestoreIngester1.Endpoint(3200) + "/flush")
			require.NoError(t, err)
			require.Equal(t, 204, res.StatusCode)

			res, err = e2e.DoGet("http://" + tracestoreIngester2.Endpoint(3200) + "/flush")
			require.NoError(t, err)
			require.Equal(t, 204, res.StatusCode)

			res, err = e2e.DoGet("http://" + tracestoreIngester3.Endpoint(3200) + "/flush")
			require.NoError(t, err)
			require.Equal(t, 204, res.StatusCode)

			// zzz
			time.Sleep(10 * time.Second)

			// search the backend. this works b/c we're passing a start/end AND setting query ingesters within min/max to 0
			now := time.Now()
			util.SearchAndAssertTraceBackend(t, apiClient, info, now.Add(-20*time.Minute).Unix(), now.Unix())

		})
	}

}

func newTracestoreServerlessGCR() *e2e.HTTPService {
	s := e2e.NewHTTPService(
		"serverless",
		"tracestore-serverless", // created by Makefile in /cmd/tracestore-serverless
		nil,
		nil,
		8080,
	)

	s.SetEnvVars(map[string]string{
		"TRACESTORE_S3_BUCKET":     "tracestore",
		"TRACESTORE_S3_ENDPOINT":   "tracestore_e2e-minio-9000:9000",
		"TRACESTORE_S3_ACCESS_KEY": e2e_db.MinioAccessKey,
		"TRACESTORE_S3_SECRET_KEY": e2e_db.MinioSecretKey,
		"TRACESTORE_S3_INSECURE":   "true",
		"TRACESTORE_BACKEND":       "s3",
	})

	s.SetBackoff(util.TracestoreBackoff())

	return s
}

func newTracestoreServerlessLambda() *e2e.HTTPService {
	s := e2e.NewHTTPService(
		"serverless",
		"tracestore-serverless-lambda", // created by build-docker-lambda-test make target
		nil,
		nil,
		9000,
	)

	s.SetEnvVars(map[string]string{
		"TRACESTORE_S3_BUCKET":     "tracestore",
		"TRACESTORE_S3_ENDPOINT":   "tracestore_e2e-minio-9000:9000",
		"TRACESTORE_S3_ACCESS_KEY": e2e_db.MinioAccessKey,
		"TRACESTORE_S3_SECRET_KEY": e2e_db.MinioSecretKey,
		"TRACESTORE_S3_INSECURE":   "true",
		"TRACESTORE_BACKEND":       "s3",
	})

	s.SetBackoff(util.TracestoreBackoff())

	return s
}
