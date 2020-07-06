package main

import (
	c "context"
	"flag"
	"github.com/digitalocean/godo"
	"github.com/google/subcommands"
	"time"
)

type upCmd struct {
	distro string
	region string
	slug string
}

func (*upCmd) Name() string     { return "up" }
func (*upCmd) Synopsis() string { return "Allows you to start up a new disposable droplet." }
func (*upCmd) Usage() string {
	return `up:
  Allows you to start up a new disposable droplet.
`
}

func (p *upCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.distro, "distro", "", "Sets the distro slug. Will default to the newest Debian release.")
	f.StringVar(&p.region, "region", "", "Sets the region. Will default to the default region within the config.")
	f.StringVar(&p.slug, "size", "", "Sets the size slug of the droplet you want. Will default to the default size slug within the config.")
}

func getLatestDebian(distros []godo.Image) string {
	debian := make([]godo.Image, 0, 1)
	for _, v := range distros {
		if v.Distribution == "Debian" {
			debian = append(debian, v)
		}
	}
	if len(debian) == 0 {
		panic("can't find latest debian x64")
	}
	var latest godo.Image
	var latestTime time.Time
	for _, v := range debian {
		t, err := time.ParseInLocation("2006-01-02T15:04:05Z", v.Created, time.UTC)
		if err != nil {
			panic(err)
		}
		if latestTime.Before(t) {
			latest = v
			latestTime = t
		}
	}
	return latest.Slug
}

func (p *upCmd) Execute(_ c.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	clientInit()
	if p.region == "" {
		p.region = config.DefaultRegion
	}
	if p.slug == "" {
		p.slug = config.DefaultSize
	}
	if p.distro == "" {
		distros, resp, err := client.Images.ListDistribution(context(), &godo.ListOptions{})
		if err != nil {
			if resp != nil && resp.StatusCode == 401 {
				println("Please run do-disposable auth to reset your token.")
				return subcommands.ExitFailure
			}
			panic(err)
		}
		p.distro = getLatestDebian(distros)
	}
	handleDisposableDroplet(p.region, p.slug, p.distro)
	return subcommands.ExitSuccess
}
