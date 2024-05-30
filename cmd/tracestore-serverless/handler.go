package serverless

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"sync"

	"github.com/google/uuid"
	"example.com/acme/kit/flagext"
	"example.com/acme/kit/user"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	"example.com/acme/tracestore/pkg/api"
	"example.com/acme/tracestore/pkg/tracestorepb"
	"example.com/acme/tracestore/pkg/traceql"
	"example.com/acme/tracestore/tracestoredb"
	"example.com/acme/tracestore/tracestoredb/backend"
	"example.com/acme/tracestore/tracestoredb/backend/azure"
	azureconfig "example.com/acme/tracestore/tracestoredb/backend/azure/config"
	"example.com/acme/tracestore/tracestoredb/backend/gcs"
	"example.com/acme/tracestore/tracestoredb/backend/local"
	"example.com/acme/tracestore/tracestoredb/backend/s3"
	"example.com/acme/tracestore/tracestoredb/encoding"
	"example.com/acme/tracestore/tracestoredb/encoding/common"
)

const (
	envConfigPrefix = "TRACESTORE"
)

// used to initialize a reader one time
var (
	reader       backend.Reader
	readerErr    error
	readerConfig *tracestoredb.Config
	readerOnce   sync.Once
)

type HTTPError struct {
	Err    error
	Status int
}

// Handler is the main entrypoint for the serverless handler. it expects a tracestorepb.SearchBlockRequest
// encoded in its parameters
func Handler(r *http.Request) (*tracestorepb.SearchResponse, *HTTPError) {
	searchReq, err := api.ParseSearchBlockRequest(r)
	if err != nil {
		return nil, httpError("parsing search request", err, http.StatusBadRequest)
	}

	maxBytes, err := api.ExtractServerlessParams(r)
	if err != nil {
		return nil, httpError("extracting serverless params", err, http.StatusBadRequest)
	}

	// load config, fields are set through env vars TRACESTORE_
	reader, cfg, err := loadBackend()
	if err != nil {
		return nil, httpError("loading backend", err, http.StatusInternalServerError)
	}

	tenant, _, err := user.ExtractOrgIDFromHTTPRequest(r)
	if err != nil {
		return nil, httpError("extracting org id", err, http.StatusBadRequest)
	}

	blockID, err := uuid.Parse(searchReq.BlockID)
	if err != nil {
		return nil, httpError("parsing uuid", err, http.StatusBadRequest)
	}

	enc, err := backend.ParseEncoding(searchReq.Encoding)
	if err != nil {
		return nil, httpError("parsing encoding", err, http.StatusBadRequest)
	}

	dc, err := backend.DedicatedColumnsFromTracestorepb(searchReq.DedicatedColumns)
	if err != nil {
		return nil, httpError("parsing dedicated columns", err, http.StatusBadRequest)
	}

	// /giphy so meta
	meta := &backend.BlockMeta{
		Version:          searchReq.Version,
		TenantID:         tenant,
		Encoding:         enc,
		IndexPageSize:    searchReq.IndexPageSize,
		TotalRecords:     searchReq.TotalRecords,
		BlockID:          blockID,
		DataEncoding:     searchReq.DataEncoding,
		Size:             searchReq.Size_,
		FooterSize:       searchReq.FooterSize,
		DedicatedColumns: dc,
	}

	block, err := encoding.OpenBlock(meta, reader)
	if err != nil {
		return nil, httpError("creating backend block", err, http.StatusInternalServerError)
	}

	opts := common.SearchOptions{
		StartPage:  int(searchReq.StartPage),
		TotalPages: int(searchReq.PagesToSearch),
		MaxBytes:   maxBytes,
	}
	cfg.Search.ApplyToOptions(&opts)

	var resp *tracestorepb.SearchResponse

	if api.IsTraceQLQuery(searchReq.SearchReq) {
		engine := traceql.NewEngine()

		spansetFetcher := traceql.NewSpansetFetcherWrapper(func(ctx context.Context, req traceql.FetchSpansRequest) (traceql.FetchSpansResponse, error) {
			return block.Fetch(ctx, req, opts)
		})
		resp, err = engine.ExecuteSearch(r.Context(), searchReq.SearchReq, spansetFetcher)
		if err != nil {
			return nil, httpError("searching block", err, http.StatusInternalServerError)
		}
	} else {
		resp, err = block.Search(r.Context(), searchReq.SearchReq, opts)
		if err != nil {
			return nil, httpError("searching block", err, http.StatusInternalServerError)
		}
	}

	runtime.GC()

	return resp, nil
}

