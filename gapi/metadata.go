package gapi

import (
	"context"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	grpcGatewayUserAgentHeader = "grpcgateway-user-agent"
	xForwardedForHeader        = "x-forwarded-for"
	grpcUserAgentHeader        = "user-agent"
)

type Metadata struct {
	ClientIP  string
	UserAgent string
}

func (s *Server) extractMetadata(ctx context.Context) Metadata {
	mtdt := Metadata{}

	// 将请求的元信息提取到ctx中
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		// 再获取ctx中的信息到Metadata中
		// 区分rpc与http请求，因为header可能不同
		//log.Printf("md: %v+\n", md)
		if userAgents := md.Get(grpcGatewayUserAgentHeader); len(userAgents) > 0 {
			mtdt.UserAgent = userAgents[0]
		}
		if clientIPs := md.Get(xForwardedForHeader); len(clientIPs) > 0 {
			mtdt.ClientIP = clientIPs[0]
		}
		if grpcUserAgents := md.Get(grpcUserAgentHeader); len(grpcUserAgents) > 0 {
			mtdt.UserAgent = grpcUserAgents[0]
		}
	}

	// rpc请求的clientIP通过peer包获取
	if peer, ok := peer.FromContext(ctx); ok {
		mtdt.ClientIP = peer.Addr.String()
	}
	return mtdt
}
