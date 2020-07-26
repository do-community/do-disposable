// +build ignore
// Written to allow me to test the copyserver.

package main

import (
	"github.com/JakeMakesStuff/do-disposable/copyserver"
	"net"
)

func main() {
	ln, err := net.Listen("tcp", "127.0.0.1:8190")
	if err != nil {
		panic(err)
	}
	err = copyserver.Copyserver(ln)
	if err != nil {
		panic(err)
	}
}
