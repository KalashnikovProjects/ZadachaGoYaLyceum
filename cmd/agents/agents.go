package main

import (
	"Zadacha/internal/api/agent"
	"github.com/xlab/closer"
)

func main() {
	agent.CreateAgents()
	closer.Hold()
}
