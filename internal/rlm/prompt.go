package rlm

import (
	_ "embed"
	"fmt"
	"strings"
	"text/template"
)

//go:embed prompts/system_prompt.md.tmpl
var systemPrompt string

var systemPromptTemplate = template.Must(template.New("systemPrompt").Parse(systemPrompt))

// BuildSystemPrompt constructs the root LLM's system prompt from graph metadata.
func BuildSystemPrompt(meta Metadata) (string, error) {
	var b strings.Builder

	err := systemPromptTemplate.Execute(&b, formatMetadata(meta))
	if err != nil {
		return "", fmt.Errorf("failed to execute system prompt template: %w", err)
	}
	return b.String(), nil
}

func formatMetadata(meta Metadata) string {
	var b strings.Builder

	b.WriteString("Knowledge Graph Metadata:\n")
	fmt.Fprintf(&b, "  Total nodes: %d\n", meta.TotalNodes)
	fmt.Fprintf(&b, "  Total edges: %d\n", meta.TotalEdges)

	b.WriteString("  Node labels:\n")
	for label, count := range meta.NodeLabels {
		fmt.Fprintf(&b, "    %s: %d\n", label, count)
	}

	b.WriteString("  Edge types:\n")
	for edgeType, count := range meta.EdgeTypes {
		fmt.Fprintf(&b, "    %s: %d\n", edgeType, count)
	}

	if len(meta.SampleNodes) > 0 {
		b.WriteString("  Sample nodes:\n")
		for _, n := range meta.SampleNodes {
			fmt.Fprintf(&b, "    [%s] labels=%v props=%v\n", n.ID, n.Labels, n.Properties)
		}
	}

	if len(meta.SampleEdges) > 0 {
		b.WriteString("  Sample edges:\n")
		for _, e := range meta.SampleEdges {
			fmt.Fprintf(&b, "    (%s)-[%s]->(%s) props=%v\n", e.SourceID, e.Type, e.TargetID, e.Properties)
		}
	}

	return b.String()
}
