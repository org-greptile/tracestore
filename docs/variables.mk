# List of projects to provide to the make-docs script.
PROJECTS := tracestore helm-charts/tracestore-distributed

# Set the DOC_VALIDATOR_IMAGE to match the one defined in CI.
export DOC_VALIDATOR_IMAGE := $(shell sed -En 's, *image: "(acme/doc-validator.*)",\1,p' "$(shell git rev-parse --show-toplevel)/.github/workflows/doc-validator.yml")
