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
	"bufio"
	"os"
	"strconv"
	"strings"
)

// FormatList is used to format a list to be shown in the console and get the result. The default pointer will result to that if a blank string is entered by the user.
func FormatList(Question string, Items []string, Default *string) string {
	Output := Question + "\n"
	for i, v := range Items {
		Output += strconv.Itoa(i) + ") " + v + "\n"
	}
	Request := "Please enter the index of the item which you want"
	if Default != nil {
		Request += " (if your response is blank, this will default to " + *Default + ")"
	}
	Request += ": "
	print(Output)
	for {
		print(Request)
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		if text == "" {
			if Default == nil {
				continue
			} else {
				return *Default
			}
		}
		i, err := strconv.ParseUint(text, 10, 64)
		if err != nil {
			continue
		}
		if i >= uint64(len(Items)) {
			continue
		}
		return Items[i]
	}
}
