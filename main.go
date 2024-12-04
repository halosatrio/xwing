package main

import (
	"github.com/halosatrio/xwing/config"
	"github.com/halosatrio/xwing/handlers"
)

func main() {
	config.LoadEnv()

	db := config.ConnectDB()
	defer db.Close()

	r := handlers.SetupRouter(db)
	r.Run(":8080")
}
