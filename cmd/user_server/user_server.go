package main

import (
	"Zadacha/internal/user_server"
	"github.com/xlab/closer"
)

func main() {
	user_server.Run()
	closer.Hold()
}
