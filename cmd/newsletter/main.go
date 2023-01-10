package main

import (
	"flag"
	"os"

	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type CmdArgs struct {
	KeyLocation string
	Name        string
	MessageFile string
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	var args CmdArgs
	flag.StringVar(&args.KeyLocation, "key", "", "sender's key location")
	flag.StringVar(&args.Name, "name", "", "ubikom name of the sender")
	flag.StringVar(&args.MessageFile, "message", "", "message file")
	flag.Parse()

	assertStringFlagSet(args.KeyLocation, "key")
	assertStringFlagSet(args.Name, "name")
	assertStringFlagSet(args.MessageFile, "message")

	privateKey, err := util.LoadKey(args.KeyLocation)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load the private key")
	}

	_ = privateKey // To calm the compiler down.
}

func assertStringFlagSet(value string, name string) {
	if value == "" {
		log.Fatal().Str("name", name).Msg("mandatory flag must be set")
	}
}
