// Copyright 2020 DigitalOcean
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//		http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
