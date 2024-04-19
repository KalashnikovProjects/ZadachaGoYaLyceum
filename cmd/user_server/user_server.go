package main

import (
	"github.com/KalashnikovProjects/ZadachaGoYaLyceum/internal/user_server"
	"github.com/xlab/closer"
)

func main() {
	user_server.Run()
	closer.Hold()
}
