package rlm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/dop251/goja"
)

// REPL is a persistent JavaScript runtime that holds state across turns.
// It exposes graph traversal functions and llm_query() as global JS functions.
type REPL struct {
	runtime *goja.Runtime
	stdout  *bytes.Buffer
	graph   KnowledgeGraph
	subLLM  LLMClient

	// ctx is set at the start of each Execute call so that JS bindings
	// use the caller's context for cancellation and deadlines.
	ctx context.Context

	// Done is set to true when FINAL() or FINAL_VAR() is called.
	Done bool
	// FinalAnswer holds the answer string after termination.
	FinalAnswer string
}

// NewREPL creates a JavaScript runtime with all graph and LLM bindings registered.
func NewREPL(graph KnowledgeGraph, subLLM LLMClient) *REPL {
	r := &REPL{
		runtime: goja.New(),
		stdout:  &bytes.Buffer{},
		graph:   graph,
		subLLM:  subLLM,
	}
	r.registerBuiltins()
	return r
}

// Execute runs a block of JavaScript code and returns the captured stdout.
// The stdout buffer is reset before each execution. If the code throws,
// the error message is returned prefixed with [JS Error].
func (r *REPL) Execute(ctx context.Context, code string) string {
	r.ctx = ctx
	r.stdout.Reset()

	_, err := r.runtime.RunString(code)
	if err != nil {
		return fmt.Sprintf("[JS Error] %s", err.Error())
	}
	return r.stdout.String()
}

func (r *REPL) registerBuiltins() {
	r.registerPrint()
	r.registerGraphFunctions()
	r.registerLLMQuery()
	r.registerFinal()
}

func (r *REPL) registerPrint() {
	printFn := func(call goja.FunctionCall) goja.Value {
		for i, arg := range call.Arguments {
			if i > 0 {
				r.stdout.WriteString(" ")
			}
			r.stdout.WriteString(arg.String())
		}
		r.stdout.WriteString("\n")
		return goja.Undefined()
	}
	r.runtime.Set("print", printFn)

	// console.log as alias
	console := r.runtime.NewObject()
	console.Set("log", printFn)
	r.runtime.Set("console", console)
}

