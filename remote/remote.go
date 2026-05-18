/*
 * Copyright Datadatdat.
 */

// Package remote provides the core remote plugin infrastructure for Datadatdat.
package remote

import (
	"context"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	remote "github.com/datadatdat/remote-sdk-go/internal/proto"
)

/*
 * SDK for Datadatdat remotes.
 */

// Tag represents a filter criteria for listing commits, with a key-value pair for matching.
type Tag struct {
	Key   string
	Value *string
}

// Commit represents a data snapshot with an identifier and associated properties.
type Commit struct {
	ID         string
	Properties map[string]interface{}
}

// Remote defines the interface for remote storage backends that can store and retrieve commits.
type Remote interface {

	/*
	 * Returns the canonical name of this provider, such as "ssh" or "s3". This must be globally unique, and must
	 * match the leading URI component (ssh://...).
	 */
	Type() (string, error)

	/*
	 * Parse a URL and return the provider-specific remote parameters in structured form. These properties will be
	 * preserved as part of the remote metadata on the server and passed to subsequent server-side operations. The
	 * additional properties map can contain properties specified by the user that don't fit the URI format well,
	 * such as "-p keyFile=/path/to/sshKey". This should return an error for a bad URL format or invalid properties.
	 * The calling context will have stripped out any query parameters or fragments.
	 */
	FromURL(url string, properties map[string]string) (map[string]interface{}, error)

	/*
	 * Convert a remote back into URI form for display. Since this is for display only, any sensitive information
	 * should be redacted (e.g. "user:****@host"). Any properties that cannot be represented in the URI can be
	 * passed back as the second part of the pair.
	 */
	ToURL(properties map[string]interface{}) (string, map[string]string, error)

	/*
	 * Given a set of remote properties, return a set of parameter properties that will be passed to each operation.
	 * This is invoked in the context of the user CLI. It can access user data, such as SSH or AWS configuration. It
	 * can also interactively prompt the user for additional input (such as a password).
	 */
	GetParameters(properties map[string]interface{}) (map[string]interface{}, error)

	/*
	 * Validates the configuration of a remote, invoked by the server whenever a remote is passed as input or read
	 * from the metadata store. This ensures that no malformed remotes are ever present.
	 */
	ValidateRemote(properties map[string]interface{}) error

	/*
	 * Validates the configuration of remote parameters.
	 */
	ValidateParameters(parameters map[string]interface{}) error

	/*
	 * Fetches a set of commits from the remote server. Commits are simply a tuple of (commitId, properties), with
	 * some properties having semantic significance (namely timestamp and tags). The remote provider should always
	 * return commits in reverse timestamp order, optionally filtered by the given tags. There are utility methods
	 * in RemoteServerUtil if remotes don't provide this functionality server-side. Tags are specified as a list of
	 * pairs, where the first element is always the key and the second is optionally the value.
	 *
	 * There is not yet support for pagination, though that will be added in the future to avoid having to fetch
	 * the entire commit history every time.
	 */
	ListCommits(properties map[string]interface{}, parameters map[string]interface{}, tags []Tag) ([]Commit, error)

	/**
	 * Fetches a single commit from the given remote. Returns nil if no such commit exists.
	 */
	GetCommit(properties map[string]interface{}, parameters map[string]interface{}, commitID string) (*Commit, error)
}

type remotePlugin struct {
	plugin.NetRPCUnsupportedPlugin
	Impl Remote
}

func (p *remotePlugin) GRPCServer(_ *plugin.GRPCBroker, s *grpc.Server) error {
	remote.RegisterRemoteServer(s, &remoteRPCServer{Impl: p.Impl})
	return nil
}

func (remotePlugin) GRPCClient(_ context.Context, _ *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &remoteRPCClient{Client: remote.NewRemoteClient(c)}, nil
}

type loadedRemote struct {
	r Remote
	c *plugin.Client
}

