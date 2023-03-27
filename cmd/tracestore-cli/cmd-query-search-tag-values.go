package main

import (
	"example.com/acme/tracestore/pkg/util"
)

type querySearchTagValuesCmd struct {
	APIEndpoint string `arg:"" help:"tracestore api endpoint"`
	Tag         string `arg:"" help:"tag name"`

	OrgID string `help:"optional orgID"`
}

func (cmd *querySearchTagValuesCmd) Run(_ *globalOptions) error {
	client := util.NewClient(cmd.APIEndpoint, cmd.OrgID)

	tagValues, err := client.SearchTagValues(cmd.Tag)
	if err != nil {
		return err
	}

	return printAsJSON(tagValues)
}
