package main

import (
	"Zadacha/internal/agent"
	"github.com/xlab/closer"
)

func main() {
	agent.CreateAgents()
	closer.Hold()
}
