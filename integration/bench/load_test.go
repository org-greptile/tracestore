package load

import (
	"fmt"
	"path/filepath"
	"testing"

	util "example.com/acme/tracestore/integration"

	"example.com/acme/e2e"
	e2e_db "example.com/acme/e2e/db"
	"github.com/stretchr/testify/require"
)

const (
	k6Image = "loadimpact/k6:latest"
)

func TestAllInOne(t *testing.T) {
	s, err := e2e.NewScenario("tracestore_e2e")
	require.NoError(t, err)
	defer s.Close()

	minio := e2e_db.NewMinio(9000, "tracestore")
	require.NotNil(t, minio)
	require.NoError(t, s.StartAndWaitReady(minio))

	require.NoError(t, util.CopyFileToSharedDir(s, "config.yaml", "config.yaml"))
	require.NoError(t, util.CopyFileToSharedDir(s, "smoke_test.js", "smoke_test.js"))
	require.NoError(t, util.CopyFileToSharedDir(s, "stress_test_write_path.js", "stress_test_write_path.js"))
	require.NoError(t, util.CopyFileToSharedDir(s, "modules/util.js", "modules/util.js"))

	tracestore := util.NewTracestoreAllInOne()
	require.NoError(t, s.StartAndWaitReady(tracestore))

	k6 := newK6Runner(tracestore)
	require.NoError(t, s.StartAndWaitReady(k6))

	require.NoError(t, runK6Test(k6, "smoke_test.js"))
	require.NoError(t, runK6Test(k6, "stress_test_write_path.js"))
}

func runK6Test(k6 *e2e.ConcreteService, testjs string) error {
	fmt.Println("------ " + testjs + " ------")
	stdout, stderr, err := k6.Exec(e2e.NewCommand("k6", "run", "--quiet", "--log-output", "none", filepath.Join(e2e.ContainerSharedDir, testjs)))
	fmt.Println("------ stdout ------")
	fmt.Println(stdout)

	if err != nil {
		fmt.Println("------ stderr ------")
		fmt.Println(stderr)
	}

	return err
}

func newK6Runner(tracestore *e2e.HTTPService) *e2e.ConcreteService {
	s := e2e.NewConcreteService(
		"k6",
		k6Image,
		e2e.NewCommandWithoutEntrypoint("sh", "-c", "sleep 3600"),
		e2e.NewCmdReadinessProbe(e2e.NewCommand("sh", "-c", "")),
	)

	s.SetUser("0") // required so k6 can read the js files passed in

	tracestoreHTTP := "http://" + tracestore.NetworkEndpoint(3200)
	tracestoreZipkin := "http://" + tracestore.NetworkEndpoint(9411)
	s.SetEnvVars(map[string]string{
		"WRITE_ENDPOINT":       tracestoreZipkin,
		"DISTRIBUTOR_ENDPOINT": tracestoreHTTP,
		"INGESTER_ENDPOINT":    tracestoreHTTP,
		"QUERY_ENDPOINT":       tracestoreHTTP,
		"QUERIER_ENDPOINT":     tracestoreHTTP,
	})

	return s
}
