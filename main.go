package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/patchbrain/simple-bank/api"
	"github.com/patchbrain/simple-bank/internal/db"
	"github.com/patchbrain/simple-bank/internal/util"
	"log"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	var err error
	cfg, err := util.LoadConfig(".")
	if err != nil {
		log.Fatalf("fail to read config, error: %s", err.Error())
		return
	}

	// 连接数据库
	conn, err := sql.Open(cfg.DbDriver, cfg.DbSource)
	if err != nil {
		log.Fatalf("fail to connection the postgresql, error: %s", err.Error())
		return
	}

	q := db.NewStore(conn)
	server := api.NewServer(q)

	if err := server.Start(cfg.ServerAddress); err != nil {
		log.Fatalf("fail to start server: %s", err.Error())
		return
	}
}