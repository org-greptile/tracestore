package ingester

import (
	"context"
	"errors"
	"runtime/debug"

	"github.com/go-kit/log/level"
	"example.com/acme/kit/user"
	"example.com/acme/tracestore/pkg/tracestorepb"
	"example.com/acme/tracestore/pkg/util/log"
)

func (i *Ingester) SearchRecent(ctx context.Context, req *tracestorepb.SearchRequest) (res *tracestorepb.SearchResponse, err error) {
	defer func() {
		if r := recover(); r != nil {
			level.Error(log.Logger).Log("msg", "recover in SearchRecent", "query", req.Query, "stack", r, string(debug.Stack()))
			err = errors.New("recovered in SearchRecent")
		}
	}()

	instanceID, err := user.ExtractOrgID(ctx)
	if err != nil {
		return nil, err
	}
	inst, ok := i.getInstanceByID(instanceID)
	if !ok || inst == nil {
		return &tracestorepb.SearchResponse{}, nil
	}

	res, err = inst.Search(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (i *Ingester) SearchTags(ctx context.Context, req *tracestorepb.SearchTagsRequest) (res *tracestorepb.SearchTagsResponse, err error) {
	defer func() {
		if r := recover(); r != nil {
			level.Error(log.Logger).Log("msg", "recover in SearchTags", "stack", r, string(debug.Stack()))
			err = errors.New("recovered in SearchTags")
		}
	}()

	instanceID, err := user.ExtractOrgID(ctx)
	if err != nil {
		return nil, err
	}
	inst, ok := i.getInstanceByID(instanceID)
	if !ok || inst == nil {
		return &tracestorepb.SearchTagsResponse{}, nil
	}

	res, err = inst.SearchTags(ctx, req.Scope)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (i *Ingester) SearchTagsV2(ctx context.Context, req *tracestorepb.SearchTagsRequest) (res *tracestorepb.SearchTagsV2Response, err error) {
	defer func() {
		if r := recover(); r != nil {
			level.Error(log.Logger).Log("msg", "recover in SearchTagsV2", "stack", r, string(debug.Stack()))
			err = errors.New("recovered in SearchTagsV2")
		}
	}()

	instanceID, err := user.ExtractOrgID(ctx)
	if err != nil {
		return nil, err
	}
	inst, ok := i.getInstanceByID(instanceID)
	if !ok || inst == nil {
		return &tracestorepb.SearchTagsV2Response{}, nil
	}

	res, err = inst.SearchTagsV2(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (i *Ingester) SearchTagValues(ctx context.Context, req *tracestorepb.SearchTagValuesRequest) (res *tracestorepb.SearchTagValuesResponse, err error) {
	defer func() {
		if r := recover(); r != nil {
			level.Error(log.Logger).Log("msg", "recover in SearchTagValues", "stack", r, string(debug.Stack()))
			err = errors.New("recovered in SearchTagValues")
		}
	}()

	instanceID, err := user.ExtractOrgID(ctx)
	if err != nil {
		return nil, err
	}
	inst, ok := i.getInstanceByID(instanceID)
	if !ok || inst == nil {
		return &tracestorepb.SearchTagValuesResponse{}, nil
	}

	res, err = inst.SearchTagValues(ctx, req.TagName, req.MaxTagValues, req.StaleValueThreshold)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (i *Ingester) SearchTagValuesV2(ctx context.Context, req *tracestorepb.SearchTagValuesRequest) (res *tracestorepb.SearchTagValuesV2Response, err error) {
	defer func() {
		if r := recover(); r != nil {
			level.Error(log.Logger).Log("msg", "recover in SearchTagValuesV2", "tag", req.TagName, "query", req.Query, "stack", r, string(debug.Stack()))
			err = errors.New("recovered in SearchTagValuesV2")
		}
	}()

	instanceID, err := user.ExtractOrgID(ctx)
	if err != nil {
		return nil, err
	}
	inst, ok := i.getInstanceByID(instanceID)
	if !ok || inst == nil {
		return &tracestorepb.SearchTagValuesV2Response{}, nil
	}

	res, err = inst.SearchTagValuesV2(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// SearchBlock only exists here to fulfill the protobuf interface. The ingester will never support
// backend search
func (i *Ingester) SearchBlock(context.Context, *tracestorepb.SearchBlockRequest) (*tracestorepb.SearchResponse, error) {
	return nil, errors.New("not implemented")
}
