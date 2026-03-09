package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	appconf "github.com/tab58/code-context/cmd/codectx/config"
	"github.com/tab58/code-context/internal/analysis"
	goextractor "github.com/tab58/code-context/internal/analysis/golang"
	tsextractor "github.com/tab58/code-context/internal/analysis/typescript"
	codedb "github.com/tab58/code-context/internal/clients/code_db"
	"github.com/tab58/code-context/internal/config"
	"github.com/tab58/code-context/internal/indexer"
	mcpserver "github.com/tab58/code-context/internal/mcp"
	"github.com/tab58/code-context/internal/repl"
)

var cfg *appconf.Config

func init() {
	conf, err := appconf.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	cfg = conf
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown on SIGINT/SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Printf("received signal %v, shutting down", sig)
		cancel()
	}()

	// Load FalkorDB configuration
	port, err := strconv.Atoi(cfg.FalkorDBPort)
	if err != nil {
		log.Fatalf("failed to convert port to int: %v", err)
	}
	falkorDBConfig := config.FalkorDBConfig{
		Host:      cfg.FalkorDBHost,
		Port:      port,
		Password:  cfg.FalkorDBPassword,
		GraphName: cfg.FalkorDBGraph,
	}

	// Connect to FalkorDB
	db, err := codedb.NewCodeDB(ctx, falkorDBConfig)
	if err != nil {
		log.Fatalf("failed to connect to FalkorDB: %v", err)
	}
	defer db.Close(ctx)

	// Create pipeline components
	idx := indexer.NewIndexer(db)
	registry := analysis.NewRegistry()
	goextractor.Register(registry)
	tsextractor.Register(registry)
	analyzer := analysis.NewAnalyzer(registry, db)

	// Create and start MCP server in goroutine
	server := mcpserver.NewServer(db, idx, analyzer)
	log.Printf("starting MCP server on HTTP :%s", cfg.MCPPort)
	go server.Serve(ctx, ":"+cfg.MCPPort)

	// Create and run REPL on main goroutine
	pipeline := repl.Pipeline{
		DB:       db,
		Indexer:  idx,
		Analyzer: analyzer,
	}
	status := repl.StatusInfo{
		FalkorDBHost: cfg.FalkorDBHost,
		FalkorDBPort: cfg.FalkorDBPort,
		MCPPort:      cfg.MCPPort,
	}
	r := repl.New(pipeline, status)
	if err := r.Run(ctx); err != nil {
		log.Printf("REPL error: %v", err)
	}

	log.Println("shutting down")
}
