package main

import (
	"database/sql"
	"flag"
	"log"
	"net"
	"time"

	"github.com/hibiken/asynq"
	_ "github.com/lib/pq"
	simplebankgrpc "github.com/orlandorode97/simple-bank/api/grpc"
	simplebankhttp "github.com/orlandorode97/simple-bank/api/http"
	"github.com/orlandorode97/simple-bank/config"
	simplebankpb "github.com/orlandorode97/simple-bank/generated/simplebank"
	"github.com/orlandorode97/simple-bank/mail"
	"github.com/orlandorode97/simple-bank/store"
	"github.com/orlandorode97/simple-bank/workers"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func flags() {
	flag.String("http-addr", ":8081", "http address") // declare needed flags
	flag.String("grpc-addr", ":8082", "grpc address")
	flag.Duration("grpc-timeout", 5*time.Second, "grpc timeout")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine) // add standard library flags set to pflags of viper
	pflag.Parse()                                    // Parsing pflag set
	viper.BindPFlags(pflag.CommandLine)              // Binding pflag from go flags
}

func main() {
	flags()
	httpAddr := viper.GetString("http-addr")
	grpcAddr := viper.GetString("grpc-addr")

	logger := zap.NewExample()
	suggar := logger.Sugar()

	defer logger.Sync()
	conf, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := sql.Open(conf.DBDriver, conf.DBSource)
	if err != nil {
		log.Fatal(err)
	}

	store := store.NewSimpleBankDB(conn)

	redisOpt := asynq.RedisClientOpt{
		Addr: conf.RedisAddr,
	}

	//Email sender
	emailSender := mail.NewSender(conf.GmailName, conf.GmailAddress, conf.GmailPassword)

	// Task distributor and processor
	taskDistributor := workers.NewRedisTaskDistributor(redisOpt, suggar)
	taskProcessor := workers.NewRedistTaskProcessor(redisOpt, store, suggar, emailSender)

	httpServer, err := simplebankhttp.NewServer(conf, store)
	if err != nil {
		log.Fatalf("unable to create http server: %v", err)
	}

	grpcServer, err := simplebankgrpc.NewServer(conf, store, suggar, taskDistributor)
	if err != nil {
		log.Fatalf("unable to create grpc server: %v", err)
	}

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			grpcServer.AuthInterceptor(),
			grpcServer.LoggerInterceptor(),
		),
	}
	server := grpc.NewServer(opts...)
	reflection.Register(server)
	simplebankpb.RegisterSimplebankServiceServer(server, grpcServer)

	tcpConn, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		log.Printf("Serving task processor")
		err := taskProcessor.Start()
		if err != nil {
			log.Fatalf("unable to start task processor: %v", err)
		}
	}()

	go func() {
		log.Printf("Serving grpc server: %v", grpcAddr)
		err := server.Serve(tcpConn)
		if err != nil {
			log.Fatalf("unable to start grpc server: %v", err)
		}
	}()

	log.Printf("Serving http server: %v", httpAddr)
	if err := httpServer.Listen(httpAddr); err != nil {
		log.Fatalf("unable to start http server: %v", err)
	}
}
