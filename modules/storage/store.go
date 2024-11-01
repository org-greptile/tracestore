package storage

import (
	"context"

	"github.com/go-kit/log"
	"example.com/acme/kit/services"

	"example.com/acme/tracestore/pkg/cache"
	"example.com/acme/tracestore/pkg/usagestats"
	"example.com/acme/tracestore/tracestoredb"
)

var (
	statCache               = usagestats.NewString("storage_cache")
	statBackend             = usagestats.NewString("storage_backend")
	statWalEncoding         = usagestats.NewString("storage_wal_encoding")
	statWalSearchEncoding   = usagestats.NewString("storage_wal_search_encoding")
	statBlockEncoding       = usagestats.NewString("storage_block_encoding")
	statBlockSearchEncoding = usagestats.NewString("storage_block_search_encoding")
)

// Store wraps the tracestoredb storage layer
type Store interface {
	services.Service

	tracestoredb.Reader
	tracestoredb.Writer
	tracestoredb.Compactor
}

type store struct {
	services.Service

	cfg Config

	tracestoredb.Reader
	tracestoredb.Writer
	tracestoredb.Compactor
}

// NewStore creates a new Tracestore Store using configuration supplied.
func NewStore(cfg Config, cacheProvider cache.Provider, logger log.Logger) (Store, error) {
	statCache.Set(cfg.Trace.Cache)
	statBackend.Set(cfg.Trace.Backend)
	statWalEncoding.Set(cfg.Trace.WAL.Encoding.String())
	statWalSearchEncoding.Set(cfg.Trace.WAL.SearchEncoding.String())
	statBlockEncoding.Set(cfg.Trace.Block.Encoding.String())
	statBlockSearchEncoding.Set(cfg.Trace.Block.SearchEncoding.String())

	r, w, c, err := tracestoredb.New(&cfg.Trace, cacheProvider, logger)
	if err != nil {
		return nil, err
	}

	s := &store{
		cfg:       cfg,
		Reader:    r,
		Writer:    w,
		Compactor: c,
	}

	s.Service = services.NewIdleService(s.starting, s.stopping)
	return s, nil
}

func (s *store) starting(_ context.Context) error {
	return nil
}

func (s *store) stopping(_ error) error {
	s.Reader.Shutdown()

	return nil
}
