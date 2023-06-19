package main

import (
	"auth-service/routes"
)

func main() {
	routes.CreateRouter()
	routes.InitializeRoute()
	routes.ServerStart()
}
