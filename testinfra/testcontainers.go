package testinfra

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type FalkorDBContainer struct {
	Container testcontainers.Container
	Host      string
	Port      string
}

func SetupFalkorDB(ctx context.Context) (FalkorDBContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "falkordb/falkordb:latest",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return FalkorDBContainer{}, fmt.Errorf("starting falkordb container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return FalkorDBContainer{}, fmt.Errorf("getting container host: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, "6379")
	if err != nil {
		return FalkorDBContainer{}, fmt.Errorf("getting mapped port: %w", err)
	}

	return FalkorDBContainer{
		Container: container,
		Host:      host,
		Port:      mappedPort.Port(),
	}, nil
}

func (f FalkorDBContainer) ConnectionAddress() string {
	return fmt.Sprintf("%s:%s", f.Host, f.Port)
}

func (f FalkorDBContainer) Teardown(ctx context.Context) error {
	if f.Container != nil {
		return f.Container.Terminate(ctx)
	}
	return nil
}
