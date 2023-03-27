package backend

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"example.com/acme/e2e"
	e2e_db "example.com/acme/e2e/db"

	"example.com/acme/tracestore/cmd/tracestore/app"
	util "example.com/acme/tracestore/integration"
	"example.com/acme/tracestore/tracestoredb/backend/azure"
)

const (
	azuriteImage = "mcr.microsoft.com/azure-storage/azurite"
	gcsImage     = "fsouza/fake-gcs-server:1.36.3"
)

func parsePort(endpoint string) (int, error) {
	substrings := strings.Split(endpoint, ":")
	portStrings := strings.Split(substrings[len(substrings)-1], "/")
	port, err := strconv.Atoi(portStrings[0])
	if err != nil {
		return 0, err
	}
	return port, nil
}

func New(scenario *e2e.Scenario, cfg app.Config) (*e2e.HTTPService, error) {
	var backendService *e2e.HTTPService
	switch cfg.StorageConfig.Trace.Backend {
	case "s3":
		port, err := parsePort(cfg.StorageConfig.Trace.S3.Endpoint)
		if err != nil {
			return nil, err
		}
		backendService = e2e_db.NewMinio(port, "tracestore")
		if backendService == nil {
			return nil, fmt.Errorf("error creating minio backend")
		}
		err = scenario.StartAndWaitReady(backendService)
		if err != nil {
			return nil, err
		}
	case "azure":
		port, err := parsePort(cfg.StorageConfig.Trace.Azure.Endpoint)
		if err != nil {
			return nil, err
		}
		backendService = NewAzurite(port)
		err = scenario.StartAndWaitReady(backendService)
		if err != nil {
			return nil, err
		}
		cfg.StorageConfig.Trace.Azure.Endpoint = backendService.Endpoint(port)
		_, err = azure.CreateContainer(context.TODO(), cfg.StorageConfig.Trace.Azure)
		if err != nil {
			return nil, err
		}
	case "gcs":
		port, err := parsePort(cfg.StorageConfig.Trace.GCS.Endpoint)
		if err != nil {
			return nil, err
		}
		backendService = NewGCS(port)
		if backendService == nil {
			return nil, fmt.Errorf("error creating gcs backend")
		}
		err = scenario.StartAndWaitReady(backendService)
		if err != nil {
			return nil, err
		}
	}

	return backendService, nil
}

func NewAzurite(port int) *e2e.HTTPService {
	s := e2e.NewHTTPService(
		"azurite",
		azuriteImage, // Create the the azurite container
		e2e.NewCommandWithoutEntrypoint("sh", "-c", "azurite -l /data --blobHost 0.0.0.0"),
		e2e.NewHTTPReadinessProbe(port, "/devstoreaccount1?comp=list", 403, 403), //If we get 403 the Azurite is ready
		port, // blob storage port
	)

	s.SetBackoff(util.TracestoreBackoff())

	return s
}

func NewGCS(port int) *e2e.HTTPService {
	commands := []string{
		"mkdir -p /data/tracestore",
		"/bin/fake-gcs-server -data /data -public-host=tracestore_e2e-gcs -port=4443",
	}
	s := e2e.NewHTTPService(
		"gcs",
		gcsImage, // Create the the gcs container
		e2e.NewCommandWithoutEntrypoint("sh", "-c", strings.Join(commands, " && ")),
		e2e.NewHTTPReadinessProbe(port, "/", 400, 400), // for lack of a better way, readiness probe does not support https at the moment
		port,
	)

	s.SetBackoff(util.TracestoreBackoff())

	return s
}
