package ingester

import (
	"example.com/acme/tracestore/modules/generator/registry"
	"example.com/acme/tracestore/modules/overrides"
	"example.com/acme/tracestore/tracestoredb/backend"
)

type ingesterOverrides interface {
	registry.Overrides

	DedicatedColumns(userID string) backend.DedicatedColumns
}

var _ ingesterOverrides = (overrides.Interface)(nil)
