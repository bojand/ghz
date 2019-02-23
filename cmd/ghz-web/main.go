package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/bojand/ghz/web/api"
	"github.com/bojand/ghz/web/config"
	"github.com/bojand/ghz/web/database"
	"github.com/bojand/ghz/web/router"
)

var (
	// set by goreleaser with -ldflags="-X main.version=..."
	version = "dev"
	beta    = true
	date    = "unknown"
	cPath   = flag.String("config", "", "Path to the config file.")
	v       = flag.Bool("v", false, "Print the version.")
)

var usage = `Usage: ghz-web [options...]
Options:
  -config	Path to the config JSON file.
  -v  Print the version.
`

func main() {
	// fix version
	if version != "dev" && beta {
		version = version + "-beta"
	}

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}

	flag.Parse()

	if *v {
		fmt.Println(version)
		os.Exit(0)
	}

	cfgPath := strings.TrimSpace(*cPath)

	conf, err := config.Read(cfgPath)
	if err != nil {
		handleError(err)
	}

	db, err := database.New(conf.Database.Type, conf.Database.Connection, conf.Log.Level == "debug")
	if err != nil {
		handleError(err)
	}
	defer func() {
		handleError(db.Close())
	}()

	info := &api.ApplicationInfo{
		Version:   version,
		BuildDate: date,
		GOVersion: runtime.Version(),
		StartTime: time.Now(),
	}

	server, err := router.New(db, info, conf)
	if err != nil {
		handleError(err)
	}

	router.PrintRoutes(server)

	hostPort := net.JoinHostPort("", strconv.FormatUint(uint64(conf.Server.Port), 10))
	server.Logger.Fatal(server.Start(hostPort))
}

func handleError(err error) {
	if err != nil {
		if errString := err.Error(); errString != "" {
			fmt.Fprintln(os.Stderr, errString)
		}
		os.Exit(1)
	}
}
