package main

import (
	"context"
	"database/sql"
	_ "github.com/golang/mock/mockgen/model"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/patchbrain/simple-bank/api"
	"github.com/patchbrain/simple-bank/db/sqlc"
	"github.com/patchbrain/simple-bank/gapi"
	"github.com/patchbrain/simple-bank/pb"
	"github.com/patchbrain/simple-bank/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	var err error
	cfg, err := util.LoadConfig(".")
	if err != nil {
		log.Fatalf("fail to read config, error: %s", err.Error())
		return
	}

	// 连接数据库
	conn, err := sql.Open(cfg.DbDriver, cfg.DbSource)
	if err != nil {
		log.Fatalf("fail to connection the postgresql, error: %s", err.Error())
		return
	}

	q := db.NewStore(conn)

	go runGatewayServer(cfg, q)
	runGrpcServer(cfg, q)
}

func runGrpcServer(config util.Config, db db.Store) {
	grpcServer := grpc.NewServer()            // 创建一个grpc服务器，用于监听端口、地址等
	server, err := gapi.NewServer(config, db) // 创建一个实现服务接口的服务器
	if err != nil {
		log.Fatal("cannot create server:", err)
	}
	pb.RegisterSimpleBankServer(grpcServer, server) // 将server注册到grpcServer，从而使得grpcServer可以调用server中实现的方法来服务
	reflection.Register(grpcServer)                 // 将服务的方法名、参数等信息公开，从而使客户端能够动态构建调用参数与响应来进行服务调用

	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal("cannot create listener:", err)
	}

	log.Printf("start grpc server at %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("fail to run grpcServer:", err)
	}
}

func runGatewayServer(config util.Config, db db.Store) {
	server, err := gapi.NewServer(config, db) // 创建一个实现服务接口的服务器
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	// 传入的参数可以使得转换出的JSON数据是.proto文件中的驼峰式命名
	grpcMux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal("fail to register handler server: ", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot create listener:", err)
	}

	log.Printf("start HTTP gateway server at %s", listener.Addr().String())
	err = http.Serve(listener, mux)
	if err != nil {
		log.Fatal("fail to run HTTP gateway server:", err)
	}
}

func runGinServer(config util.Config, db db.Store) {
	server, err := api.NewServer(config, db)
	if err != nil {
		log.Fatalf("fail to start the server, error: %s", err.Error())
		return
	}

	if err := server.Start(config.HTTPServerAddress); err != nil {
		log.Fatalf("fail to start server: %s", err.Error())
		return
	}
}
