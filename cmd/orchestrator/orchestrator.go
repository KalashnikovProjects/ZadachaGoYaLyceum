package main

import (
	"Zadacha/internal/api/orchestrator"
	"github.com/xlab/closer"
)

func main() {
	orchestrator.Run()
	closer.Hold()
}
