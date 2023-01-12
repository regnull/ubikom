package main

import (
	"bufio"
	"context"
	"flag"
	"io/ioutil"
	"os"
	"strings"

	"github.com/regnull/ubikom/globals"
	"github.com/regnull/ubikom/lookup"
	"github.com/regnull/ubikom/mail"
	"github.com/regnull/ubikom/protoutil"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type CmdArgs struct {
	KeyLocation           string
	Name                  string
	MessageFile           string
	RecipientsFile        string
	Network               string
	NodeUrl               string
	ProjectId             string
	ContractAddress       string
	Subject               string
	LegacyLookupServerUrl string
	legacyNodeUrl         string
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	var args CmdArgs
	flag.StringVar(&args.KeyLocation, "key", "", "sender's key location")
	flag.StringVar(&args.Name, "name", "", "ubikom name of the sender")
	flag.StringVar(&args.MessageFile, "message", "", "message file")
	flag.StringVar(&args.RecipientsFile, "recipients", "", "recipients file")
	flag.StringVar(&args.Network, "network", "main", "Ethereum network")
	flag.StringVar(&args.NodeUrl, "node-url", "", "Ethereum node URL")
	flag.StringVar(&args.ProjectId, "project-id", "", "Infura project ID")
	flag.StringVar(&args.ContractAddress, "contract-address", "", "contract address")
	flag.StringVar(&args.Subject, "subject", "", "email subject")
	flag.StringVar(&args.LegacyLookupServerUrl, "legacy-lookup-url", globals.PublicLookupServiceURL, "legacy lookup server URL")
	flag.StringVar(&args.legacyNodeUrl, "legacy-node-url", globals.BlockchainNodeURL, "legacy blockchain node URL")
	flag.Parse()

	assertStringFlagSet(args.KeyLocation, "key")
	assertStringFlagSet(args.Name, "name")
	assertStringFlagSet(args.MessageFile, "message")
	assertStringFlagSet(args.RecipientsFile, "recipients")
	assertStringFlagSet(args.Subject, "subject")

	privateKey, err := util.LoadKey(args.KeyLocation)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load the private key")
	}

	lookupService, cleanup, err := lookup.Get(args.Network, args.ProjectId, args.ContractAddress,
		args.LegacyLookupServerUrl, args.legacyNodeUrl, false)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get lookup service")
	}
	if cleanup != nil {
		defer cleanup()
	}

	content, err := ioutil.ReadFile(args.MessageFile)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read the message file")
	}

	recipientsFile, err := os.Open(args.RecipientsFile)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read the recipients file")
	}
	defer recipientsFile.Close()

	ctx := context.Background()

	scanner := bufio.NewScanner(recipientsFile)
	for scanner.Scan() {
		name := scanner.Text()
		name = util.StripDomainName(name)
		name = strings.ToLower(name)

		mailMessage := mail.NewMessage(name, args.Name, args.Subject, string(content))

		log.Info().Str("recipient", name).Msg("sending message")
		err = protoutil.SendEmail(ctx, privateKey, []byte(mailMessage), args.Name, name, lookupService)
		if err != nil {
			log.Fatal().Err(err).Msg("error sending message")
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal().Err(err).Msg("error scanning recipients file")
	}

	log.Info().Msg("Done!")
}

func assertStringFlagSet(value string, name string) {
	if value == "" {
		log.Fatal().Str("name", name).Msg("mandatory flag must be set")
	}
}
