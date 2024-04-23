package main

import (
	"context"
	"github.com/KalashnikovProjects/ZadachaGoYaLyceum/internal/agent"
	"github.com/xlab/closer"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	agent.ManagerAgent(ctx)
	closer.Hold()
	cancel()
}
