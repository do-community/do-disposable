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