func loadBackend() (backend.Reader, *tracestoredb.Config, error) {
	readerOnce.Do(func() {
		cfg, err := loadConfig()
		if err != nil {
			readerErr = err
			return
		}

		var r backend.RawReader

		// Create the backend with NewNoConfirm() to prevent an extra call to the various backends on
		// startup. This extra call exists just to confirm the bucket is accessible and force the
		// standard Tracestore components to fail during startup. If permissions are not correct this Lambda
		// will fail instantly anyway and in a heavy query environment the extra calls will start to add up.
		switch cfg.Backend {
		case backend.Local:
			err = fmt.Errorf("local backend not supported for serverless functions")
		case backend.GCS:
			r, _, _, err = gcs.NewNoConfirm(cfg.GCS)
		case backend.S3:
			r, _, _, err = s3.NewNoConfirm(cfg.S3)
		case backend.Azure:
			r, _, _, err = azure.NewNoConfirm(cfg.Azure)
		default:
			err = fmt.Errorf("unknown backend %s", cfg.Backend)
		}
		if err != nil {
			readerErr = err
			return
		}

		readerConfig = cfg
		reader = backend.NewReader(r)
	})

	return reader, readerConfig, readerErr
}

func loadConfig() (*tracestoredb.Config, error) {
	defaultConfig := &tracestoredb.Config{
		Search: &tracestoredb.SearchConfig{
			ChunkSizeBytes:     tracestoredb.DefaultSearchChunkSizeBytes,
			PrefetchTraceCount: tracestoredb.DefaultPrefetchTraceCount,
		},
		Local: &local.Config{},
		GCS:   &gcs.Config{},
		S3:    &s3.Config{},
		Azure: &azureconfig.Config{},
	}

	// horrible viper dance since it won't unmarshal to a struct from env: https://github.com/spf13/viper/issues/188
	v := viper.NewWithOptions()
	b, err := yaml.Marshal(defaultConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal default config: %w", err)
	}
	v.SetConfigType("yaml")
	if err = v.MergeConfig(bytes.NewReader(b)); err != nil {
		return nil, fmt.Errorf("failed to merge config: %w", err)
	}

	v.AutomaticEnv()
	v.SetEnvPrefix(envConfigPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	cfg := &tracestoredb.Config{}
	err = v.Unmarshal(cfg, setTagName, setDecodeHooks)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}

// this forces mapstructure to use the yaml tag that is already defined
// for all config struct fields
func setTagName(d *mapstructure.DecoderConfig) {
	d.TagName = "yaml"
}

// install all required decodeHooks so that viper will parse the yaml properly
func setDecodeHooks(c *mapstructure.DecoderConfig) {
	c.DecodeHook = mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		stringToFlagExt(),
	)
}

// stringToFlagExt returns a DecodeHookFunc that converts
// strings to kit.FlagExt values.
func stringToFlagExt() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{},
	) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(flagext.Secret{}) {
			return data, nil
		}
		return flagext.SecretWithValue(data.(string)), nil
	}
}

func httpError(action string, err error, status int) *HTTPError {
	return &HTTPError{
		Err:    fmt.Errorf("serverless [%s]: %w", action, err),
		Status: status,
	}
}
