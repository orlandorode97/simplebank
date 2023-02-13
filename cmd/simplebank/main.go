package main

import (
	"database/sql"
	"flag"
	"log"
	"net"
	"time"

	_ "github.com/lib/pq"
	simplebankgrpc "github.com/orlandorode97/simple-bank/api/grpc"
	simplebankhttp "github.com/orlandorode97/simple-bank/api/http"
	"github.com/orlandorode97/simple-bank/config"
	simplebankpb "github.com/orlandorode97/simple-bank/generated/simplebank"
	"github.com/orlandorode97/simple-bank/store"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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

	conf, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal(err)
	}
	conn, err := sql.Open(conf.DBDriver, conf.DBSource)
	if err != nil {
		log.Fatal(err)
	}

	store := store.NewSimpleBankDB(conn)

	httpServer, err := simplebankhttp.NewServer(conf, store)
	if err != nil {
		log.Fatalf("unable to create http server: %v", err)
	}

	grpcServer, err := simplebankgrpc.NewServer(conf, store)
	if err != nil {
		log.Fatalf("unable to create grpc server: %v", err)
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(grpcServer.UnaryInterceptor()),
	}
	server := grpc.NewServer(opts...)
	reflection.Register(server)
	simplebankpb.RegisterSimplebankServiceServer(server, grpcServer)

	tcpConn, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatal(err)
	}

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
