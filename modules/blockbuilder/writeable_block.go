package blockbuilder

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"example.com/acme/tracestore/tracestoredb"
	"example.com/acme/tracestore/tracestoredb/backend"
	"example.com/acme/tracestore/tracestoredb/encoding"
	"example.com/acme/tracestore/tracestoredb/encoding/common"
	"go.uber.org/atomic"
)

// Overrides is just the set of overrides needed here.
type Overrides interface {
	MaxBytesPerTrace(string) int
	DedicatedColumns(string) backend.DedicatedColumns
}

const nameFlushed = "flushed"

var _ tracestoredb.WriteableBlock = (*WriteableBlock)(nil)

// WriteableBlock is a trimmed down version of ingester.LocalBlock
type WriteableBlock struct {
	backendBlock common.BackendBlock

	reader backend.Reader
	writer backend.Writer

	flushedTime atomic.Int64 // protecting flushedTime b/c it's accessed from the store on flush and from the ingester instance checking flush time
}

func NewWriteableBlock(backendBlock common.BackendBlock, r backend.Reader, w backend.Writer) tracestoredb.WriteableBlock {
	return &WriteableBlock{
		backendBlock: backendBlock,
		reader:       r,
		writer:       w,
	}
}

func (b *WriteableBlock) BlockMeta() *backend.BlockMeta { return b.backendBlock.BlockMeta() }

func (b *WriteableBlock) Write(ctx context.Context, w backend.Writer) error {
	err := encoding.CopyBlock(ctx, b.BlockMeta(), b.reader, w)
	if err != nil {
		return fmt.Errorf("error copying block from local to remote backend: %w", err)
	}

	err = b.setFlushed(ctx)
	return err
}

// TODO - Necessary?
func (b *WriteableBlock) setFlushed(ctx context.Context) error {
	flushedTime := time.Now()
	flushedBytes, err := flushedTime.MarshalText()
	if err != nil {
		return fmt.Errorf("error marshalling flush time to text: %w", err)
	}

	err = b.writer.Write(ctx, nameFlushed, (uuid.UUID)(b.BlockMeta().BlockID), b.BlockMeta().TenantID, flushedBytes, nil)
	if err != nil {
		return fmt.Errorf("error writing ingester block flushed file: %w", err)
	}

	b.flushedTime.Store(flushedTime.Unix())
	return nil
}
