package main

import (
	c "context"
	"flag"
	"github.com/google/subcommands"
	"os"
)

func main() {
	subcommands.Register(&authCmd{}, "")
	subcommands.Register(&setRegionCmd{}, "")
	subcommands.Register(&setSizeCmd{}, "")
	subcommands.Register(&upCmd{}, "")

	flag.Parse()
	ctx := c.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
