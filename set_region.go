package main

import (
	c "context"
	"flag"
	"github.com/google/subcommands"
)

type setRegionCmd struct {}

func (*setRegionCmd) Name() string     { return "setregion" }
func (*setRegionCmd) Synopsis() string { return "Allows you to modify the region. Note that you need to go through the setup with do-disposable auth first (that will also configure this for the first time)." }
func (*setRegionCmd) Usage() string {
	return `setregion:
  Allows you to modify the region. Note that you need to go through the setup with do-disposable auth first (that will also configure this for the first time).
`
}

func (p *setRegionCmd) SetFlags(_ *flag.FlagSet) {}

func (p *setRegionCmd) Execute(_ c.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	clientInit()
	setDefaultRegion()
	return subcommands.ExitSuccess
}
