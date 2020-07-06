package main

import (
	c "context"
	"flag"
	"github.com/google/subcommands"
)

type setSizeCmd struct {}

func (*setSizeCmd) Name() string     { return "setsize" }
func (*setSizeCmd) Synopsis() string { return "Allows you to modify the droplet size. Note that you need to go through the setup with do-disposable auth first (that will also configure this for the first time)." }
func (*setSizeCmd) Usage() string {
	return `setsize:
  Allows you to modify the droplet size. Note that you need to go through the setup with do-disposable auth first (that will also configure this for the first time).
`
}

func (p *setSizeCmd) SetFlags(_ *flag.FlagSet) {}

func (p *setSizeCmd) Execute(_ c.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	clientInit()
	setSize(nil)
	return subcommands.ExitSuccess
}
