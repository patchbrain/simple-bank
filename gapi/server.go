package gapi

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/patchbrain/simple-bank/db/sqlc"
	"github.com/patchbrain/simple-bank/pb"
	"github.com/patchbrain/simple-bank/token"
	"github.com/patchbrain/simple-bank/util"
)

type Server struct {
	pb.UnimplementedSimpleBankServer
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

	return s, nil
}

func (s *Server) Start(addr string) error {
	return s.Router.Run(addr)
}

func errorResponse(err error) gin.H {
	return map[string]any{"error": err.Error()}
}
