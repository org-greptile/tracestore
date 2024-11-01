package api

import (
	"encoding/json"
	"math/rand/v2"
	"testing"

	"example.com/acme/tracestore/pkg/util/test"
	"example.com/acme/tracestore/tracestoredb/backend"
	"github.com/stretchr/testify/require"
)

func TestDedicatedColumnsToJson(t *testing.T) {
	d := NewDedicatedColumnsToJSON()

	testCols := []backend.DedicatedColumns{}
	for i := 0; i < 10; i++ {
		testCols = append(testCols, randoDedicatedCols())
	}

	// do all test cols 2x to test caching
	for i := 0; i < 2; i++ {
		for _, cols := range testCols {
			expectedJSON := dedicatedColsToJSON(t, cols)
			actualJSON, err := d.JSONForDedicatedColumns(cols)
			require.NoError(t, err)

			require.Equal(t, expectedJSON, actualJSON, "iteration %d, cols: %v", i, cols)
		}
	}
}

func dedicatedColsToJSON(t *testing.T, cols backend.DedicatedColumns) string {
	t.Helper()

	proto, err := cols.ToTracestorepb()
	require.NoError(t, err)

	jsonBytes, err := json.Marshal(proto)
	require.NoError(t, err)

	return string(jsonBytes)
}

// randoDedicatedCols generates a random set of cols for testing
func randoDedicatedCols() backend.DedicatedColumns {
	colCount := rand.IntN(5) + 1
	ret := make([]backend.DedicatedColumn, 0, colCount)

	for i := 0; i < colCount; i++ {
		scope := backend.DedicatedColumnScopeSpan
		if rand.IntN(2) == 0 {
			scope = backend.DedicatedColumnScopeResource
		}

		col := backend.DedicatedColumn{
			Scope: scope,
			Name:  test.RandomString(),
			Type:  backend.DedicatedColumnTypeString,
		}

		ret = append(ret, col)
	}

	return ret
}
