package main

import (
	"github.com/KalashnikovProjects/ZadachaGoYaLyceum/internal/agent"
	"github.com/xlab/closer"
)

func main() {
	agent.ManagerAgent()
	closer.Hold()
}