var registeredRemotes = map[string]Remote{}
var loadedRemotes = map[string]loadedRemote{}

// Register a new remote. This should be called from the init() function of a remote implementation. The remotes can
// later be accessed via the Get() method.
//
// Panics if remote.Type() returns an error. This is intentional: Register is an init-time call, and a remote that
// cannot report its own type is a programmer error that should fail loudly at process start rather than be surfaced
// later as a missing-registration at request time.
func Register(remote Remote) {
	remoteType, err := remote.Type()
	if err != nil {
		panic(err)
	}

	registeredRemotes[remoteType] = remote
}

// Get a remote by type.
func Get(remoteType string) Remote {
	return registeredRemotes[remoteType]
}

// Clear any registered or loaded remotes. Should only be used for testing.
func Clear() {
	registeredRemotes = map[string]Remote{}

	for _, v := range loadedRemotes {
		v.c.Kill()
	}

	loadedRemotes = map[string]loadedRemote{}
}

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "Datadatdat",
	MagicCookieValue: "dba4fe2b-56ff-4a16-9bfc-bf651b8f12d6",
}

// pluginName is the hclog logger name and the plugin-map key used by
// both Serve and Load when wiring up the go-plugin RPC channel; both
// sides have to agree on it for Dispense to find the implementation.
const pluginName = "remote"

// Serve runs the remote as a plugin server, to be invoked from the main method of the remote implementation.
//
// Serve has no error return and does not return under normal operation: plugin.Serve blocks until the parent
// process closes stdin (per the go-plugin contract) and then exits the process. There is intentionally nothing
// for a caller to do with control flow after Serve returns.
func Serve(remoteType string) {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   pluginName,
		Output: os.Stdout,
		Level:  hclog.Error,
	})

	remote := Get(remoteType)

	var pluginMap = map[string]plugin.Plugin{
		pluginName: &remotePlugin{Impl: remote},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
		GRPCServer:      plugin.DefaultGRPCServer,
		Logger:          logger,
	})
}

const windowsGOOS = "windows"

// pluginBinaryPath joins pluginPath and remoteType into the path of the plugin executable, adding the OS-appropriate
// executable suffix. On Windows, Go's toolchain emits binaries with a .exe suffix, and exec.Command does not auto-
// append .exe when the path contains a separator — so the SDK has to add it explicitly or Load() can never spawn its
// plugin subprocess.
func pluginBinaryPath(pluginPath string, remoteType string) string {
	binName := remoteType
	if runtime.GOOS == windowsGOOS && !strings.HasSuffix(strings.ToLower(binName), ".exe") {
		binName += ".exe"
	}
	return filepath.Join(pluginPath, binName)
}

// Load a remote via the plugin interface. These plugins will remain loaded until Unload() or Clear() is called.
func Load(remoteType string, pluginPath string) (Remote, error) {
	if v, ok := loadedRemotes[remoteType]; ok {
		return v.r, nil
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   pluginName,
		Output: os.Stdout,
		Level:  hclog.Error,
	})

	remote := Get(remoteType)

	var pluginMap = map[string]plugin.Plugin{
		pluginName: &remotePlugin{Impl: remote},
	}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  handshakeConfig,
		Plugins:          pluginMap,
		Cmd:              exec.Command(pluginBinaryPath(pluginPath, remoteType)), // #nosec G204 -- pluginPath and remoteType are controlled inputs
		Logger:           logger,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	})

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, err
	}

	raw, err := rpcClient.Dispense(pluginName)
	if err != nil {
		client.Kill()
		return nil, err
	}

	loadedRemotes[remoteType] = loadedRemote{
		r: raw.(Remote),
		c: client,
	}

	return raw.(Remote), nil
}

// Unload terminates and removes a loaded remote plugin from memory.
func Unload(remoteType string) {
	if val, ok := loadedRemotes[remoteType]; ok {
		val.c.Kill()
		delete(loadedRemotes, remoteType)
	}
}
