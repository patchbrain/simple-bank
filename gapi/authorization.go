package gapi

import (
	"context"
	"fmt"
	"github.com/patchbrain/simple-bank/token"
	"google.golang.org/grpc/metadata"
	"strings"
)

const (
	AuthorizationHeader = "authorization"
	BearTokenType       = "bearer"
)

func (s *Server) AuthorizeUser(ctx context.Context) (*token.Payload, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("fail to get metadata from context")
	}

	values := md.Get(AuthorizationHeader)
	if len(values) < 1 {
		return nil, fmt.Errorf("no authorizationHeader")
	}
	value := values[0]

	fields := strings.Fields(value)
	if len(fields) < 2 {
		return nil, fmt.Errorf("invalid authorization")
	}
	if strings.ToLower(fields[0]) != BearTokenType {
		return nil, fmt.Errorf("unsupported token type")
	}

	payload, err := s.TokenMaker.VerifyToken(fields[1])
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return payload, nil
}
