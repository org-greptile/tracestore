**Running the integration tests**


`-count=1` is passed to disable cache during test runs.

```sh
# build latest image
make docker-tracestore
make docker-tracestore-query

# run all tests
go test -count=1 -v ./integration/e2e/...

# run a particular test "TestMicroservices"
go test -count=1 -v ./integration/e2e/... -run TestMicroservices$

# build and run a particular test "TestMicroservicesWithKVStores"
make docker-tracestore && go test -count=1 -v ./integration/e2e/... -run TestMicroservicesWithKVStores$

# run a single e2e tests with timeout
go test -timeout 3m -count=1 -v ./integration/e2e/... -run ^TestMultiTenantSearch$

# follow and watch logs while tests are running (assuming e2e test container is named tracestore_e2e-tracestore)
docker logs $(docker container ls -f name=tracestore_e2e-tracestore -q) -f
```

**How to debug Tracestore while running an integration test**

1. Build latest debug image
    ```sh
        make docker-tracestore-debug
    ```
2. Use the function ``NewTracestoreAllInOneDebug`` in your test to spin a Tracestore instance with debug capabilities
3. Set a breakpoint after ``require.NoError(t, s.StartAndWaitReady(tracestore))`` and before the action you want debug
4. Get the port of Delve debugger inside the container
    ```sh
    docker ps --format '{{.Ports}}'  
        # 0.0.0.0:53467->2345
    ```
5. Run the debugger against that port as is specified [here](https://example.com/acme/tracestore/tree/main/example/docker-compose/debug)
