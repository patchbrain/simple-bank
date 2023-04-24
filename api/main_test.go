package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/patchbrain/simple-bank/db/sqlc"
	"github.com/patchbrain/simple-bank/util"
	"log"
	"os"
	"testing"
	"time"
)

func newTestServer(store db.Store) *Server {
	config := util.Config{
		TokenSecretKey: "qweasdzxqweasdzxqweasdzxqweasdzx",
		TokenDuration:  time.Minute,
	}

	s, err := NewServer(config, store)
	if err != nil {
		log.Fatalf("fail to start the test server, err: %s\n", err.Error())
	}

	return s
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	// 测试程序运行
	os.Exit(m.Run())
}
