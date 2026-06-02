// Package main provides the echo remote server executable.
package main

import (
	"github.com/ditdotdev/remote-sdk-go/internal/echo"
	"github.com/ditdotdev/remote-sdk-go/remote"
)

func main() {
	remote.Register(echo.Remote{})
	remote.Serve("echo")
}
