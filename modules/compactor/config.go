package compactor

import (
	"flag"
	"net"
	"strconv"
	"time"

	"github.com/go-kit/log"
	"example.com/acme/kit/flagext"
	"example.com/acme/kit/ring"
	"example.com/acme/tracestore/pkg/util"
	"example.com/acme/tracestore/tracestoredb"
)

type Config struct {
	Disabled        bool                    `yaml:"disabled,omitempty"`
	ShardingRing    RingConfig              `yaml:"ring,omitempty"`
	Compactor       tracestoredb.CompactorConfig `yaml:"compaction"`
	OverrideRingKey string                  `yaml:"override_ring_key"`
}

// RegisterFlagsAndApplyDefaults registers the flags.
func (cfg *Config) RegisterFlagsAndApplyDefaults(prefix string, f *flag.FlagSet) {
	cfg.Compactor = tracestoredb.CompactorConfig{
		ChunkSizeBytes:          tracestoredb.DefaultChunkSizeBytes, // 5 MiB
		FlushSizeBytes:          tracestoredb.DefaultFlushSizeBytes,
		CompactedBlockRetention: time.Hour,
		RetentionConcurrency:    tracestoredb.DefaultRetentionConcurrency,
		IteratorBufferSize:      tracestoredb.DefaultIteratorBufferSize,
		MaxTimePerTenant:        tracestoredb.DefaultMaxTimePerTenant,
		CompactionCycle:         tracestoredb.DefaultCompactionCycle,
	}

	flagext.DefaultValues(&cfg.ShardingRing)
	cfg.ShardingRing.KVStore.Store = "" // by default compactor is not sharded

	f.DurationVar(&cfg.Compactor.BlockRetention, util.PrefixConfig(prefix, "compaction.block-retention"), 14*24*time.Hour, "Duration to keep blocks/traces.")
	f.IntVar(&cfg.Compactor.MaxCompactionObjects, util.PrefixConfig(prefix, "compaction.max-objects-per-block"), 6000000, "Maximum number of traces in a compacted block.")
	f.Uint64Var(&cfg.Compactor.MaxBlockBytes, util.PrefixConfig(prefix, "compaction.max-block-bytes"), 100*1024*1024*1024 /* 100GB */, "Maximum size of a compacted block.")
	f.DurationVar(&cfg.Compactor.MaxCompactionRange, util.PrefixConfig(prefix, "compaction.compaction-window"), time.Hour, "Maximum time window across which to compact blocks.")
	f.BoolVar(&cfg.Disabled, util.PrefixConfig(prefix, "disabled"), false, "Disable compaction.")
	cfg.OverrideRingKey = compactorRingKey
}

func toBasicLifecyclerConfig(cfg RingConfig, logger log.Logger) (ring.BasicLifecyclerConfig, error) {
	instanceAddr, err := ring.GetInstanceAddr(cfg.InstanceAddr, cfg.InstanceInterfaceNames, logger, cfg.EnableInet6)
	if err != nil {
		return ring.BasicLifecyclerConfig{}, err
	}

	instancePort := ring.GetInstancePort(cfg.InstancePort, cfg.ListenPort)

	instanceAddrPort := net.JoinHostPort(instanceAddr, strconv.Itoa(instancePort))

	return ring.BasicLifecyclerConfig{
		ID:              cfg.InstanceID,
		Addr:            instanceAddrPort,
		HeartbeatPeriod: cfg.HeartbeatPeriod,
		NumTokens:       ringNumTokens,
	}, nil
}
