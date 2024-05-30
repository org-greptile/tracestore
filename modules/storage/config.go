package storage

import (
	"flag"
	"time"

	"example.com/acme/tracestore/pkg/cache"
	"example.com/acme/tracestore/pkg/util"
	"example.com/acme/tracestore/tracestoredb"
	"example.com/acme/tracestore/tracestoredb/backend"
	azure "example.com/acme/tracestore/tracestoredb/backend/azure/config"
	"example.com/acme/tracestore/tracestoredb/backend/gcs"
	"example.com/acme/tracestore/tracestoredb/backend/local"
	"example.com/acme/tracestore/tracestoredb/backend/s3"
	"example.com/acme/tracestore/tracestoredb/encoding"
	"example.com/acme/tracestore/tracestoredb/encoding/common"
	"example.com/acme/tracestore/tracestoredb/pool"
	"example.com/acme/tracestore/tracestoredb/wal"
)

// Config is the Tracestore storage configuration
type Config struct {
	Trace tracestoredb.Config `yaml:"trace"`
}

// RegisterFlagsAndApplyDefaults registers the flags.
func (cfg *Config) RegisterFlagsAndApplyDefaults(prefix string, f *flag.FlagSet) {
	cfg.Trace.BlocklistPollFallback = true
	cfg.Trace.BlocklistPollConcurrency = tracestoredb.DefaultBlocklistPollConcurrency
	cfg.Trace.BlocklistPollTenantIndexBuilders = tracestoredb.DefaultTenantIndexBuilders
	cfg.Trace.BlocklistPollTolerateConsecutiveErrors = tracestoredb.DefaultTolerateConsecutiveErrors

	f.StringVar(&cfg.Trace.Backend, util.PrefixConfig(prefix, "trace.backend"), "", "Trace backend (s3, azure, gcs, local)")
	f.DurationVar(&cfg.Trace.BlocklistPoll, util.PrefixConfig(prefix, "trace.blocklist_poll"), tracestoredb.DefaultBlocklistPoll, "Period at which to run the maintenance cycle.")

	cfg.Trace.WAL = &wal.Config{}
	f.StringVar(&cfg.Trace.WAL.Filepath, util.PrefixConfig(prefix, "trace.wal.path"), "/var/tracestore/wal", "Path at which store WAL blocks.")
	cfg.Trace.WAL.Encoding = backend.EncSnappy
	cfg.Trace.WAL.SearchEncoding = backend.EncNone
	cfg.Trace.WAL.IngestionSlack = 2 * time.Minute

	cfg.Trace.Search = &tracestoredb.SearchConfig{}
	cfg.Trace.Search.RegisterFlagsAndApplyDefaults(prefix, f)

	cfg.Trace.Block = &common.BlockConfig{}
	cfg.Trace.Block.Version = encoding.DefaultEncoding().Version()
	cfg.Trace.Block.RegisterFlagsAndApplyDefaults(prefix, f)

	cfg.Trace.Azure = &azure.Config{}
	cfg.Trace.Azure.RegisterFlagsAndApplyDefaults(util.PrefixConfig(prefix, "trace"), f)

	cfg.Trace.S3 = &s3.Config{}
	cfg.Trace.S3.RegisterFlagsAndApplyDefaults(util.PrefixConfig(prefix, "trace"), f)

	cfg.Trace.GCS = &gcs.Config{}
	cfg.Trace.GCS.RegisterFlagsAndApplyDefaults(util.PrefixConfig(prefix, "trace"), f)

	cfg.Trace.Local = &local.Config{}
	cfg.Trace.Local.RegisterFlagsAndApplyDefaults(util.PrefixConfig(prefix, "trace"), f)

	cfg.Trace.BackgroundCache = &cache.BackgroundConfig{}
	cfg.Trace.BackgroundCache.WriteBackBuffer = 10000
	cfg.Trace.BackgroundCache.WriteBackGoroutines = 10

	cfg.Trace.Pool = &pool.Config{}
	f.IntVar(&cfg.Trace.Pool.MaxWorkers, util.PrefixConfig(prefix, "trace.pool.max-workers"), 400, "Workers in the worker pool.")
	f.IntVar(&cfg.Trace.Pool.QueueDepth, util.PrefixConfig(prefix, "trace.pool.queue-depth"), 20000, "Work item queue depth.")
}
