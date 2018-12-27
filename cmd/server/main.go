package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/bojand/ghz/web/config"
	"github.com/bojand/ghz/web/database"
	"github.com/bojand/ghz/web/router"
)

var (
	// set by goreleaser with -ldflags="-X main.version=..."
	version = "dev"
	cPath   = flag.String("config", "", "Path to the config file.")
	v       = flag.Bool("v", false, "Print the version.")
)

var usage = `Usage: ghz-web [options...]
Options:
  -config	Path to the config JSON file.
  -v  Print the version.
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(usage, runtime.NumCPU()))
	}

	flag.Parse()

	if *v {
		fmt.Println(version)
		os.Exit(0)
	}

	cfgPath := strings.TrimSpace(*cPath)

	conf, err := config.Read(cfgPath)
	if err != nil {
		panic(err)
	}

	db, err := database.New(conf.Database.Type, conf.Database.Connection)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	server, err := router.New(db, conf)
	if err != nil {
		panic(err)
	}

	router.PrintRoutes(server)

	hostPort := net.JoinHostPort("localhost", strconv.FormatUint(uint64(conf.Server.Port), 10))
	server.Logger.Fatal(server.Start(hostPort))
}
