package main

import (
	"example.com/acme/tracestore/pkg/util"
)

type querySearchTagsCmd struct {
	APIEndpoint string `arg:"" help:"tracestore api endpoint"`

	OrgID string `help:"optional orgID"`
}

func (cmd *querySearchTagsCmd) Run(_ *globalOptions) error {
	client := util.NewClient(cmd.APIEndpoint, cmd.OrgID)

	tags, err := client.SearchTags()
	if err != nil {
		return err
	}

	return printAsJSON(tags)
}
