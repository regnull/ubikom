package main

import (
	"flag"
	"os"

	"github.com/emersion/go-imap/client"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	testUser1 = "int-test-1"
)

type Args struct {
	URL      string
	Password string
	LogLevel string
	UseTLS   bool
}

var tests = []struct {
	name string
	exec func(*client.Client, *Args) error
}{
	{"TestLoginLogout", TestLoginLogout},
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})

	args := &Args{}
	flag.StringVar(&args.URL, "url", "", "server url")
	flag.StringVar(&args.Password, "password", "", "test users password")
	flag.StringVar(&args.LogLevel, "log-level", "debug", "log level")
	flag.BoolVar(&args.UseTLS, "use-tls", true, "use TLS for connection")
	flag.Parse()

	if args.URL == "" {
		log.Fatal().Msg("server url must be specified")
	}

	if args.Password == "" {
		log.Fatal().Msg("password must be specified")
	}

	// Set the log level.
	logLevel, err := zerolog.ParseLevel(args.LogLevel)
	if err != nil {
		log.Fatal().Str("level", args.LogLevel).Msg("invalid log level")
	}

	zerolog.SetGlobalLevel(logLevel)

	var c *client.Client
	if args.UseTLS {
		c, err = client.DialTLS(args.URL, nil)
	} else {
		c, err = client.Dial(args.URL)
	}
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to the server")
	}
	for _, t := range tests {
		log.Info().Str("name", t.name).Msg("running...")
		err := t.exec(c, args)
		if err != nil {
			log.Error().Err(err).Msg("test FAILED")
		} else {
			log.Info().Str("name", t.name).Msg("success!")
		}
	}
}

func TestLoginLogout(c *client.Client, args *Args) error {
	err := c.Login(testUser1, args.Password)
	if err != nil {
		return err
	}
	return c.Logout()
}
