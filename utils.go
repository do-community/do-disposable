package main

import (
	"bufio"
	"os"
	"strings"
)

// GetInput is used to get the input which a user types.
func GetInput(query string) string {
	print(query)
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	text = strings.Replace(text, "\n", "", -1)
	text = strings.Replace(text, "\r", "", -1)
	return text
}
