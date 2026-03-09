# code-context

A Model Context Protocol (MCP) server that indexes code repositories into a FalkorDB graph database and exposes code intelligence tools to AI agents. Point it at a codebase, and it builds a knowledge graph of every file, function, class, module, and their relationships — then lets AI agents search and explore that graph via MCP tool calls.

## What It Does

code-context turns source code into a queryable knowledge graph:

1. **Indexes** a repository's directory structure (files, folders, `.gitignore` rules, binary/symlink detection)
2. **Parses** source files into ASTs using tree-sitter, extracting functions, classes, modules, imports, and call relationships
3. **Embeds** function and class source text as 768-dimensional vectors using nomic-embed-text for semantic search
4. **Serves** 4 MCP tools over Streamable HTTP so AI agents can search the graph
5. **Provides** an interactive REPL for indexing repositories and managing the server

### Supported Languages

- Go
- TypeScript / TSX

## MCP Tools

| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `find_function` | Exact function name lookup | `repository`, `name` |
| `find_file` | Glob pattern match on file paths, enriched with symbols | `repository`, `pattern` |
| `vector_search` | Semantic similarity search on functions and classes | `repository`, `query`, `limit` |
| `search_code` | Unified search with automatic strategy selection | `repository`, `query`, `limit` |

`search_code` classifies queries using heuristics (glob patterns, path-like strings, camelCase/snake_case identifiers, natural language) and dispatches to the appropriate strategy. Falls back to hybrid (exact + vector) when unsure.

## Connecting an Agent

