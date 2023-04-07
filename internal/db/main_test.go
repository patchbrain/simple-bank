package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var testQueries *Queries

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:123456@localhost:5432/simple-bank?sslmode=disable"
)

func TestMain(m *testing.M) {
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
