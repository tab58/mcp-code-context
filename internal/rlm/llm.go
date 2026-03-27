package rlm

import "context"

// LLMClient is a stateless interface for LLM chat completions.
// Used for both the root LLM (multi-turn orchestrator) and sub-LLM (single-turn analyst).
type LLMClient interface {
	Complete(ctx context.Context, messages []Message) (string, error)
}
