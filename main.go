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
