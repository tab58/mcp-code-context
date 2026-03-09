package config

// FalkorDBConfig holds FalkorDB connection settings.
// FalkorDB uses Redis protocol — no Scheme or Username fields needed.
// All repositories share a single graph; GraphName controls its name.
type FalkorDBConfig struct {
	Host      string // required
	Port      int    // default 6379
	Password  string // optional (FalkorDB supports no-auth for local dev)
	GraphName string // shared graph name (default "codecontext")
}
