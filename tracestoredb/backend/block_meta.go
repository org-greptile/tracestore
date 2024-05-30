package backend

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/google/uuid"

	"example.com/acme/tracestore/pkg/tracestorepb"
	"example.com/acme/tracestore/pkg/traceql"
)

// DedicatedColumnType is the type of the values in the dedicated attribute column. Only 'string' is supported.
type DedicatedColumnType string

// DedicatedColumnScope is the scope of the attribute that is stored in a dedicated column. Possible values are
// 'resource' and 'span'.
type DedicatedColumnScope string

const (
	DedicatedColumnTypeString DedicatedColumnType = "string"

	DedicatedColumnScopeResource DedicatedColumnScope = "resource"
	DedicatedColumnScopeSpan     DedicatedColumnScope = "span"

	DefaultDedicatedColumnType  = DedicatedColumnTypeString
	DefaultDedicatedColumnScope = DedicatedColumnScopeSpan

	maxSupportedSpanColumns     = 10
	maxSupportedResourceColumns = 10
)

func DedicatedColumnTypeFromTracestorepb(t tracestorepb.DedicatedColumn_Type) (DedicatedColumnType, error) {
	switch t {
	case tracestorepb.DedicatedColumn_STRING:
		return DedicatedColumnTypeString, nil
	default:
		return "", fmt.Errorf("invalid value for tracestorepb.DedicatedColumn_Type '%v'", t)
	}
}

func (t DedicatedColumnType) ToTracestorepb() (tracestorepb.DedicatedColumn_Type, error) {
	switch t {
	case DedicatedColumnTypeString:
		return tracestorepb.DedicatedColumn_STRING, nil
	default:
		return 0, fmt.Errorf("invalid value for dedicated column type '%v'", t)
	}
}

func (t DedicatedColumnType) ToStaticType() (traceql.StaticType, error) {
	switch t {
	case DedicatedColumnTypeString:
		return traceql.TypeString, nil
	default:
		return traceql.TypeNil, fmt.Errorf("unsupported dedicated column type '%s'", t)
	}
}

func DedicatedColumnScopeFromTracestorepb(s tracestorepb.DedicatedColumn_Scope) (DedicatedColumnScope, error) {
	switch s {
	case tracestorepb.DedicatedColumn_SPAN:
		return DedicatedColumnScopeSpan, nil
	case tracestorepb.DedicatedColumn_RESOURCE:
		return DedicatedColumnScopeResource, nil
	default:
		return "", fmt.Errorf("invalid value for tracestorepb.DedicatedColumn_Scope '%v'", s)
	}
}

func (s DedicatedColumnScope) ToTracestorepb() (tracestorepb.DedicatedColumn_Scope, error) {
	switch s {
	case DedicatedColumnScopeSpan:
		return tracestorepb.DedicatedColumn_SPAN, nil
	case DedicatedColumnScopeResource:
		return tracestorepb.DedicatedColumn_RESOURCE, nil
	default:
		return 0, fmt.Errorf("invalid value for dedicated column scope '%v'", s)
	}
}

type CompactedBlockMeta struct {
	BlockMeta

	CompactedTime time.Time `json:"compactedTime"`
}

