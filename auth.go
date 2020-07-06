package main

import (
	c "context"
	"flag"
	"github.com/google/subcommands"
)

type authCmd struct {}

func (*authCmd) Name() string     { return "auth" }
func (*authCmd) Synopsis() string { return "Authenticates the user and creates the configuration if this doesn't exist." }
func (*authCmd) Usage() string {
	return `auth:
  Authenticates the user and creates the configuration if this doesn't exist.
`
}

func (p *authCmd) SetFlags(_ *flag.FlagSet) {}

func (p *authCmd) Execute(_ c.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	fp, exists := loadConfig()
	if exists {
		config.Token = setToken()
		writeConfig(fp)
		genSSHKey()
	} else {
		inputSaveConfig(fp)
	}
	return subcommands.ExitSuccess
}
