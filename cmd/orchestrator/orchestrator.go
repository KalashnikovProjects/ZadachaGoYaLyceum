package main

import (
	"github.com/KalashnikovProjects/ZadachaGoYaLyceum/internal/api"
	"github.com/xlab/closer"
)

func main() {
	api.Run()
	closer.Hold()
}
