package blockbuilder

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"example.com/acme/tracestore/pkg/tracestorepb"
	"example.com/acme/tracestore/pkg/util"
	"example.com/acme/tracestore/tracestoredb"
	"example.com/acme/tracestore/tracestoredb/encoding"
	"example.com/acme/tracestore/tracestoredb/wal"
	"golang.org/x/sync/errgroup"
)

const flushConcurrency = 4

type partitionSectionWriter interface {
	pushBytes(ts time.Time, tenant string, req *tracestorepb.PushBytesRequest) error
	flush(ctx context.Context, store tracestoredb.Writer) error
}

type writer struct {
	logger log.Logger

	blockCfg      BlockConfig
	partition     uint64
	startOffset   uint64
	startTime     time.Time
	cycleDuration time.Duration
	slackDuration time.Duration

	overrides Overrides
	wal       *wal.WAL
	enc       encoding.VersionedEncoding

	mtx sync.Mutex
	m   map[string]*tenantStore
}

func newPartitionSectionWriter(logger log.Logger, partition, firstOffset uint64, startTime time.Time, cycleDuration, slackDuration time.Duration, blockCfg BlockConfig, overrides Overrides, wal *wal.WAL, enc encoding.VersionedEncoding) *writer {
	return &writer{
		logger:        logger,
		partition:     partition,
		startOffset:   firstOffset,
		startTime:     startTime,
		cycleDuration: cycleDuration,
		slackDuration: slackDuration,
		blockCfg:      blockCfg,
		overrides:     overrides,
		wal:           wal,
		enc:           enc,
		mtx:           sync.Mutex{},
		m:             make(map[string]*tenantStore),
	}
}

func (p *writer) pushBytes(ts time.Time, tenant string, req *tracestorepb.PushBytesRequest) error {
	level.Debug(p.logger).Log(
		"msg", "pushing bytes",
		"tenant", tenant,
		"num_traces", len(req.Traces),
		"id", idsToString(req.Ids),
	)

	i, err := p.instanceForTenant(tenant)
	if err != nil {
		return err
	}

	for j, trace := range req.Traces {
		if err := i.AppendTrace(req.Ids[j], trace.Slice, ts); err != nil {
			return err
		}
	}

	return nil
}

func (p *writer) flush(ctx context.Context, store tracestoredb.Writer) error {
	// TODO - Retry with backoff?

	// Flush tenants concurrently
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(flushConcurrency)

	for _, i := range p.m {
		g.Go(func() error {
			i := i
			st := time.Now()

			level.Info(p.logger).Log("msg", "flushing tenant", "tenant", i.tenantID)
			err := i.Flush(ctx, store)
			if err != nil {
				return err
			}
			level.Info(p.logger).Log("msg", "flushed tenant", "tenant", i.tenantID, "elapsed", time.Since(st))
			return nil
		})
	}
	return g.Wait()
}

func (p *writer) instanceForTenant(tenant string) (*tenantStore, error) {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	if i, ok := p.m[tenant]; ok {
		return i, nil
	}

	i, err := newTenantStore(tenant, p.partition, p.startOffset, p.startTime, p.cycleDuration, p.slackDuration, p.blockCfg, p.logger, p.wal, p.enc, p.overrides)
	if err != nil {
		return nil, err
	}

	p.m[tenant] = i

	return i, nil
}

func idsToString(ids [][]byte) string {
	b := strings.Builder{}
	b.WriteString("[")
	for i, id := range ids {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(util.TraceIDToHexString(id))
	}
	b.WriteString("]")

	return b.String()
}
