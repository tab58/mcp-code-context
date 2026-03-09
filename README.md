# mcp-code-context

A Model Context Protocol (MCP) server that indexes code repositories into a FalkorDB graph database and exposes code intelligence tools to AI agents. Point it at a codebase, and it builds a knowledge graph of every file, function, class, module, and their relationships — then lets AI agents search and traverse that graph via MCP tool calls.

## What It Does

code-context turns source code into a queryable knowledge graph:

1. **Indexes** a repository's directory structure (files, folders, `.gitignore` rules, binary/symlink detection)
2. **Parses** source files into ASTs using tree-sitter, extracting functions, classes, modules, imports, and call relationships
3. **Serves** 8 MCP tools over Streamable HTTP so AI agents can search and traverse the graph
4. **Provides** an interactive REPL for indexing repositories and managing the server

### Supported Languages

- Go
- TypeScript / TSX

## MCP Tools

### Search Tools

| Tool            | Description                                          | Key Parameters                 |
| --------------- | ---------------------------------------------------- | ------------------------------ |
| `find_function` | Exact function name lookup                           | `repository`, `name`           |
| `find_file`     | Glob pattern match on file paths, enriched with symbols | `repository`, `pattern`     |
| `search_code`   | Unified search with automatic strategy selection     | `repository`, `query`, `limit` |

`search_code` classifies queries using heuristics (glob patterns, path-like strings, camelCase/snake_case identifiers) and dispatches to the appropriate strategy — file search or exact function match. Falls back to per-token exact matching when unsure.

### Traversal Tools

| Tool                  | Description                                               | Key Parameters                              |
| --------------------- | --------------------------------------------------------- | ------------------------------------------- |
| `get_callers`         | Find functions that call a given function                 | `repository`, `name`, `depth` (1-3)         |
| `get_callees`         | Find functions called by a given function                 | `repository`, `name`, `depth` (1-3)         |
| `get_class_hierarchy` | Find parent/child classes and interface implementations   | `repository`, `name`, `direction`, `depth`  |
| `get_dependencies`    | Find module dependencies or file imports                  | `repository`, `name`, `depth` (1-3)         |
| `get_references`      | Find all references to a symbol (auto-detects type)       | `repository`, `name`                        |

Traversal tools support multi-hop graph traversal (up to depth 3) and return lightweight results (name, path, signature) without source code.

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

This runs the full pipeline (index → analyze). The repository name used in MCP tool calls will be the directory name of the ingested path.

## REPL Commands

When you run `codectx`, it starts the MCP server in the background and opens an interactive REPL:

| Command         | Description                                                    |
| --------------- | -------------------------------------------------------------- |
| `ingest <path>` | Run full pipeline (index → analyze) on a local directory       |
| `list`          | List all indexed repositories                                  |
| `status`        | Display current configuration (FalkorDB host/port, MCP port)   |
| `help`          | Show available commands                                        |
| `quit`          | Shut down the application                                      |

## Knowledge Graph

The graph schema has 7 node types across two layers:

**Structural Layer** — mirrors the filesystem:

- `Repository` — top-level project with name, path, remote URL, primary language
- `Folder` — directory within a repository
- `File` — source file with language and line count

**Code Layer** — extracted from AST parsing:

- `Module` — package/module declaration with import path, visibility, kind (ESM/CJS)
- `Class` — class, struct, interface, or enum with optional decorators
- `Function` — function or method with signature and cyclomatic complexity
- `ExternalReference` — external package import with import path

Relationships include `CONTAINS`, `BELONGS_TO`, `DEFINES`, `IMPORTS`, `CALLS`, `INHERITS`, `IMPLEMENTS`, `OVERRIDES`, `HAS_METHOD`, `HAS_MODULE`, `EXPORTS`, and `DEPENDS_ON`.

## Prerequisites

- Go 1.25+
- C/C++ compiler (CGO required for tree-sitter)
- [FalkorDB](https://falkordb.com/) instance (Redis protocol, default port 6379)
- [Task](https://taskfile.dev/) runner (optional, for build commands)
- [go-ormql CLI](https://github.com/tab58/go-ormql) (`go install github.com/tab58/go-ormql/cmd/gormql@latest`)

## Setup

1. Clone the repository:

   ```bash
   git clone https://github.com/tab58/code-context.git
   cd code-context
   ```

2. Start FalkorDB (via Docker):

   ```bash
   task env-up
   ```

   This starts FalkorDB on port 6379 (Redis) and 3030 (web UI).

3. Generate code from the GraphQL schema:

   ```bash
   task generate
   ```

4. Build:

   ```bash
   task build
   ```

5. Configure environment variables (or use a `.env` file):

   ```bash
   FALKORDB_HOST=localhost       # required
   FALKORDB_PORT=6379            # default: 6379
   FALKORDB_PASSWORD=your-pass   # required
   FALKORDB_GRAPH=codecontext    # default: codecontext
   MCP_PORT=8080                 # default: 8080
   ```

6. Run:
   ```bash
   ./build/codectx
   ```
   This starts the MCP server on `http://localhost:8080` and opens an interactive REPL for indexing repositories. See [Connecting an Agent](#connecting-an-agent) for how to wire up an AI client.

## Project Structure

```
cmd/
  codectx/
    main.go              -- Entry point: config, FalkorDB connect, pipeline setup, MCP server
    config/config.go     -- Environment variable loader (.env support via godotenv)

internal/
  clients/code_db/
    schema.graphql       -- Source-of-truth GraphQL schema (7 nodes, 12 relationships)
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
    external_refs.go     -- External reference classification and persistence
    types.go             -- Symbol, Reference, FileAnalysis types
    golang/              -- Go language extractor
    typescript/          -- TypeScript + TSX extractors

  repl/
    repl.go              -- Interactive REPL loop (stdin/stdout)
    commands.go          -- Command handlers: ingest, list, status, help

  mcp/
    server.go            -- MCP server setup, tool registration, Streamable HTTP transport
    tools.go             -- Search tool implementations (find_function, find_file, search_code)
    traversal.go         -- Graph traversal tools (callers, callees, hierarchy, deps, refs)
    search.go            -- Query classification and search strategies
    types.go             -- SearchResult, TraversalResult, response types

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
      |         Pass 1: extract symbols (Function, Class, Module, ExternalReference nodes)
      |         Pass 2: resolve relationships (CALLS, IMPORTS, INHERITS, etc.)
      v
  MCP Server - exposes 8 tools (3 search + 5 traversal) over Streamable HTTP
               AI agents query the graph via tool calls
```

### Schema-Driven Codegen

The GraphQL schema in `schema.graphql` is the single source of truth. Running `gormql generate --target falkordb` produces:

- Go structs for all node types and inputs
- A typed client with GraphQL-to-Cypher translation

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
