To run the integration tests, use the following commands


`-count=1` is passed to disable cache during test runs.

```
# build latest image
make docker-tracestore
make docker-tracestore-query

# run all tests
go test -count=1 -v ./integration/e2e/...

# run a particular test "TestMicroservices"
go test -count=1 -v ./integration/e2e/... -run TestMicroservices$

# build and run a particular test "TestMicroservicesWithKVStores"
make docker-tracestore && go test -count=1 -v ./integration/e2e/... -run TestMicroservicesWithKVStores$
```
