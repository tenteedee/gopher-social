package main

import (
	"fmt"
	"log"

	"github.com/tenteedee/gopher-social/internal/db"
	"github.com/tenteedee/gopher-social/internal/env"
	"github.com/tenteedee/gopher-social/internal/store"
)

func main() {
	env.Init()

	addr := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable", env.DbUser, env.DbPassword, env.DbHost, env.DbPort, env.DbName)
	conn, err := db.New(addr, 3, 3, "15m")
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	store := store.NewStorage(conn)

	db.Seed(store, conn)
}
