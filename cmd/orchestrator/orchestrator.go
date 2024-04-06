package main

import (
	"Zadacha/internal/api"
	"github.com/xlab/closer"
)

func main() {
	api.Run()
	closer.Hold()
}
