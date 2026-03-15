package main

import (
	"github.com/fredrikaverpil/pocket/pk"
	"github.com/fredrikaverpil/pocket/tasks/github"
	"github.com/fredrikaverpil/pocket/tasks/golang"
)

var Config = &pk.Config{
	Auto: pk.Serial(
		golang.Tasks(),
		pk.WithOptions(
			github.Tasks(),
			pk.WithFlags(
				github.WorkflowFlags{
					Platforms: []github.Platform{github.Ubuntu},
				},
			),
		),
	),
}
