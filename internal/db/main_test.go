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

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:123456@localhost:5432/simple-bank?sslmode=disable"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())
	// 连接数据库
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatalf("fail to connection the postgresql, error: %s", err.Error())
		return
	}
	testQueries = New(conn)
	// 测试程序运行
	os.Exit(m.Run())
}
