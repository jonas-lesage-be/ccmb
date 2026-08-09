package main

import (
	"ccmb/internal/config"
	"ccmb/internal/pipeline"
)

func main() {
	cfg := config.Load()
	pipeline.Execute(cfg)
}
