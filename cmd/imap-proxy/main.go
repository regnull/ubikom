package main

import (
	"flag"
	"fmt"
	"os"

	proxy "github.com/emersion/go-imap-proxy"
	"github.com/emersion/go-imap/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type CmdArgs struct {
	TargetURL string
	LogLevel  string
	Port      int
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})

	args := &CmdArgs{}
	flag.IntVar(&args.Port, "port", 1143, "IMAP server port")
	flag.StringVar(&args.TargetURL, "target-url", "", "target URL to forward IMAP calls")
	flag.StringVar(&args.LogLevel, "log-level", "debug", "log level")
	flag.Parse()

	if args.TargetURL == "" {
		log.Fatal().Msg("--target-url must be specified")
	}

	// Set the log level.
	logLevel, err := zerolog.ParseLevel(args.LogLevel)
	if err != nil {
		log.Fatal().Str("level", args.LogLevel).Msg("invalid log level")
	}
	zerolog.SetGlobalLevel(logLevel)

	be := proxy.NewTLS(args.TargetURL, nil)

	// Create a new server
	s := server.New(be)
	s.Debug = os.Stderr
	s.Addr = fmt.Sprintf(":%d", args.Port)
	// Since we will use this server for testing only, we can allow plain text
	// authentication over unencrypted connections
	s.AllowInsecureAuth = true

	log.Info().Int("port", args.Port).Msg("Starting IMAP server")
	if err := s.ListenAndServe(); err != nil {
		log.Fatal().Err(err).Msg("server exited")
	}
}