// The BlockMeta data that is stored for each individual block.
type BlockMeta struct {
	// A Version that indicates the block format. This includes specifics of how the indexes and data is stored.
	Version string `json:"format"`
	// BlockID is a unique identifier of the block.
	BlockID uuid.UUID `json:"blockID"`
	// MinID is the smallest object id stored in this block.
	MinID []byte `json:"minID"`
	// MaxID is the largest object id stored in this block.
	MaxID []byte `json:"maxID"`
	// A TenantID that defines the tenant to which this block belongs.
	TenantID string `json:"tenantID"`
	// StartTime roughly matches when the first obj was written to this block. It is used to determine block.
	// age for different purposes (caching, etc)
	StartTime time.Time `json:"startTime"`
	// EndTime roughly matches to the time the last obj was written to this block. Is currently mostly meaningless.
	EndTime time.Time `json:"endTime"`
	// TotalObjects counts the number of objects in this block.
	TotalObjects int `json:"totalObjects"`
	// The Size in bytes of the block.
	Size uint64 `json:"size"`
	// CompactionLevel defines the number of times this block has been compacted.
	CompactionLevel uint8 `json:"compactionLevel"`
	// Encoding and compression format (used only in v2)
	Encoding Encoding `json:"encoding"`
	// IndexPageSize holds the size of each index page in bytes (used only in v2)
	IndexPageSize uint32 `json:"indexPageSize"`
	// TotalRecords holds the total Records stored in the index file (used only in v2)
	TotalRecords uint32 `json:"totalRecords"`
	// DataEncoding is tracked by tracestoredb and indicates the way the bytes are encoded.
	DataEncoding string `json:"dataEncoding"`
	// BloomShardCount represents the number of bloom filter shards.
	BloomShardCount uint16 `json:"bloomShards"`
	// FooterSize contains the size of the footer in bytes (used by parquet)
	FooterSize uint32 `json:"footerSize"`
	// DedicatedColumns configuration for attributes (used by vParquet3)
	DedicatedColumns DedicatedColumns `json:"dedicatedColumns,omitempty"`
	// ReplicationFactor is the number of times the data written in this block has been replicated.
	// It's left unset if replication factor is 3. Default is 0 (RF3).
	ReplicationFactor uint32 `json:"replicationFactor,omitempty"`
}

// DedicatedColumn contains the configuration for a single attribute with the given name that should
// be stored in a dedicated column instead of the generic attribute column.
type DedicatedColumn struct {
	// The Scope of the attribute
	Scope DedicatedColumnScope `yaml:"scope" json:"s,omitempty"`
	// The Name of the attribute stored in the dedicated column
	Name string `yaml:"name" json:"n"`
	// The Type of attribute value
	Type DedicatedColumnType `yaml:"type" json:"t,omitempty"`
}

func (dc *DedicatedColumn) MarshalJSON() ([]byte, error) {
	type dcAlias DedicatedColumn // alias required to avoid recursive calls of MarshalJSON

	cpy := (dcAlias)(*dc)
	if cpy.Scope == DefaultDedicatedColumnScope {
		cpy.Scope = ""
	}
	if cpy.Type == DefaultDedicatedColumnType {
		cpy.Type = ""
	}
	return json.Marshal(&cpy)
}

func (dc *DedicatedColumn) UnmarshalJSON(b []byte) error {
	type dcAlias DedicatedColumn // alias required to avoid recursive calls of UnmarshalJSON

	err := json.Unmarshal(b, (*dcAlias)(dc))
	if err != nil {
		return err
	}
	if dc.Scope == "" {
		dc.Scope = DefaultDedicatedColumnScope
	}
	if dc.Type == "" {
		dc.Type = DefaultDedicatedColumnType
	}
	return nil
}

// DedicatedColumns represents a set of configured dedicated columns.
type DedicatedColumns []DedicatedColumn

func NewBlockMeta(tenantID string, blockID uuid.UUID, version string, encoding Encoding, dataEncoding string) *BlockMeta {
	return NewBlockMetaWithDedicatedColumns(tenantID, blockID, version, encoding, dataEncoding, nil)
}

func NewBlockMetaWithDedicatedColumns(tenantID string, blockID uuid.UUID, version string, encoding Encoding, dataEncoding string, dc DedicatedColumns) *BlockMeta {
	b := &BlockMeta{
		Version:          version,
		BlockID:          blockID,
		MinID:            []byte{},
		MaxID:            []byte{},
		TenantID:         tenantID,
		Encoding:         encoding,
		DataEncoding:     dataEncoding,
		DedicatedColumns: dc,
	}

	return b
}

// ObjectAdded updates the block meta appropriately based on information about an added record
// start/end are unix epoch seconds
func (b *BlockMeta) ObjectAdded(id []byte, start, end uint32) {
	if start > 0 {
		startTime := time.Unix(int64(start), 0)
		if b.StartTime.IsZero() || startTime.Before(b.StartTime) {
			b.StartTime = startTime
		}
	}

	if end > 0 {
		endTime := time.Unix(int64(end), 0)
		if b.EndTime.IsZero() || endTime.After(b.EndTime) {
			b.EndTime = endTime
		}
	}

	if len(b.MinID) == 0 || bytes.Compare(id, b.MinID) == -1 {
		b.MinID = id
	}
	if len(b.MaxID) == 0 || bytes.Compare(id, b.MaxID) == 1 {
		b.MaxID = id
	}

	b.TotalObjects++
}

