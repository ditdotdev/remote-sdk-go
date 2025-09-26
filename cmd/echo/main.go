// Package main provides the echo remote server executable.
package main

import (
	"github.com/datadatdat/remote-sdk-go/internal/echo"
	"github.com/datadatdat/remote-sdk-go/remote"
)

type MockRemote struct {
}

func main() {
	remote.Register(echo.Remote{})
	remote.Serve("echo")
}
