package main

import (
	"github.com/danmuck/dps_http/server"
	"github.com/danmuck/dps_http/users"
	"github.com/danmuck/dps_lib/logs"
)

var CONFIG = "config.toml"
var SERVER *server.HTTPServer

func ConfigureServices() {
	users := users.Configure(SERVER.Mongo)
	if users == nil {
		logs.Fatal("Failed to initialize UserService")
		return
	}
	logs.Dev("UserService initialized successfully")
	users.Up(SERVER.Router())
}

func main() {
	SERVER = server.NewHTTPServer()
	logs.Dev("Starting server with configuration: %+v", SERVER)

	ConfigureServices()
	err := SERVER.Start()
	if err != nil {
		logs.Err("Error starting server: %v", err)
		return
	}
}
