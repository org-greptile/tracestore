package ingester

import (
	"context"

	"example.com/acme/tracestore/pkg/tracestorepb"
	"github.com/pkg/errors"
	"github.com/weaveworks/common/user"
)

func (i *Ingester) SearchRecent(ctx context.Context, req *tracestorepb.SearchRequest) (*tracestorepb.SearchResponse, error) {
	instanceID, err := user.ExtractOrgID(ctx)
	if err != nil {
		return nil, err
	}
	inst, ok := i.getInstanceByID(instanceID)
	if !ok || inst == nil {
		return &tracestorepb.SearchResponse{}, nil
	}

	res, err := inst.Search(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (i *Ingester) SearchTags(ctx context.Context, req *tracestorepb.SearchTagsRequest) (*tracestorepb.SearchTagsResponse, error) {
	instanceID, err := user.ExtractOrgID(ctx)
	if err != nil {
		return nil, err
	}
	inst, ok := i.getInstanceByID(instanceID)
	if !ok || inst == nil {
		return &tracestorepb.SearchTagsResponse{}, nil
	}

	res, err := inst.SearchTags(ctx, req.Scope)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (i *Ingester) SearchTagsV2(ctx context.Context, req *tracestorepb.SearchTagsRequest) (*tracestorepb.SearchTagsV2Response, error) {
	instanceID, err := user.ExtractOrgID(ctx)
	if err != nil {
		return nil, err
	}
	inst, ok := i.getInstanceByID(instanceID)
	if !ok || inst == nil {
		return &tracestorepb.SearchTagsV2Response{}, nil
	}

	res, err := inst.SearchTagsV2(ctx, req.Scope)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (i *Ingester) SearchTagValues(ctx context.Context, req *tracestorepb.SearchTagValuesRequest) (*tracestorepb.SearchTagValuesResponse, error) {
	instanceID, err := user.ExtractOrgID(ctx)
	if err != nil {
		return nil, err
	}
	inst, ok := i.getInstanceByID(instanceID)
	if !ok || inst == nil {
		return &tracestorepb.SearchTagValuesResponse{}, nil
	}

	res, err := inst.SearchTagValues(ctx, req.TagName)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (i *Ingester) SearchTagValuesV2(ctx context.Context, req *tracestorepb.SearchTagValuesRequest) (*tracestorepb.SearchTagValuesV2Response, error) {
	instanceID, err := user.ExtractOrgID(ctx)
	if err != nil {
		return nil, err
	}
	inst, ok := i.getInstanceByID(instanceID)
	if !ok || inst == nil {
		return &tracestorepb.SearchTagValuesV2Response{}, nil
	}

	res, err := inst.SearchTagValuesV2(ctx, req)
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
