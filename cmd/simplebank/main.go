package main

import (
	"database/sql"
	"flag"
	"log"

	_ "github.com/lib/pq"
	"github.com/orlandorode97/simple-bank/api"
	"github.com/orlandorode97/simple-bank/config"
	"github.com/orlandorode97/simple-bank/store"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func flags() {
	flag.String("http-addr", ":8081", "http address") // declare needed flags

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine) // add standard library flags set to pflags of viper
	pflag.Parse()                                    // Parsing pflag set
	viper.BindPFlags(pflag.CommandLine)              // Binding pflag from go flags
}

func main() {
	flags()
	httpAddr := viper.GetString("http-addr")

	conf, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal(err)
	}
	conn, err := sql.Open(conf.DBDriver, conf.DBSource)
	if err != nil {
		log.Fatal(err)
	}

	store := store.NewSimpleBankDB(conn)

	server, err := api.NewServer(conf, store)
	if err != nil {
		log.Fatalf("unable to create server: %v", err)
	}

	if err := server.Listen(httpAddr); err != nil {
		log.Fatalf("unable to start server: %v", err)
	}
}
