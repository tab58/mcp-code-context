package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/tab58/code-context/api/rest"
	appconf "github.com/tab58/code-context/cmd/codectx/config"
	"github.com/tab58/code-context/internal/analysis"
	goextractor "github.com/tab58/code-context/internal/analysis/golang"
	jsextractor "github.com/tab58/code-context/internal/analysis/javascript"
	pyextractor "github.com/tab58/code-context/internal/analysis/python"
	rbextractor "github.com/tab58/code-context/internal/analysis/ruby"
	tsextractor "github.com/tab58/code-context/internal/analysis/typescript"
	"github.com/tab58/code-context/internal/app"
	codedb "github.com/tab58/code-context/internal/clients/code_db"
	"github.com/tab58/code-context/internal/config"
	"github.com/tab58/code-context/internal/indexer"
	"github.com/tab58/code-context/internal/rlm"

	"github.com/tab58/go-ormql/pkg/driver"
	falkordbdrv "github.com/tab58/go-ormql/pkg/driver/falkordb"
)

var Version string

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
	// Connect to FalkorDB — single driver shared by CodeDB and RLM engine
	drv, err := falkordbdrv.NewFalkorDBDriver(driver.Config{
		Host:         cfg.FalkorDBHost,
		Port:         port,
		Scheme:       "redis",
		Password:     cfg.FalkorDBPassword,
		Database:     cfg.FalkorDBGraph,
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
	})
	if err != nil {
		log.Fatalf("failed to connect to FalkorDB: %v", err)
	}

	db, err := codedb.NewCodeDB(ctx, config.FalkorDBConfig{
		Host:      cfg.FalkorDBHost,
		Port:      port,
		Password:  cfg.FalkorDBPassword,
		GraphName: cfg.FalkorDBGraph,
	}, codedb.WithDriver(drv))
	if err != nil {
		log.Fatalf("failed to connect to FalkorDB: %v", err)
	}

	// Create optional trace logger for the RLM engine
	var traceLogger rlm.TraceLogger
	if cfg.RLMTraceLog != "" {
		tl, err := rlm.NewFileTraceLogger(cfg.RLMTraceLog)
		if err != nil {
			log.Fatalf("failed to open RLM trace log: %v", err)
		}
		defer tl.Close()
		traceLogger = tl
		log.Printf("RLM trace logging to %s", cfg.RLMTraceLog)
	}

	// Create RLM engine
	rlmEngine, err := rlm.NewEngine(rlm.EngineConfig{
		RootLLM: rlm.NewAnthropicLLM(rlm.AnthropicConfig{
			APIKey: cfg.AnthropicAPIKey,
			Model:  rlm.AnthropicClaudeOpus,
		}),
		SubLLM: rlm.NewAnthropicLLM(rlm.AnthropicConfig{
			APIKey: cfg.AnthropicAPIKey,
			Model:  rlm.AnthropicClaudeSonnet,
		}),
		Graph:         rlm.NewFalkorGraph(drv),
		MaxIterations: 10,
		TruncateMax:   1000,
		TraceLogger:   traceLogger,
	})
	if err != nil {
		log.Fatalf("failed to create RLM engine: %v", err)
	}

	// create application dependencies
	idx := indexer.NewIndexer(db)
	registry := analysis.NewRegistry()
	goextractor.Register(registry)
	tsextractor.Register(registry)
	jsextractor.Register(registry)
	pyextractor.Register(registry)
	rbextractor.Register(registry)
	analyzer := analysis.NewAnalyzer(registry, db)
	// Create application
	application := app.NewApplication(&app.ApplicationConfig{
		AppVersion:  Version,
		DB:          db,
		Indexer:     idx,
		Analyzer:    analyzer,
		QueryEngine: rlmEngine,
	})

	// Create and start MCP server in goroutine
	// server := mcpserver.NewServer(application)
	// log.Printf("starting MCP server on HTTP :%s", cfg.MCPPort)
	// go func() {
	// 	if err := server.Serve(ctx, ":"+cfg.MCPPort); err != nil {
	// 		log.Printf("MCP server error: %v", err)
	// 	}
	// }()

	// Create and run REPL on main goroutine
	// pipeline := repl.Pipeline{
	// 	DB:       db,
	// 	Indexer:  idx,
	// 	Analyzer: analyzer,
	// }
	// status := repl.StatusInfo{
	// 	FalkorDBHost: cfg.FalkorDBHost,
	// 	FalkorDBPort: cfg.FalkorDBPort,
	// 	MCPPort:      cfg.MCPPort,
	// }
	// r := repl.New(pipeline, status)
	// if err := r.Run(ctx); err != nil {
	// 	log.Printf("REPL error: %v", err)
	// }

	restAddr := ":" + cfg.ServerPort
	restServer := rest.NewServer(application)
	restServer.Start(restAddr)
	log.Printf("REST server listening on %s", restAddr)

	// Block until shutdown signal
	<-ctx.Done()
	if err := restServer.Stop(context.Background()); err != nil {
		log.Printf("REST server shutdown error: %v", err)
	}
	log.Println("shutting down")
}
