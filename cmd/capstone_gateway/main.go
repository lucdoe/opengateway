package main

import "github.com/lucdoe/capstone_gateway/internal/app/databases"

func main() {
	// init server here
	databases.InitializeRedis()
}
