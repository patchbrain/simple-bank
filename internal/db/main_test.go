package db

import (
	"database/sql"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

var testQueries *Queries
var testDb *sql.DB

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:123456@localhost:5432/simple-bank?sslmode=disable"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())
	var err error
	// 连接数据库
	testDb, err = sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatalf("fail to connection the postgresql, error: %s", err.Error())
		return
	}
	testQueries = New(testDb)
	// 测试程序运行
	os.Exit(m.Run())
}
