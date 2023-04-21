package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/patchbrain/simple-bank/db/sqlc"
)

type Server struct {
	Store  db.Store
	Router *gin.Engine
}

func NewServer(store db.Store) *Server {
	s := new(Server)
	s.Store = store
	r := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	r.POST("/accounts", s.createAccount)
	r.GET("/accounts/:id", s.getAccount)
	r.GET("/accounts", s.listAccounts)
	r.POST("/transfers", s.createTransfer)

	s.Router = r
	return s
}

func (s *Server) Start(addr string) error {
	return s.Router.Run(addr)
}

func errorResponse(err error) gin.H {
	return map[string]any{"error": err.Error()}
}