func (b *BlockMeta) DedicatedColumnsHash() uint64 {
	return b.DedicatedColumns.Hash()
}

func DedicatedColumnsFromTracestorepb(tracestorepbCols []*tracestorepb.DedicatedColumn) (DedicatedColumns, error) {
	cols := make(DedicatedColumns, 0, len(tracestorepbCols))

	for _, c := range tracestorepbCols {
		scope, err := DedicatedColumnScopeFromTracestorepb(c.Scope)
		if err != nil {
			return nil, fmt.Errorf("unable to convert dedicated column '%s': %w", c.Name, err)
		}

		typ, err := DedicatedColumnTypeFromTracestorepb(c.Type)
		if err != nil {
			return nil, fmt.Errorf("unable to convert dedicated column '%s': %w", c.Name, err)
		}

		cols = append(cols, DedicatedColumn{
			Scope: scope,
			Name:  c.Name,
			Type:  typ,
		})
	}

	return cols, nil
}

func (dcs DedicatedColumns) ToTracestorepb() ([]*tracestorepb.DedicatedColumn, error) {
	tracestorepbCols := make([]*tracestorepb.DedicatedColumn, 0, len(dcs))

	for _, c := range dcs {
		scope, err := c.Scope.ToTracestorepb()
		if err != nil {
			return nil, fmt.Errorf("unable to convert dedicated column '%s': %w", c.Name, err)
		}

		typ, err := c.Type.ToTracestorepb()
		if err != nil {
			return nil, fmt.Errorf("unable to convert dedicated column '%s': %w", c.Name, err)
		}

		tracestorepbCols = append(tracestorepbCols, &tracestorepb.DedicatedColumn{
			Scope: scope,
			Name:  c.Name,
			Type:  typ,
		})
	}

	return tracestorepbCols, nil
}

func (dcs DedicatedColumns) Validate() error {
	var countSpan, countRes int
	for _, dc := range dcs {
		err := dc.Validate()
		if err != nil {
			return err
		}
		switch dc.Scope {
		case DedicatedColumnScopeSpan:
			countSpan++
		case DedicatedColumnScopeResource:
			countRes++
		}
	}
	if countSpan > maxSupportedSpanColumns {
		return fmt.Errorf("number of dedicated columns with scope 'span' must be <= %d but was %d", maxSupportedSpanColumns, countSpan)
	}
	if countRes > maxSupportedResourceColumns {
		return fmt.Errorf("number of dedicated columns with scope 'resource' must be <= %d but was %d", maxSupportedResourceColumns, countRes)
	}
	return nil
}

func (dc *DedicatedColumn) Validate() error {
	if dc.Name == "" {
		return errors.New("dedicated column invalid: name must not be empty")
	}
	_, err := dc.Type.ToTracestorepb()
	if err != nil {
		return fmt.Errorf("dedicated column '%s' invalid: %w", dc.Name, err)
	}
	_, err = dc.Scope.ToTracestorepb()
	if err != nil {
		return fmt.Errorf("dedicated column '%s' invalid: %w", dc.Name, err)
	}
	return nil
}

// separatorByte is a byte that cannot occur in valid UTF-8 sequences
var separatorByte = []byte{255}

// Hash hashes the given dedicated columns configuration
func (dcs DedicatedColumns) Hash() uint64 {
	if len(dcs) == 0 {
		return 0
	}
	h := xxhash.New()
	for _, c := range dcs {
		_, _ = h.WriteString(string(c.Scope))
		_, _ = h.Write(separatorByte)
		_, _ = h.WriteString(c.Name)
		_, _ = h.Write(separatorByte)
		_, _ = h.WriteString(string(c.Type))
	}
	return h.Sum64()
}
