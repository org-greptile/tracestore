package registry

import (
	"time"

	"example.com/acme/tracestore/modules/overrides"
)

type Overrides interface {
	MetricsGeneratorMaxActiveSeries(userID string) uint32
	MetricsGeneratorCollectionInterval(userID string) time.Duration
	MetricsGeneratorDisableCollection(userID string) bool
}

var _ Overrides = (overrides.Interface)(nil)