code-context exposes its MCP tools over [Streamable HTTP](https://modelcontextprotocol.io/specification/2025-03-26/basic/transports#streamable-http) on port 8080 (configurable via `MCP_PORT`). Any MCP-compatible client can connect to it.

### Claude Code

Add the server to your Claude Code MCP settings (`~/.claude/settings.json` or `.mcp.json`):

```json
{
  "mcpServers": {
    "code-context": {
      "type": "url",
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

### Claude Desktop

Add to your Claude Desktop config (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "code-context": {
      "type": "url",
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

### Other MCP Clients

Point any MCP client at the Streamable HTTP endpoint:

```
URL: http://localhost:8080/mcp
Transport: Streamable HTTP
```

### Indexing a Repository

Before an agent can search a repository, you must index it using the REPL:

```bash
./codectx
> ingest /path/to/your/repo
```

This runs the full pipeline (index → analyze → embed). The repository name used in MCP tool calls will be the directory name of the ingested path.

## REPL Commands

When you run `codectx`, it starts the MCP server in the background and opens an interactive REPL:

| Command | Description |
|---------|-------------|
| `ingest <path>` | Run full pipeline (index → analyze → embed) on a local directory |
| `list` | List all indexed repositories with last-indexed timestamps |
| `status` | Display current configuration (database, MCP port, embedding model) |
| `help` | Show available commands |
| `quit` | Shut down the application |

## Knowledge Graph

The graph schema has 6 node types across two layers:

**Structural Layer** — mirrors the filesystem:
- `Repository` — top-level project
- `Folder` — directory within a repository
- `File` — source file with language and line count

**Code Layer** — extracted from AST parsing:
- `Module` — package/module declaration
- `Class` — class, struct, interface, or enum
- `Function` — function or method, with optional embedding vector

Relationships include `CONTAINS`, `BELONGS_TO`, `DEFINES`, `IMPORTS`, `CALLS`, `INHERITS`, `IMPLEMENTS`, `OVERRIDES`, `HAS_METHOD`, `EXPORTS`, and `DEPENDS_ON`.

## Prerequisites

- Go 1.25+
- C/C++ compiler (CGO required for tree-sitter and llama.cpp)
- [FalkorDB](https://falkordb.com/) instance (Redis protocol, default port 6379)
- [Task](https://taskfile.dev/) runner (optional, for build commands)
- [go-ormql CLI](https://github.com/tab58/go-ormql) (`go install github.com/tab58/go-ormql/cmd/gormql@latest`)
- nomic-embed-text GGUF model file

## Setup

1. Clone with submodules:
   ```bash
   git clone --recurse-submodules https://github.com/tab58/code-context.git
   cd code-context
   ```

2. Build the llama.cpp embedding shim (one-time):
   ```bash
   task build-llama
   ```

3. Generate code from the GraphQL schema:
   ```bash
   task generate
   ```

4. Build:
   ```bash
   CGO_ENABLED=1 task build
   ```

5. Configure environment variables (or use a `.env` file):
   ```bash
   FALKORDB_HOST=localhost       # required
   FALKORDB_PORT=6379            # default: 6379
   FALKORDB_PASSWORD=your-pass   # required
   FALKORDB_DATABASE=codecontext # default: codecontext
   MCP_PORT=8080                 # default: 8080
   EMBEDDING_MODEL_PATH=models/nomic-embed-text  # path to GGUF model
   ```

6. Run:
   ```bash
   ./codectx
   ```
   This starts the MCP server on `http://localhost:8080` and opens an interactive REPL for indexing repositories. See [Connecting an Agent](#connecting-an-agent) for how to wire up an AI client.

## Project Structure

```
cmd/
  codectx/
    main.go              -- Entry point: config, FalkorDB connect, pipeline setup, MCP server
    config/config.go     -- Environment variable loader (.env support via godotenv)
  embed/main.go          -- Standalone embedding utility

internal/
  clients/code_db/
    schema.graphql       -- Source-of-truth GraphQL schema (6 nodes, 2 vector indexes)
    codedb.go            -- CodeDB wrapper: FalkorDB driver lifecycle, typed client
    generated/           -- go-ormql generated code (models, client, indexes)

  config/
    falkordb.go          -- FalkorDBConfig struct

  indexer/
    indexer.go           -- Directory walker + structural graph persistence
    gitignore.go         -- .gitignore pattern matching
    detect.go            -- Binary detection, language detection, utilities

  analysis/
    analyzer.go          -- Two-pass orchestrator: AST parse then relationship resolution
    registry.go          -- Maps file extensions to tree-sitter grammars + extractors
    extractor.go         -- Extractor interface
    types.go             -- Symbol, Reference, FileAnalysis types
    golang/              -- Go language extractor
    typescript/          -- TypeScript + TSX extractors

  search/
    embedder.go          -- Vector embedding pipeline for Function/Class nodes

  repl/
    repl.go              -- Interactive REPL loop (stdin/stdout)
    commands.go          -- Command handlers: ingest, list, status, help

  mcp/
    server.go            -- MCP server setup, tool registration, Streamable HTTP transport
    tools.go             -- Tool handler implementations
    search.go            -- Query classification and search strategies
    types.go             -- SearchResult, SearchResponse types

  embedding/
    llama.go             -- CGo bindings to llama.cpp for nomic-embed-text
    shim/                -- C shim for llama.cpp embedding API
    llama.cpp/           -- llama.cpp submodule

testinfra/
  testcontainers.go      -- FalkorDB Testcontainers helper for integration tests
```

## How It Works

### Pipeline

```
Repository path
      |
      v
  Indexer ---- walks directory tree, respects .gitignore
      |         creates Repository/Folder/File nodes in FalkorDB
      v
  Analyzer --- parses each file with tree-sitter
      |         Pass 1: extract symbols (Function, Class, Module nodes)
      |         Pass 2: resolve relationships (CALLS, IMPORTS, INHERITS, etc.)
      v
  Embedder --- queries Function/Class nodes missing embeddings
      |         computes 768-dim vectors via nomic-embed-text
      |         writes embedding field back to FalkorDB
      v
  MCP Server - exposes 4 search tools over Streamable HTTP
               AI agents query the graph via tool calls
```

### Schema-Driven Codegen

The GraphQL schema in `schema.graphql` is the single source of truth. Running `gormql generate --target falkordb` produces:
- Go structs for all node types and inputs
- A typed client with GraphQL-to-Cypher translation
- Vector index DDL for the `function_embedding` and `class_embedding` indexes

All graph reads and writes go through `Client().Execute()` with GraphQL mutation strings — no raw Cypher in application code.

### Incremental Re-indexing

The indexer compares filesystem modification times against `lastUpdated` timestamps stored on graph nodes. Unchanged files and folders are skipped, so re-indexing a large repository only processes what changed.

## Testing

```bash
go test ./...              # unit tests
go test -race ./...        # with race detection
go test ./... -cover       # with coverage

# integration tests (requires Docker for Testcontainers)
go test -tags integration ./internal/indexer/
```

## License

See [LICENSE](LICENSE) for details.
