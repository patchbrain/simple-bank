package gapi

import (
	"context"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"time"
)

func GrpcLogger(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	startTime := time.Now()
	resp, err = handler(ctx, req)
	duration := time.Since(startTime)

	logger := log.Info()
	if err != nil {
		logger = log.Err(err)
	}

	statusCode := codes.Unknown
	status, ok := status.FromError(err)
	if ok {
		statusCode = status.Code()
	}

	logger.Str("protocol", "grpc").
		Dur("duration", duration).
		Int("status_code", int(statusCode)).
		Str("status_info", statusCode.String()).
		Str("method", info.FullMethod).
		Msg("received a rpc request")

	return resp, err
}

type HttpResponseWriter struct {
	http.ResponseWriter
	StatusCode int
	Body       []byte
}

func (h *HttpResponseWriter) WriteHeader(code int) {
	h.StatusCode = code
	h.ResponseWriter.WriteHeader(code)
}

func (h *HttpResponseWriter) Write(data []byte) (int, error) {
	h.Body = data
	return h.ResponseWriter.Write(data)
}

func HttpLogger(handler http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rw := &HttpResponseWriter{ResponseWriter: w}
		rw.StatusCode = http.StatusOK

		startTime := time.Now()
		handler.ServeHTTP(rw, req)
		duration := time.Since(startTime)

		logger := log.Info()
		if rw.StatusCode != http.StatusOK {
			logger = log.Error()
			logger.Bytes("body", rw.Body)
		}
		logger.Str("protocol", "http").
			Dur("duration", duration).
			//Int("status_code", int(statusCode)).
			//Str("status_info", statusCode.String()).
			Str("method", req.Method).
			Str("path", req.RequestURI).
			Msg("received a http request")
	})
}
