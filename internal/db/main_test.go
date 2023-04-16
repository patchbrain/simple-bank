package db

import (
	"database/sql"
	"github.com/patchbrain/simple-bank/internal/util"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

var testQueries *Queries
var testDb *sql.DB

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())
	var err error
	cfg, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatalf("fail to read config, error: %s", err.Error())
		return
	}
	// 连接数据库
	testDb, err = sql.Open(cfg.DbDriver, cfg.DbSource)
	if err != nil {
		log.Fatalf("fail to connection the postgresql, error: %s", err.Error())
		return
	}
	testQueries = New(testDb)
	// 测试程序运行
	os.Exit(m.Run())
}
