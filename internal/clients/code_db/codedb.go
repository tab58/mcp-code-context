package codedb

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/tab58/code-context/internal/clients/code_db/generated"
	"github.com/tab58/code-context/internal/config"
	"github.com/tab58/go-ormql/pkg/client"
	"github.com/tab58/go-ormql/pkg/cypher"
	"github.com/tab58/go-ormql/pkg/driver"
	falkordbdrv "github.com/tab58/go-ormql/pkg/driver/falkordb"
)

// defaultScheme is the connection scheme used for FalkorDB (Redis protocol).
const defaultScheme = "redis"

// defaultGraphName is the shared graph name when none is configured.
const defaultGraphName = "codecontext"

// gqlListRepositories queries all Repository nodes in the shared graph.
const gqlListRepositories = `query { repositorys { name } }`

// sharedClient holds the single driver and client for the shared graph.
type sharedClient struct {
	drv    driver.Driver
	client *client.Client
}

// CodeDB manages a single shared FalkorDB graph that holds all repository
// data. Repositories are logically isolated via BELONGS_TO edges and
// query-level filtering. The shared client is lazily initialized on first
// ForRepo call.
type CodeDB struct {
	mu     sync.Mutex
	cfg    config.FalkorDBConfig
	shared *sharedClient
	opts   []Option
	closed bool
}

// Option configures CodeDB construction.
type Option func(*options)

type options struct {
	driver driver.Driver
}

// WithDriver injects a pre-built driver (useful for testing).
func WithDriver(drv driver.Driver) Option {
	return func(o *options) {
		o.driver = drv
	}
}

// applyOptions builds an options struct from the stored Option funcs.
func (db *CodeDB) applyOptions() *options {
	o := &options{}
	for _, opt := range db.opts {
		opt(o)
	}
	return o
}

// graphName returns the configured graph name, falling back to the default.
func (db *CodeDB) graphName() string {
	if db.cfg.GraphName != "" {
		return db.cfg.GraphName
	}
	return defaultGraphName
}

// NewCodeDB validates the config and returns a CodeDB instance.
// No driver connection or index creation happens at boot — those are
// deferred to ForRepo. Call Close when done.
func NewCodeDB(ctx context.Context, cfg config.FalkorDBConfig, opts ...Option) (*CodeDB, error) {
	if ctx == nil {
		return nil, errors.New("codedb: context must not be nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("codedb: context error: %w", err)
	}
	if cfg.Host == "" {
		return nil, errors.New("codedb: Host is required")
	}
	if cfg.Port <= 0 {
		return nil, fmt.Errorf("codedb: Port must be positive, got %d", cfg.Port)
	}

	return &CodeDB{
		cfg:  cfg,
		opts: opts,
	}, nil
}

// initShared lazily creates the shared driver and client. Must be called
// with db.mu held.
func (db *CodeDB) initShared(ctx context.Context) error {
	if db.shared != nil {
		return nil
	}

	o := db.applyOptions()
	drv := o.driver
	if drv == nil {
		drvCfg := driver.Config{
			Host:         db.cfg.Host,
			Port:         db.cfg.Port,
			Scheme:       defaultScheme,
			Password:     db.cfg.Password,
			Database:     db.graphName(),
			ReadTimeout:  5 * time.Minute,
			WriteTimeout: 5 * time.Minute,
		}

		var err error
		drv, err = falkordbdrv.NewFalkorDBDriver(drvCfg)
		if err != nil {
			return fmt.Errorf("codedb: failed to connect to FalkorDB graph %q: %w", db.graphName(), err)
		}
	}

	if err := createIndexes(ctx, drv); err != nil {
		if o.driver == nil {
			_ = drv.Close(ctx)
		}
		return fmt.Errorf("codedb: failed to create indexes for graph %q: %w", db.graphName(), err)
	}

	c := generated.NewClient(drv)
	db.shared = &sharedClient{drv: drv, client: c}
	return nil
}

// ForRepo returns a *client.Client for the shared graph. The name parameter
// identifies the repository for callers but does not affect which graph is
// used — all repositories share a single graph.
func (db *CodeDB) ForRepo(ctx context.Context, _ string) (*client.Client, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.closed {
		return nil, errors.New("codedb: closed")
	}

	if err := db.initShared(ctx); err != nil {
		return nil, err
	}

	return db.shared.client, nil
}

// ListRepos queries the shared graph for all Repository nodes and returns
// their names sorted alphabetically.
func (db *CodeDB) ListRepos(ctx context.Context) ([]string, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.closed {
		return nil, errors.New("codedb: closed")
	}

	if err := db.initShared(ctx); err != nil {
		return nil, fmt.Errorf("codedb: failed to initialize shared client: %w", err)
	}

	result, err := db.shared.client.Execute(ctx, gqlListRepositories, nil)
	if err != nil {
		return nil, fmt.Errorf("codedb: failed to list repositories: %w", err)
	}

	var names []string
	data := result.Data()
	repos, ok := data["repositorys"].([]any)
	if !ok {
		sort.Strings(names)
		return names, nil
	}
	for _, r := range repos {
		repo, ok := r.(map[string]any)
		if !ok {
			continue
		}
		if name, ok := repo["name"].(string); ok {
			names = append(names, name)
		}
	}

	sort.Strings(names)
	return names, nil
}

// DeleteRepo removes all nodes and edges belonging to a repository from the
// shared graph. Deletes the Repository node, plus all Folder, File, Function,
// Class, Module, and ExternalReference nodes connected via BELONGS_TO.
func (db *CodeDB) DeleteRepo(ctx context.Context, repoName string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.closed {
		return errors.New("codedb: closed")
	}
	if repoName == "" {
		return errors.New("codedb: repository name is required")
	}

	if err := db.initShared(ctx); err != nil {
		return fmt.Errorf("codedb: failed to initialize shared client: %w", err)
	}

	// Delete all nodes that BELONG_TO the repository, then the repository itself.
	// Two-step: first delete dependent nodes (with edges), then the repo node.
	deleteDependent := cypher.Statement{
		Query: "MATCH (r:Repository {name: $name})<-[:BELONGS_TO]-(n) DETACH DELETE n",
		Params: map[string]any{"name": repoName},
	}
	if _, err := db.shared.drv.ExecuteWrite(ctx, deleteDependent); err != nil {
		return fmt.Errorf("codedb: failed to delete dependent nodes for %q: %w", repoName, err)
	}

	deleteRepo := cypher.Statement{
		Query:  "MATCH (r:Repository {name: $name}) DETACH DELETE r",
		Params: map[string]any{"name": repoName},
	}
	if _, err := db.shared.drv.ExecuteWrite(ctx, deleteRepo); err != nil {
		return fmt.Errorf("codedb: failed to delete repository %q: %w", repoName, err)
	}

	return nil
}

// Close shuts down the shared FalkorDB driver connection.
// Safe to call multiple times.
func (db *CodeDB) Close(ctx context.Context) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.closed {
		return nil
	}
	db.closed = true

	if db.shared != nil && db.shared.drv != nil {
		return db.shared.drv.Close(ctx)
	}
	return nil
}
