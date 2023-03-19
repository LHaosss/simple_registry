package main

import (
	registry "simple_registry"
)

func main() {
	opts := registry.InitOptions()
	registry.StartService(opts)
}
