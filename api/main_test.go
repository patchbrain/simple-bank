package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/patchbrain/simple-bank/db/sqlc"
	"github.com/patchbrain/simple-bank/util"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func newTestServer(t *testing.T, store db.Store) *Server {
	config := util.Config{
		TokenSecretKey: "qweasdzxqweasdzxqweasdzxqweasdzx",
		TokenDuration:  time.Minute,
	}

	s, err := NewServer(config, store)
	require.NoError(t, err)

	return s
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	// 测试程序运行
	os.Exit(m.Run())
}
