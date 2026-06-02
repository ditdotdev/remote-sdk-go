# Project Development

For general information about contributing changes, see the
[Contributor Guidelines](https://github.com/ditdotdev/.github/blob/master/CONTRIBUTING.md).

## How it Works

The Remote SDK currently provides interfaces only for use by the client. This includes the ability to register
remote providers, and parse URIs. Future work will add server-side capabilities for use in Dit-server, enabling
the EOL of the legacy kotlin remote providers.

Remotes can be directly imported into go programs, or loaded dynamically using
Hashicorp's [go-plugin](https://github.com/hashicorp/go-plugin). The SDK wraps all of this implementation, so that
client can just call `Remote.Get()` (for directly imported remotes) or `Remote.Load()` (for dynamically loaded
remotes). To be used as a plugin, remotes must have a command with a `main` function that invokes
`Remote.Serve()`.

## Building

Run `go build -v ./...`.

If you need to change the gRPC protobuf RPC specification, you will need to rebuild the generated `remote.pb.go`
file. To do this, you will first need to:

- Install the protobuf compiler (`brew install protobuf` on MacOS)
- Install the go protobuf generator (`go get -u github.com/golang/protobuf/protoc-gen-go` and ensure `$GOPATH/bin` is 
  in your path).
  
Once those prerequisites are complete, run `protoc --go_out=plugins=grpc:. remote.proto` from the `internal/proto`
directory.

## Testing

Prior to running tests, you will need to build the `echo` plugin in the `remote` directory, which can be done
via: `go build -o remote/echo ./cmd/echo`.

To run all tests, run `go test -v ./...`.

## Releasing

Push a tag of the form `v<X>.<Y>.<Z>`, and publish the draft release in GitHub.
