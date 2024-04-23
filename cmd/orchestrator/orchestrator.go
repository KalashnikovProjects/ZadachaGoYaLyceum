package main

import (
	"context"
	"github.com/KalashnikovProjects/ZadachaGoYaLyceum/internal/api"
	"github.com/xlab/closer"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	api.Run(ctx)
	closer.Hold()
	cancel()
}
