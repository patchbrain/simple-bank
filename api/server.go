package api

import (
	"github.com/gin-gonic/gin"
	"github.com/patchbrain/simple-bank/internal/db"
)

type Server struct {
	Store  *db.Store
	Router *gin.Engine
}

func NewServer(store *db.Store) *Server {
	s := new(Server)
	s.Store = store
	r := gin.Default()

	r.POST("/accounts", s.createAccount)
	r.GET("/accounts/:id", s.getAccount)
	r.GET("/accounts", s.listAccounts)

	s.Router = r
	return s
}

func (s *Server) Start(addr string) error {
	return s.Router.Run(addr)
}

func errorResponse(err error) gin.H {
	return map[string]any{"error": err.Error()}
}
