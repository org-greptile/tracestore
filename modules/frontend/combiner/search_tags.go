package combiner

import (
	"example.com/acme/tracestore/pkg/tracestorepb"
	"example.com/acme/tracestore/pkg/util"
)

var (
	_ GRPCCombiner[*tracestorepb.SearchTagsResponse]   = (*genericCombiner[*tracestorepb.SearchTagsResponse])(nil)
	_ GRPCCombiner[*tracestorepb.SearchTagsV2Response] = (*genericCombiner[*tracestorepb.SearchTagsV2Response])(nil)
)

func NewSearchTags(limitBytes int) Combiner {
	d := util.NewDistinctStringCollector(limitBytes)

	return &genericCombiner[*tracestorepb.SearchTagsResponse]{
		httpStatusCode: 200,
		new:            func() *tracestorepb.SearchTagsResponse { return &tracestorepb.SearchTagsResponse{} },
		current:        &tracestorepb.SearchTagsResponse{TagNames: make([]string, 0)},
		combine: func(partial, final *tracestorepb.SearchTagsResponse, _ PipelineResponse) error {
			for _, v := range partial.TagNames {
				d.Collect(v)
			}
			return nil
		},
		finalize: func(response *tracestorepb.SearchTagsResponse) (*tracestorepb.SearchTagsResponse, error) {
			response.TagNames = d.Strings()
			return response, nil
		},
		quit: func(_ *tracestorepb.SearchTagsResponse) bool {
			return d.Exceeded()
		},
		diff: func(response *tracestorepb.SearchTagsResponse) (*tracestorepb.SearchTagsResponse, error) {
			response.TagNames = d.Diff()
			return response, nil
		},
	}
}

func NewTypedSearchTags(limitBytes int) GRPCCombiner[*tracestorepb.SearchTagsResponse] {
	return NewSearchTags(limitBytes).(GRPCCombiner[*tracestorepb.SearchTagsResponse])
}

func NewSearchTagsV2(limitBytes int) Combiner {
	// Distinct collector map to collect scopes and scope values
	distinctValues := map[string]*util.DistinctStringCollector{}

	return &genericCombiner[*tracestorepb.SearchTagsV2Response]{
		httpStatusCode: 200,
		new:            func() *tracestorepb.SearchTagsV2Response { return &tracestorepb.SearchTagsV2Response{} },
		current:        &tracestorepb.SearchTagsV2Response{Scopes: make([]*tracestorepb.SearchTagsV2Scope, 0)},
		combine: func(partial, final *tracestorepb.SearchTagsV2Response, _ PipelineResponse) error {
			for _, res := range partial.GetScopes() {
				dvc := distinctValues[res.Name]
				if dvc == nil {
					dvc = util.NewDistinctStringCollector(limitBytes)
					distinctValues[res.Name] = dvc
				}
				for _, tag := range res.Tags {
					dvc.Collect(tag)
				}
			}
			return nil
		},
		finalize: func(final *tracestorepb.SearchTagsV2Response) (*tracestorepb.SearchTagsV2Response, error) {
			final.Scopes = make([]*tracestorepb.SearchTagsV2Scope, 0, len(distinctValues))

			for scope, dvc := range distinctValues {
				final.Scopes = append(final.Scopes, &tracestorepb.SearchTagsV2Scope{
					Name: scope,
					Tags: dvc.Strings(),
				})
			}
			return final, nil
		},
		quit: func(_ *tracestorepb.SearchTagsV2Response) bool {
			for _, dvc := range distinctValues {
				if dvc.Exceeded() {
					return true
				}
			}
			return false
		},
		diff: func(response *tracestorepb.SearchTagsV2Response) (*tracestorepb.SearchTagsV2Response, error) {
			response.Scopes = make([]*tracestorepb.SearchTagsV2Scope, 0, len(distinctValues))

			for scope, dvc := range distinctValues {
				diff := dvc.Diff()
				if len(diff) == 0 {
					continue
				}

				response.Scopes = append(response.Scopes, &tracestorepb.SearchTagsV2Scope{
					Name: scope,
					Tags: diff,
				})
			}

			return response, nil
		},
	}
}

func NewTypedSearchTagsV2(limitBytes int) GRPCCombiner[*tracestorepb.SearchTagsV2Response] {
	return NewSearchTagsV2(limitBytes).(GRPCCombiner[*tracestorepb.SearchTagsV2Response])
}
