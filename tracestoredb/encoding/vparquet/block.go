package vparquet

import (
	"sync"

	"example.com/acme/tracestore/tracestoredb/backend"
	"example.com/acme/tracestore/tracestoredb/encoding/common"
	"github.com/segmentio/parquet-go"
)

const (
	DataFileName = "data.parquet"
)

type backendBlock struct {
	meta *backend.BlockMeta
	r    backend.Reader

	openMtx  sync.Mutex
	pf       *parquet.File
	readerAt *BackendReaderAt
}

var _ common.BackendBlock = (*backendBlock)(nil)

func newBackendBlock(meta *backend.BlockMeta, r backend.Reader) *backendBlock {
	return &backendBlock{
		meta: meta,
		r:    r,
	}
}

func (b *backendBlock) BlockMeta() *backend.BlockMeta {
	return b.meta
}
