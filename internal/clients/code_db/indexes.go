package codedb

import (
	"context"
	"fmt"
	"strings"

	"github.com/tab58/go-ormql/pkg/cypher"
	"github.com/tab58/go-ormql/pkg/driver"
)

// isAlreadyIndexed returns true if the error indicates the attribute is already indexed.
func isAlreadyIndexed(err error) bool {
	return err != nil && strings.Contains(err.Error(), "already indexed")
}

// rangeIndexes defines property indexes for efficient MERGE/MATCH lookups.
var rangeIndexes = []cypher.Statement{
	{Query: "CREATE INDEX FOR (n:Repository) ON (n.name)"},
	{Query: "CREATE INDEX FOR (n:Folder) ON (n.path)"},
	{Query: "CREATE INDEX FOR (n:File) ON (n.path)"},
	{Query: "CREATE INDEX FOR (n:Function) ON (n.name)"},
	{Query: "CREATE INDEX FOR (n:Function) ON (n.path)"},
	{Query: "CREATE INDEX FOR (n:Class) ON (n.name)"},
	{Query: "CREATE INDEX FOR (n:Class) ON (n.path)"},
	{Query: "CREATE INDEX FOR (n:Module) ON (n.name)"},
	{Query: "CREATE INDEX FOR (n:Module) ON (n.path)"},
	{Query: "CREATE INDEX FOR (n:ExternalReference) ON (n.name)"},
	{Query: "CREATE INDEX FOR (n:ExternalReference) ON (n.importPath)"},
}

// createIndexes creates range indexes for graph nodes.
func createIndexes(ctx context.Context, drv driver.Driver) error {
	for _, stmt := range rangeIndexes {
		if _, err := drv.ExecuteWrite(ctx, stmt); err != nil && !isAlreadyIndexed(err) {
			return fmt.Errorf("failed to create range index: %w", err)
		}
	}
	return nil
}
