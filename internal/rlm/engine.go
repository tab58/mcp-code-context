package rlm

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

var (
	codeBlockRe = regexp.MustCompile("(?s)```repl\\s*\n(.*?)```")
	finalVarRe  = regexp.MustCompile(`FINAL_VAR\("([^"]+)"\)`)
	finalQuotRe = regexp.MustCompile(`FINAL\("([^"]+)"\)`)
	finalRe     = regexp.MustCompile(`FINAL\(([^)]+)\)`)
)

// EngineConfig controls the RLM loop behavior.
type EngineConfig struct {
	RootLLM       LLMClient
	SubLLM        LLMClient
	Graph         KnowledgeGraph
	MaxIterations int         // Upper bound on loop turns. Default: 30.
	TruncateMax   int         // Max chars for stdout truncation. Default: 2000.
	TraceLogger   TraceLogger // Optional logger for LLM and REPL events.
}

// Engine orchestrates the RLM loop between the root LLM, REPL, and sub-LLM.
type Engine struct {
	rootLLM LLMClient
	subLLM  LLMClient
	graph   KnowledgeGraph

	maxIterations int
	truncateMax   int
	trace         TraceLogger
}

// NewEngine creates an RLM engine wired to the given LLM clients and graph backend.
// RootLLM and Graph are required; SubLLM falls back to RootLLM if nil.
func NewEngine(cfg EngineConfig) (*Engine, error) {
	if cfg.RootLLM == nil {
		return nil, fmt.Errorf("root LLM client is required")
	}
	if cfg.Graph == nil {
		return nil, fmt.Errorf("knowledge graph is required")
	}
	subLLM := cfg.SubLLM
	if subLLM == nil {
		subLLM = cfg.RootLLM
	}

	maxIterations := cfg.MaxIterations
	if maxIterations <= 0 {
		maxIterations = 30
	}
	truncateMax := cfg.TruncateMax
	if truncateMax <= 0 {
		truncateMax = 2000
	}

	return &Engine{
		rootLLM:       cfg.RootLLM,
		subLLM:        subLLM,
		graph:         cfg.Graph,
		maxIterations: maxIterations,
		truncateMax:   truncateMax,
		trace:         cfg.TraceLogger,
	}, nil
}

// Run executes the RLM inference loop for a single user query and returns the final answer.
func (e *Engine) Run(ctx context.Context, query string) (string, error) {
	repl := NewREPL(e.graph, e.subLLM)

	meta, err := e.graph.Metadata(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get graph metadata: %w", err)
	}

	systemPrompt, err := BuildSystemPrompt(meta)
	if err != nil {
		return "", fmt.Errorf("failed to build system prompt: %w", err)
	}
	history := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: query},
	}

	for i := 0; i < e.maxIterations; i++ {
		turn := i + 1
		response, err := e.rootLLM.Complete(ctx, history)
		if err != nil {
			return "", fmt.Errorf("root LLM error on turn %d: %w", turn, err)
		}

		if e.trace != nil {
			e.trace.LogLLMResponse(turn, response)
		}

		// Check for FINAL/FINAL_VAR in text outside code blocks.
		if answer, ok := detectFinalInText(response); ok {
			if e.trace != nil {
				e.trace.LogFinal(turn, answer)
			}
			return answer, nil
		}

		codeBlocks := extractCodeBlocks(response)

		// No code blocks — nudge the root LLM.
		if len(codeBlocks) == 0 {
			history = append(history,
				Message{Role: "assistant", Content: response},
				Message{Role: "user", Content: "[No code block detected. Write ```repl code to interact with the graph, or use FINAL(answer) to return your answer.]"},
			)
			continue
		}

		// Execute code blocks sequentially.
		var allOutput strings.Builder
		for j, code := range codeBlocks {
			if e.trace != nil {
				e.trace.LogREPLInput(turn, j, code)
			}

			stdout := repl.Execute(ctx, code)
			allOutput.WriteString(stdout)

			if e.trace != nil {
				e.trace.LogREPLOutput(turn, j, stdout)
			}

			if repl.Done {
				if e.trace != nil {
					e.trace.LogFinal(turn, repl.FinalAnswer)
				}
				return repl.FinalAnswer, nil
			}
		}

		truncated := Truncate(allOutput.String(), e.truncateMax)
		history = append(history,
			Message{Role: "assistant", Content: response},
			Message{Role: "user", Content: "REPL output:\n" + truncated},
		)
	}

	return "", fmt.Errorf("exceeded max iterations (%d)", e.maxIterations)
}

// extractCodeBlocks finds all ```repl code blocks in a response.
func extractCodeBlocks(response string) []string {
	matches := codeBlockRe.FindAllStringSubmatch(response, -1)
	blocks := make([]string, 0, len(matches))
	for _, m := range matches {
		blocks = append(blocks, m[1])
	}
	return blocks
}

// detectFinalInText checks for FINAL(...) or FINAL_VAR(...) in text outside code blocks.
func detectFinalInText(response string) (string, bool) {
	// Strip code blocks so we only check prose text.
	stripped := codeBlockRe.ReplaceAllString(response, "")

	// FINAL_VAR(name) in text — return the variable name as a graceful fallback.
	if m := finalVarRe.FindStringSubmatch(stripped); m != nil {
		return m[1], true
	}

	// FINAL("...") with a quoted string.
	if m := finalQuotRe.FindStringSubmatch(stripped); m != nil {
		return m[1], true
	}

	// FINAL(unquoted) — take everything inside the parens.
	if m := finalRe.FindStringSubmatch(stripped); m != nil {
		return strings.TrimSpace(m[1]), true
	}

	return "", false
}
