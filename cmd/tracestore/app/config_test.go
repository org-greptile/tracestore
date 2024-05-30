package app

import (
	"testing"
	"time"

	"example.com/acme/tracestore/tracestoredb/backend/s3"
	"github.com/stretchr/testify/assert"

	"example.com/acme/tracestore/modules/distributor"
	"example.com/acme/tracestore/modules/storage"
	"example.com/acme/tracestore/tracestoredb"
	"example.com/acme/tracestore/tracestoredb/backend"
	"example.com/acme/tracestore/tracestoredb/encoding/common"
	v2 "example.com/acme/tracestore/tracestoredb/encoding/v2"
	"example.com/acme/tracestore/tracestoredb/encoding/vparquet3"
)

func TestConfig_CheckConfig(t *testing.T) {
	tt := []struct {
		name   string
		config *Config
		expect []ConfigWarning
	}{
		{
			name:   "check default cfg and expect no warnings",
			config: newDefaultConfig(),
			expect: nil,
		},
		{
			name: "hit all except local backend warnings",
			config: &Config{
				Target: MetricsGenerator,
				StorageConfig: storage.Config{
					Trace: tracestoredb.Config{
						Cache:         "supercache",
						Backend:       backend.S3,
						BlocklistPoll: time.Minute,
						Block: &common.BlockConfig{
							Version: "v2",
						},
						S3: &s3.Config{
							NativeAWSAuthEnabled: true,
						},
					},
				},
				Distributor: distributor.Config{
					LogReceivedSpans: distributor.LogReceivedSpansConfig{
						Enabled: true,
					},
				},
			},
			expect: []ConfigWarning{
				warnCompleteBlockTimeout,
				warnBlockRetention,
				warnRetentionConcurrency,
				warnStorageTraceBackendS3,
				warnBlocklistPollConcurrency,
				warnLogReceivedTraces,
				warnNativeAWSAuthEnabled,
				warnConfiguredLegacyCache,
			},
		},
		{
			name: "hit local backend warnings",
			config: func() *Config {
				cfg := newDefaultConfig()
				cfg.StorageConfig.Trace = tracestoredb.Config{
					Backend:                  backend.Local,
					BlocklistPollConcurrency: 1,
					Block: &common.BlockConfig{
						Version: "v2",
					},
				}
				cfg.Target = "something"
				return cfg
			}(),
			expect: []ConfigWarning{warnStorageTraceBackendLocal},
		},
		{
			name: "warnings for v2 settings when they drift from default",
			config: func() *Config {
				cfg := newDefaultConfig()
				cfg.StorageConfig.Trace.Block.Version = vparquet3.VersionString
				cfg.StorageConfig.Trace.Block.IndexDownsampleBytes = 1
				cfg.StorageConfig.Trace.Block.IndexPageSizeBytes = 1
				cfg.Compactor.Compactor.ChunkSizeBytes = 1
				cfg.Compactor.Compactor.FlushSizeBytes = 1
				cfg.Compactor.Compactor.IteratorBufferSize = 1
				return cfg
			}(),
			expect: []ConfigWarning{
				newV2Warning("v2_index_downsample_bytes"),
				newV2Warning("v2_index_page_size_bytes"),
				newV2Warning("v2_in_buffer_bytes"),
				newV2Warning("v2_out_buffer_bytes"),
				newV2Warning("v2_prefetch_traces_count"),
			},
		},
		{
			name: "no warnings for v2 settings when they drift from default and v2 is the block version",
			config: func() *Config {
				cfg := newDefaultConfig()
				cfg.StorageConfig.Trace.Block.Version = v2.VersionString
				cfg.StorageConfig.Trace.Block.IndexDownsampleBytes = 1
				cfg.StorageConfig.Trace.Block.IndexPageSizeBytes = 1
				cfg.Compactor.Compactor.ChunkSizeBytes = 1
				cfg.Compactor.Compactor.FlushSizeBytes = 1
				cfg.Compactor.Compactor.IteratorBufferSize = 1
				return cfg
			}(),
			expect: nil,
		},
		{
			name: "trace storage conflicts with overrides storage - local",
			config: func() *Config {
				cfg := newDefaultConfig()
				cfg.StorageConfig.Trace.Backend = backend.Local
				cfg.StorageConfig.Trace.Local.Path = "/var/tracestore"
				cfg.Overrides.UserConfigurableOverridesConfig.Client.Backend = backend.Local
				cfg.Overrides.UserConfigurableOverridesConfig.Client.Local.Path = "/var/tracestore"
				return cfg
			}(),
			expect: []ConfigWarning{warnTracesAndUserConfigurableOverridesStorageConflict},
		},
		{
			name: "trace storage conflicts with overrides storage - gcs",
			config: func() *Config {
				cfg := newDefaultConfig()
				cfg.StorageConfig.Trace.Backend = backend.GCS
				cfg.StorageConfig.Trace.GCS.BucketName = "bucketname"
				cfg.StorageConfig.Trace.GCS.Prefix = "tracestore"
				cfg.Overrides.UserConfigurableOverridesConfig.Client.Backend = backend.GCS
				cfg.Overrides.UserConfigurableOverridesConfig.Client.GCS.BucketName = "bucketname"
				cfg.Overrides.UserConfigurableOverridesConfig.Client.GCS.Prefix = "tracestore"
				return cfg
			}(),
			expect: []ConfigWarning{warnTracesAndUserConfigurableOverridesStorageConflict},
		},
		{
			name: "trace storage conflicts with overrides storage - different backends",
			config: func() *Config {
				cfg := newDefaultConfig()
				cfg.StorageConfig.Trace.Backend = backend.GCS
				cfg.StorageConfig.Trace.GCS.BucketName = "my-bucket"
				cfg.Overrides.UserConfigurableOverridesConfig.Client.Backend = backend.S3
				cfg.Overrides.UserConfigurableOverridesConfig.Client.S3.Bucket = "my-bucket"
				return cfg
			}(),
			expect: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			warnings := tc.config.CheckConfig()
			assert.Equal(t, tc.expect, warnings)
		})
	}
}
