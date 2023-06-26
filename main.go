package main

import (
	"context"
	"database/sql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang/mock/mockgen/model"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hibiken/asynq"
	_ "github.com/lib/pq"
	"github.com/patchbrain/simple-bank/api"
	"github.com/patchbrain/simple-bank/db/sqlc"
	_ "github.com/patchbrain/simple-bank/doc/statik"
	"github.com/patchbrain/simple-bank/gapi"
	"github.com/patchbrain/simple-bank/mail"
	"github.com/patchbrain/simple-bank/pb"
	"github.com/patchbrain/simple-bank/util"
	"github.com/patchbrain/simple-bank/worker"
	"github.com/rakyll/statik/fs"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	var err error
	cfg, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("fail to read config")
		return
	}

	if cfg.Environment == "development" {
		// 开发模式下，日志以易读的方式打印
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// 连接数据库
	conn, err := sql.Open(cfg.DbDriver, cfg.DbSource)
	if err != nil {
		log.Fatal().Err(err).Msg("fail to connection the postgresql")
	}

	// 迁移数据库
	migration, err := migrate.New(cfg.MigrateUrl, cfg.DbSource)
	if err != nil {
		log.Fatal().Err(err).Msg("fail to create migration")
	}

	err = migration.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatal().Err(err).Msg("fail to migrate up")
		return
	}

	log.Info().Msg("successfully migrated.")

	// 创建 distributor
	redisOpt := asynq.RedisClientOpt{
		Addr: cfg.RedisAddress,
	}
	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

	q := db.NewStore(conn)

	// Start Processor
	go runProcessor(cfg, redisOpt, q)

	go runGatewayServer(cfg, q, taskDistributor)
	runGrpcServer(cfg, q, taskDistributor)
}

func runProcessor(cfg util.Config, opt asynq.RedisConnOpt, store db.Store) {
	sender := mail.NewWangYiEmailSender(cfg.FromEmailAddress, cfg.FromEmailPassword, "SimpleBank")
	processor := worker.NewRedisTaskProcessor(opt, store, sender)
	log.Info().Msg("start processor")
	err := processor.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("fail to start processor")
	}
}

func runGrpcServer(config util.Config, db db.Store, distributor worker.TaskDistributor) {
	server, err := gapi.NewServer(config, db, distributor) // 创建一个实现服务接口的服务器
	if err != nil {
		log.Err(err)
		log.Fatal().Msg("cannot create server")
	}

	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcServer := grpc.NewServer(grpcLogger) // 创建一个grpc服务器，用于监听端口、地址等

	pb.RegisterSimpleBankServer(grpcServer, server) // 将server注册到grpcServer，从而使得grpcServer可以调用server中实现的方法来服务
	reflection.Register(grpcServer)                 // 将服务的方法名、参数等信息公开，从而使客户端能够动态构建调用参数与响应来进行服务调用

	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Err(err)
		log.Fatal().Msg("cannot create listener")
	}

	log.Printf("start grpc server at %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Err(err)
		log.Fatal().Msg("fail to run grpcServer")
	}
}

func runGatewayServer(config util.Config, db db.Store, distributor worker.TaskDistributor) {
	server, err := gapi.NewServer(config, db, distributor) // 创建一个实现服务接口的服务器
	if err != nil {
		log.Err(err)
		log.Fatal().Msg("cannot create server")
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
		log.Err(err)
		log.Fatal().Msg("fail to register handler server")
	}

	mux := http.NewServeMux()
	mux.Handle("/", gapi.HttpLogger(grpcMux))

	// 把swagger静态文件放在内存中
	statikFS, err := fs.New()
	if err != nil {
		log.Err(err)
		log.Fatal().Msg("fail to create statik file server")
	}
	swaggerFs := http.FileServer(statikFS)
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", swaggerFs))

	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Err(err)
		log.Fatal().Msg("cannot create listener")
	}

	log.Printf("start HTTP gateway server at %s", listener.Addr().String())
	err = http.Serve(listener, mux)
	if err != nil {
		log.Err(err)
		log.Fatal().Msg("fail to run HTTP gateway server")
	}
}

func runGinServer(config util.Config, db db.Store) {
	server, err := api.NewServer(config, db)
	if err != nil {
		log.Err(err)
		log.Fatal().Msg("fail to start the server")
		return
	}

	if err = server.Start(config.HTTPServerAddress); err != nil {
		log.Err(err)
		log.Fatal().Msg("fail to start server")
		return
	}
}