func (r *REPL) registerGraphFunctions() {
	r.runtime.Set("graph_metadata", func(call goja.FunctionCall) goja.Value {
		meta, err := r.graph.Metadata(r.ctx)
		if err != nil {
			panic(r.runtime.NewGoError(err))
		}
		return r.marshalToJS(meta)
	})

	r.runtime.Set("graph_search", func(call goja.FunctionCall) goja.Value {
		query := call.Argument(0).String()

		var labels []string
		if arg := call.Argument(1); arg != nil && !goja.IsUndefined(arg) && !goja.IsNull(arg) {
			if err := r.runtime.ExportTo(arg, &labels); err != nil {
				panic(r.runtime.NewGoError(fmt.Errorf("graph_search: invalid labels argument: %w", err)))
			}
		}

		limit := 20
		if arg := call.Argument(2); arg != nil && !goja.IsUndefined(arg) && !goja.IsNull(arg) {
			limit = int(arg.ToInteger())
		}

		nodes, err := r.graph.Search(r.ctx, query, labels, limit)
		if err != nil {
			panic(r.runtime.NewGoError(err))
		}
		return r.marshalToJS(nodes)
	})

	r.runtime.Set("graph_neighbors", func(call goja.FunctionCall) goja.Value {
		nodeID := call.Argument(0).String()

		edgeType := ""
		if arg := call.Argument(1); arg != nil && !goja.IsUndefined(arg) && !goja.IsNull(arg) {
			edgeType = arg.String()
		}

		limit := 50
		if arg := call.Argument(2); arg != nil && !goja.IsUndefined(arg) && !goja.IsNull(arg) {
			limit = int(arg.ToInteger())
		}

		sg, err := r.graph.Neighbors(r.ctx, nodeID, edgeType, limit)
		if err != nil {
			panic(r.runtime.NewGoError(err))
		}
		return r.marshalToJS(sg)
	})

	r.runtime.Set("graph_get_node", func(call goja.FunctionCall) goja.Value {
		id := call.Argument(0).String()
		node, err := r.graph.GetNode(r.ctx, id)
		if err != nil {
			panic(r.runtime.NewGoError(err))
		}
		return r.marshalToJS(node)
	})

	r.runtime.Set("graph_shortest_path", func(call goja.FunctionCall) goja.Value {
		fromID := call.Argument(0).String()
		toID := call.Argument(1).String()

		maxHops := 6
		if arg := call.Argument(2); arg != nil && !goja.IsUndefined(arg) && !goja.IsNull(arg) {
			maxHops = int(arg.ToInteger())
		}

		sg, err := r.graph.ShortestPath(r.ctx, fromID, toID, maxHops)
		if err != nil {
			panic(r.runtime.NewGoError(err))
		}
		return r.marshalToJS(sg)
	})

	r.runtime.Set("graph_cypher", func(call goja.FunctionCall) goja.Value {
		cypher := call.Argument(0).String()

		var params map[string]any
		if arg := call.Argument(1); arg != nil && !goja.IsUndefined(arg) && !goja.IsNull(arg) {
			if err := r.runtime.ExportTo(arg, &params); err != nil {
				panic(r.runtime.NewGoError(fmt.Errorf("graph_cypher: invalid params: %w", err)))
			}
		}

		sg, err := r.graph.RunCypher(r.ctx, cypher, params)
		if err != nil {
			panic(r.runtime.NewGoError(err))
		}
		return r.marshalToJS(sg)
	})

	r.runtime.Set("graph_aggregate", func(call goja.FunctionCall) goja.Value {
		cypher := call.Argument(0).String()

		var params map[string]any
		if arg := call.Argument(1); arg != nil && !goja.IsUndefined(arg) && !goja.IsNull(arg) {
			if err := r.runtime.ExportTo(arg, &params); err != nil {
				panic(r.runtime.NewGoError(fmt.Errorf("graph_aggregate: invalid params: %w", err)))
			}
		}

		rows, err := r.graph.Aggregate(r.ctx, cypher, params)
		if err != nil {
			panic(r.runtime.NewGoError(err))
		}
		return r.marshalToJS(rows)
	})
}

func (r *REPL) registerLLMQuery() {
	r.runtime.Set("llm_query", func(call goja.FunctionCall) goja.Value {
		prompt := call.Argument(0).String()

		response, err := r.subLLM.Complete(r.ctx, []Message{
			{Role: "user", Content: prompt},
		})
		if err != nil {
			panic(r.runtime.NewGoError(err))
		}
		return r.runtime.ToValue(response)
	})
}

func (r *REPL) registerFinal() {
	r.runtime.Set("FINAL", func(call goja.FunctionCall) goja.Value {
		r.FinalAnswer = call.Argument(0).String()
		r.Done = true
		return goja.Undefined()
	})

	r.runtime.Set("FINAL_VAR", func(call goja.FunctionCall) goja.Value {
		varName := call.Argument(0).String()
		val := r.runtime.Get(varName)
		if val == nil || goja.IsUndefined(val) {
			r.FinalAnswer = varName // graceful fallback per spec
			r.Done = true
			return goja.Undefined()
		}
		exported := val.Export()
		jsonBytes, err := json.Marshal(exported)
		if err != nil {
			r.FinalAnswer = fmt.Sprintf("%v", exported)
		} else {
			r.FinalAnswer = string(jsonBytes)
		}
		r.Done = true
		return goja.Undefined()
	})
}

// marshalToJS converts a Go value to a native JS object by marshaling to JSON
// and parsing it back through JSON.parse (safe, no eval).
func (r *REPL) marshalToJS(v any) goja.Value {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		panic(r.runtime.NewGoError(fmt.Errorf("marshal error: %w", err)))
	}
	jsonParse, ok := goja.AssertFunction(r.runtime.Get("JSON").ToObject(r.runtime).Get("parse"))
	if !ok {
		panic(r.runtime.NewGoError(fmt.Errorf("JSON.parse not available")))
	}
	val, err := jsonParse(goja.Undefined(), r.runtime.ToValue(string(jsonBytes)))
	if err != nil {
		panic(r.runtime.NewGoError(fmt.Errorf("JSON.parse error: %w", err)))
	}
	return val
}
