package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/patchbrain/simple-bank/db/sqlc"
	"github.com/patchbrain/simple-bank/token"
	"github.com/patchbrain/simple-bank/util"
)

type Server struct {
	Store      db.Store
	Router     *gin.Engine
	TokenMaker token.Maker
	Config     util.Config
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	maker, err := token.NewPasetoMaker(config.TokenSecretKey)
	if err != nil {
		return nil, fmt.Errorf("can't create a TokenMaker: %w", err)
	}

	s := new(Server)
	s.Store = store
	s.TokenMaker = maker
	s.Config = config

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	route(s)

	return s, nil
}

func route(s *Server) {
	r := gin.Default()

	r.POST("/users", s.createUser)
	r.POST("/users/login", s.loginUser)

	authRoute := r.Group("/").Use(authMiddleware(s.TokenMaker))
	authRoute.POST("/accounts", s.createAccount)
	authRoute.GET("/accounts/:id", s.getAccount)
	authRoute.GET("/accounts", s.listAccounts)
	authRoute.POST("/transfers", s.createTransfer)

	s.Router = r
}

func (s *Server) Start(addr string) error {
	return s.Router.Run(addr)
}

func errorResponse(err error) gin.H {
	return map[string]any{"error": err.Error()}
}
