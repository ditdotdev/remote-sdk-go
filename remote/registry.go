/*
 * Copyright Datadatdat.
 */

package remote

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

// Registry holds a set of registered Remote implementations and the plugin subprocesses they have been loaded into.
// All methods are safe for concurrent use.
//
// Most callers should use the package-level free functions (Register, Get, Load, Unload, Loaded), which delegate to
// the Default registry. The Registry type is exported for callers that need multiple isolated registries — for
// example, multi-tenant test setups that want to avoid sharing state across parallel sub-tests.
type Registry struct {
	mu         sync.RWMutex
	registered map[string]Remote
	loaded     map[string]loadedRemote
}

// New returns a fresh empty Registry. Use this when you need a registry isolated from Default — most production
// callers should not need to construct their own.
func New() *Registry {
	return &Registry{
		registered: map[string]Remote{},
		loaded:     map[string]loadedRemote{},
	}
}

// Default is the package-level Registry that the free functions (Register, Get, Load, ...) operate on. Existing
// callers that rely on Register-from-init() get the same behavior they always have.
var Default = New()

// Register adds a Remote to the registry. Typically called from a remote implementation's init() function.
//
// Panics if rem.Type() returns an error. This is intentional: Register is an init-time call, and a remote that
// cannot report its own type is a programmer error that should fail loudly at process start rather than be surfaced
// later as a missing-registration at request time.
func (r *Registry) Register(rem Remote) {
	remoteType, err := rem.Type()
	if err != nil {
		panic(err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.registered[remoteType] = rem
}

// Get returns the registered Remote for the given type. The boolean is false if no Remote has been registered for
// that type — callers should check it and avoid dereferencing the returned Remote when ok is false.
func (r *Registry) Get(remoteType string) (Remote, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rem, ok := r.registered[remoteType]
	return rem, ok
}

// snapshot returns a slice of every currently-registered Remote. Used by routing logic that needs to iterate
// without holding the lock for the entire iteration (which could deadlock if a callback re-enters the registry).
func (r *Registry) snapshot() []Remote {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Remote, 0, len(r.registered))
	for _, rem := range r.registered {
		out = append(out, rem)
	}
	return out
}

// Load starts (or returns the cached handle for) the plugin subprocess for the given remote type.
//
// The cache check + plugin spawn is performed under the registry lock, so two concurrent goroutines calling
// Load(...) for the same uncached type will share a single subprocess: the second goroutine will see the cache
// entry the first one wrote and return that. Without this serialization, both goroutines would spawn their own
// subprocess and the loser would leak — see issue #46 finding #1.
//
// Load does not require remoteType to be registered locally: the plugin subprocess registers itself when it
// starts. Local registration is only needed if the same process is going to call Serve(remoteType) as well.
func (r *Registry) Load(remoteType string, pluginPath string) (Remote, error) {
	// Fast path: check cache under a read lock.
	r.mu.RLock()
	if v, ok := r.loaded[remoteType]; ok {
		r.mu.RUnlock()
		return v.r, nil
	}
	r.mu.RUnlock()

	// Slow path: take the write lock, re-check the cache (another goroutine may have populated it while we
	// were upgrading), then spawn the plugin subprocess. Holding the lock across NewClient + Dispense is what
	// prevents the leaked-subprocess race.
	r.mu.Lock()
	defer r.mu.Unlock()

	if v, ok := r.loaded[remoteType]; ok {
		return v.r, nil
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   pluginName,
		Output: os.Stdout,
		Level:  hclog.Error,
	})

	// The local Impl in pluginMap is only used by the server side (Serve). On the client side that
	// Load() implements, only GRPCClient is invoked, so a nil Impl is fine if remoteType was never
	// registered locally — the subprocess provides its own implementation.
	pluginMap := map[string]plugin.Plugin{
		pluginName: &remotePlugin{Impl: r.registered[remoteType]},
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

	rem, ok := raw.(Remote)
	if !ok {
		client.Kill()
		return nil, fmt.Errorf("plugin %q did not return a Remote", remoteType)
	}

	r.loaded[remoteType] = loadedRemote{r: rem, c: client}
	return rem, nil
}

// Unload terminates the plugin subprocess for the given type and removes the cache entry. Silently succeeds if
// the type was never loaded; use Loaded() to distinguish.
func (r *Registry) Unload(remoteType string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if val, ok := r.loaded[remoteType]; ok {
		val.c.Kill()
		delete(r.loaded, remoteType)
	}
}

// Loaded reports whether the given remote type currently has a live plugin subprocess in the cache.
func (r *Registry) Loaded(remoteType string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.loaded[remoteType]
	return ok
}

// Clear removes every registration and terminates every loaded plugin subprocess. Intended for test isolation —
// production code should not need to call it.
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.registered = map[string]Remote{}
	for _, v := range r.loaded {
		v.c.Kill()
	}
	r.loaded = map[string]loadedRemote{}
}
