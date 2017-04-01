package main

import (
	"./app"
)

func main() {
	app.Core.Version = "0.0.2b"
	app.Core.ConfigFile = "./config.yaml"
	app.Core.LoadConfig()
	app.Core.Run()
}
