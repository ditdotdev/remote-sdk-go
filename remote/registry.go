// Copyright Dit 2026
// SPDX-License-Identifier: BUSL-1.1

package remote

import (
	"container/list"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

// RegistryOption configures a Registry at construction. Pass options to New().
type RegistryOption func(*registryConfig)

type registryConfig struct {
	maxLoaded   int
	onEvict     func(remoteType string)
	onCacheHit  func(remoteType string)
	onCacheMiss func(remoteType string)
}

// WithMaxLoaded caps the number of plugin subprocesses cached in a Registry. Once the cap is reached, the next
// uncached Load() evicts the least-recently-used entry to make room. Pass 0 (the default) to disable LRU and keep
// the original "grow forever" behavior.
//
// Use this on long-running consumers (servers, daemons) that handle many distinct remote types over their lifetime
// — without a cap, every never-Unload'd Load accumulates a subprocess.
func WithMaxLoaded(n int) RegistryOption {
	return func(c *registryConfig) { c.maxLoaded = n }
}

// WithOnEvict registers a callback invoked once per eviction, just after the subprocess is killed and removed from
// cache. The callback runs while the registry lock is held — keep it fast and non-blocking, or use it to enqueue
// work onto a channel.
func WithOnEvict(fn func(remoteType string)) RegistryOption {
	return func(c *registryConfig) { c.onEvict = fn }
}

// WithOnCacheHit registers a callback invoked when Load() returns a cached entry. Useful for hit-rate metrics.
// Runs under the registry lock.
func WithOnCacheHit(fn func(remoteType string)) RegistryOption {
	return func(c *registryConfig) { c.onCacheHit = fn }
}

// WithOnCacheMiss registers a callback invoked when Load() spawns a new subprocess (cold load). Useful for
// miss-rate metrics. Runs under the registry lock.
func WithOnCacheMiss(fn func(remoteType string)) RegistryOption {
	return func(c *registryConfig) { c.onCacheMiss = fn }
}

// loadedRemote holds the state for a cached plugin subprocess and the bookkeeping needed to evict it safely:
//
//   - refCount tracks in-flight method calls. Eviction waits on drainCond until refCount drops to zero, so callers
//     never see "connection refused" mid-RPC.
//   - lruElem points at the entry's slot in the LRU list. Cache hits move the element to the back (most recent);
//     eviction picks the front (least recent).
type loadedRemote struct {
	r          Remote
	c          *plugin.Client
	refCount   int
	drainCond  *sync.Cond
	lruElem    *list.Element
	remoteType string
	handle     *handleRemote // cached wrapper returned by Load; one per entry for pointer-identity stability
}

// Registry holds a set of registered Remote implementations and the plugin subprocesses they have been loaded into.
// All methods are safe for concurrent use.
//
// Most callers should use the package-level free functions (Register, Get, Load, Unload, Loaded), which delegate to
// the Default registry. The Registry type is exported for callers that need multiple isolated registries — for
// example, multi-tenant test setups that want to avoid sharing state across parallel sub-tests, or long-running
// servers that want an LRU cap (see WithMaxLoaded).
type Registry struct {
	mu         sync.Mutex
	registered map[string]Remote
	loaded     map[string]*loadedRemote
	lru        *list.List // values are *loadedRemote; front = LRU, back = MRU
	config     registryConfig
}

// New returns a fresh empty Registry. Pass RegistryOptions to configure LRU and observability callbacks.
// Without options, the returned registry has no eviction (matches the pre-#50 default behavior).
func New(opts ...RegistryOption) *Registry {
	r := &Registry{
		registered: map[string]Remote{},
		loaded:     map[string]*loadedRemote{},
		lru:        list.New(),
	}
	for _, opt := range opts {
		opt(&r.config)
	}
	return r
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
	r.mu.Lock()
	defer r.mu.Unlock()
	rem, ok := r.registered[remoteType]
	return rem, ok
}

// snapshot returns a slice of every currently-registered Remote. Used by routing logic that needs to iterate
// without holding the lock for the entire iteration (which could deadlock if a callback re-enters the registry).
func (r *Registry) snapshot() []Remote {
	r.mu.Lock()
	defer r.mu.Unlock()
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
// If WithMaxLoaded(n) was set and the cache is already at capacity, the least-recently-used entry is evicted to
// make room. Eviction waits for any in-flight method calls on the LRU entry to drain before killing the subprocess
// (see WithMaxLoaded for details).
//
// Load does not require remoteType to be registered locally: the plugin subprocess registers itself when it
// starts. Local registration is only needed if the same process is going to call Serve(remoteType) as well.
func (r *Registry) Load(remoteType string, pluginPath string) (Remote, error) {
	r.mu.Lock()
	if entry, ok := r.loaded[remoteType]; ok {
		r.touch(entry)
		if r.config.onCacheHit != nil {
			r.config.onCacheHit(remoteType)
		}
		r.mu.Unlock()
		return r.newHandle(entry), nil
	}
	if r.config.onCacheMiss != nil {
		r.config.onCacheMiss(remoteType)
	}

	if r.config.maxLoaded > 0 && len(r.loaded) >= r.config.maxLoaded {
		r.evictLRULocked()
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

	// Release the lock during the (potentially slow) subprocess spawn. Concurrent goroutines for the same
	// remoteType may race here; the loser drops its subprocess after re-acquiring the lock below.
	r.mu.Unlock()

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

	r.mu.Lock()
	defer r.mu.Unlock()

	// Another goroutine may have completed a Load for the same type while we were spawning. If so, drop our
	// freshly-spawned subprocess and return their cached entry — first-writer-wins.
	if existing, ok := r.loaded[remoteType]; ok {
		client.Kill()
		r.touch(existing)
		return r.newHandle(existing), nil
	}

	entry := &loadedRemote{
		r:          rem,
		c:          client,
		remoteType: remoteType,
	}
	entry.drainCond = sync.NewCond(&r.mu)
	entry.lruElem = r.lru.PushBack(entry)
	r.loaded[remoteType] = entry

	return r.newHandle(entry), nil
}

// touch marks an entry as most-recently-used. Caller must hold r.mu.
func (r *Registry) touch(entry *loadedRemote) {
	if entry.lruElem != nil {
		r.lru.MoveToBack(entry.lruElem)
	}
}

// evictLRULocked picks the least-recently-used entry, drains in-flight calls, kills its subprocess, and removes it
// from the registry. Caller must hold r.mu. If the LRU list is empty, this is a no-op.
func (r *Registry) evictLRULocked() {
	front := r.lru.Front()
	if front == nil {
		return
	}
	entry, ok := front.Value.(*loadedRemote)
	if !ok {
		return
	}
	r.killEntryLocked(entry)
}

// killEntryLocked drains any in-flight calls on the entry, kills the subprocess, removes the entry from the loaded
// map and the LRU list, and fires the OnEvict callback. Caller must hold r.mu; the lock is briefly released while
// drainCond.Wait() is blocked, then re-acquired before mutation.
func (r *Registry) killEntryLocked(entry *loadedRemote) {
	for entry.refCount > 0 {
		entry.drainCond.Wait()
	}
	entry.c.Kill()
	if entry.lruElem != nil {
		r.lru.Remove(entry.lruElem)
		entry.lruElem = nil
	}
	delete(r.loaded, entry.remoteType)
	if r.config.onEvict != nil {
		r.config.onEvict(entry.remoteType)
	}
}

// Unload terminates the plugin subprocess for the given type and removes the cache entry. Waits for in-flight
// method calls to drain before killing the subprocess, so callers never see "connection refused" mid-RPC. Silently
// succeeds if the type was never loaded; use Loaded() to distinguish.
func (r *Registry) Unload(remoteType string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if entry, ok := r.loaded[remoteType]; ok {
		r.killEntryLocked(entry)
	}
}

// Loaded reports whether the given remote type currently has a live plugin subprocess in the cache.
func (r *Registry) Loaded(remoteType string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.loaded[remoteType]
	return ok
}

// Clear removes every registration and terminates every loaded plugin subprocess. Intended for test isolation —
// production code should not need to call it.
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.registered = map[string]Remote{}
	for _, entry := range r.loaded {
		// Drain even during Clear: callers in the middle of a method call shouldn't get connection-refused.
		for entry.refCount > 0 {
			entry.drainCond.Wait()
		}
		entry.c.Kill()
	}
	r.loaded = map[string]*loadedRemote{}
	r.lru.Init()
}

// newHandle returns the cached handle wrapper for an entry, creating it on first call. The wrapper is reused
// across Load() invocations so callers can compare returned values by pointer identity (assert.Same etc.). Every
// method call on the returned Remote increments the entry's refCount on entry and decrements it on return, so
// eviction can wait for in-flight calls to drain. Caller must hold r.mu.
func (r *Registry) newHandle(entry *loadedRemote) Remote {
	if entry.handle == nil {
		entry.handle = &handleRemote{reg: r, entry: entry}
	}
	return entry.handle
}

// handleRemote wraps a *loadedRemote so that every Remote method invocation increments and decrements the entry's
// refCount. This is what makes the in-flight drain on eviction work: while a method is executing, refCount > 0,
// and any concurrent evictLRULocked or Unload call blocks on drainCond until the method returns.
type handleRemote struct {
	reg   *Registry
	entry *loadedRemote
}

func (h *handleRemote) acquire() Remote {
	h.reg.mu.Lock()
	h.entry.refCount++
	rem := h.entry.r
	h.reg.mu.Unlock()
	return rem
}

func (h *handleRemote) release() {
	h.reg.mu.Lock()
	h.entry.refCount--
	if h.entry.refCount == 0 {
		h.entry.drainCond.Broadcast()
	}
	h.reg.mu.Unlock()
}

func (h *handleRemote) Type() (string, error) {
	rem := h.acquire()
	defer h.release()
	return rem.Type()
}

func (h *handleRemote) FromURL(url string, properties map[string]string) (map[string]interface{}, error) {
	rem := h.acquire()
	defer h.release()
	return rem.FromURL(url, properties)
}

func (h *handleRemote) ToURL(properties map[string]interface{}) (string, map[string]string, error) {
	rem := h.acquire()
	defer h.release()
	return rem.ToURL(properties)
}

func (h *handleRemote) GetParameters(properties map[string]interface{}) (map[string]interface{}, error) {
	rem := h.acquire()
	defer h.release()
	return rem.GetParameters(properties)
}

func (h *handleRemote) ValidateRemote(properties map[string]interface{}) error {
	rem := h.acquire()
	defer h.release()
	return rem.ValidateRemote(properties)
}

func (h *handleRemote) ValidateParameters(parameters map[string]interface{}) error {
	rem := h.acquire()
	defer h.release()
	return rem.ValidateParameters(parameters)
}

func (h *handleRemote) ListCommits(properties map[string]interface{}, parameters map[string]interface{}, tags []Tag) ([]Commit, error) {
	rem := h.acquire()
	defer h.release()
	return rem.ListCommits(properties, parameters, tags)
}

func (h *handleRemote) GetCommit(properties map[string]interface{}, parameters map[string]interface{}, commitID string) (*Commit, error) {
	rem := h.acquire()
	defer h.release()
	return rem.GetCommit(properties, parameters, commitID)
}
