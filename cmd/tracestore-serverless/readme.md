# tracestore-serverless

This folder is intended to wrap a simple handler for searching backend storage for traces. Subfolders 
`./cloud-run` and `./lambda` contain specific code necessary to run this handler in the respective
environments.

## Configuration
The serverless handler is setup to be configured via environment variables starting with the `TRACESTORE_` prefix.
```
TRACESTORE_GCS_BUCKET_NAME=some-random-gcs-bucket
TRACESTORE_GCS_CHUNK_BUFFER_SIZE=10485760
TRACESTORE_GCS_HEDGE_REQUESTS_AT=400ms
TRACESTORE_BACKEND=gcs
```
All fields in tracestoredb.Config are accessible using their all caps yaml names. Also config objects can be descended
using the `_` character. Note that in the above example `TRACESTORE_BCS_BUCKET_NAME` refers to tracestoredb.Config.GCS.BucketName.

## Make

### make build-docker

Builds two docker images named `tracestore-serverlesss` and `tracestore-serverless-lambda` that are used for e2e testing.

```
docker run --rm -p 8080:8080 tracestore-serverless
curl http://localhost:8080
```

### make build-docker-gcr-binary
### make build-docker-gcr

Building the GCR docker image is split into two steps to make it easier to build as part of CI. (See /.drone/drone.jsonnet).
Together these two steps build a binary out of container and then package that up using a Dockerfile. The docker image
created in `build-docker-gcr` is the code artifact that is shipped to Google Cloud Run.

### make build-docker-lambda-test

Builds the docker image for lambda e2e testing. This image is not shipped anywhere.

### make build-lambda-zip

Builds a lambda zip file. This builds the code artifact that is shipped to AWS Lambda.

