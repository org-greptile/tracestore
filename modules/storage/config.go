package storage

import (
	"flag"
	"time"

	"example.com/acme/tracestore/pkg/cache"
	"example.com/acme/tracestore/pkg/util"
	"example.com/acme/tracestore/tracestoredb"
	"example.com/acme/tracestore/tracestoredb/backend"
	"example.com/acme/tracestore/tracestoredb/backend/azure"
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
	f.StringVar(&cfg.Trace.Azure.StorageAccountName, util.PrefixConfig(prefix, "trace.azure.storage_account_name"), "", "Azure storage account name.")
	f.Var(&cfg.Trace.Azure.StorageAccountKey, util.PrefixConfig(prefix, "trace.azure.storage_account_key"), "Azure storage access key.")
	f.StringVar(&cfg.Trace.Azure.ContainerName, util.PrefixConfig(prefix, "trace.azure.container_name"), "", "Azure container name to store blocks in.")
	f.StringVar(&cfg.Trace.Azure.Prefix, util.PrefixConfig(prefix, "trace.azure.prefix"), "", "Azure container prefix to store blocks in.")
	f.StringVar(&cfg.Trace.Azure.Endpoint, util.PrefixConfig(prefix, "trace.azure.endpoint"), "blob.core.windows.net", "Azure endpoint to push blocks to.")
	f.IntVar(&cfg.Trace.Azure.MaxBuffers, util.PrefixConfig(prefix, "trace.azure.max_buffers"), 4, "Number of simultaneous uploads.")
	cfg.Trace.Azure.BufferSize = 3 * 1024 * 1024
	cfg.Trace.Azure.HedgeRequestsUpTo = 2

	cfg.Trace.S3 = &s3.Config{}
	f.StringVar(&cfg.Trace.S3.Bucket, util.PrefixConfig(prefix, "trace.s3.bucket"), "", "s3 bucket to store blocks in.")
	f.StringVar(&cfg.Trace.S3.Prefix, util.PrefixConfig(prefix, "trace.s3.prefix"), "", "s3 root directory to store blocks in.")
	f.StringVar(&cfg.Trace.S3.Endpoint, util.PrefixConfig(prefix, "trace.s3.endpoint"), "", "s3 endpoint to push blocks to.")
	f.StringVar(&cfg.Trace.S3.AccessKey, util.PrefixConfig(prefix, "trace.s3.access_key"), "", "s3 access key.")
	f.StringVar(&cfg.Trace.S3.MinVersion, util.PrefixConfig(prefix, "trace.s3.tls_min_version"), "VersionTLS12", "minimum version of TLS to use when connecting to s3.")
	f.Var(&cfg.Trace.S3.SecretKey, util.PrefixConfig(prefix, "trace.s3.secret_key"), "s3 secret key.")
	f.Var(&cfg.Trace.S3.SessionToken, util.PrefixConfig(prefix, "trace.s3.session_token"), "s3 session token.")
	cfg.Trace.S3.HedgeRequestsUpTo = 2

	cfg.Trace.GCS = &gcs.Config{}
	f.StringVar(&cfg.Trace.GCS.BucketName, util.PrefixConfig(prefix, "trace.gcs.bucket"), "", "gcs bucket to store traces in.")
	f.StringVar(&cfg.Trace.GCS.Prefix, util.PrefixConfig(prefix, "trace.gcs.prefix"), "", "gcs bucket prefix to store traces in.")
	cfg.Trace.GCS.ChunkBufferSize = 10 * 1024 * 1024
	cfg.Trace.GCS.HedgeRequestsUpTo = 2

	cfg.Trace.Local = &local.Config{}
	f.StringVar(&cfg.Trace.Local.Path, util.PrefixConfig(prefix, "trace.local.path"), "", "path to store traces at.")

	cfg.Trace.BackgroundCache = &cache.BackgroundConfig{}
	cfg.Trace.BackgroundCache.WriteBackBuffer = 10000
	cfg.Trace.BackgroundCache.WriteBackGoroutines = 10

	cfg.Trace.Pool = &pool.Config{}
	f.IntVar(&cfg.Trace.Pool.MaxWorkers, util.PrefixConfig(prefix, "trace.pool.max-workers"), 400, "Workers in the worker pool.")
	f.IntVar(&cfg.Trace.Pool.QueueDepth, util.PrefixConfig(prefix, "trace.pool.queue-depth"), 20000, "Work item queue depth.")
}
