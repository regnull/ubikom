package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/globals"
	"github.com/regnull/ubikom/mail"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

const (
	testUser1 = "int-test-1"
	testUser2 = "int-test-2"
	testUser3 = "int-test-3"
)

type Args struct {
	URL              string
	Password         string
	LogLevel         string
	UseTLS           bool
	LookupServiceURL string
	Timeout          time.Duration
}

var tests = []struct {
	name string
	exec func(*Args) error
}{
	{"TestLoginLogout", TestLoginLogout},
	{"TestListMailboxes", TestListMailboxes},
	{"TestSendReceive", TestSendReceive},
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})

	args := &Args{}
	flag.StringVar(&args.URL, "url", "", "server url")
	flag.StringVar(&args.Password, "password", "", "test users password")
	flag.StringVar(&args.LogLevel, "log-level", "debug", "log level")
	flag.BoolVar(&args.UseTLS, "use-tls", true, "use TLS for connection")
	flag.StringVar(&args.LookupServiceURL, "lookup-url", globals.PublicLookupServiceURL, "lookup service URL")
	flag.DurationVar(&args.Timeout, "timeout", 2*time.Second, "connection timeout")
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

	for _, t := range tests {
		log.Info().Str("name", t.name).Msg("running...")
		err := t.exec(args)
		if err != nil {
			log.Error().Err(err).Msg("test FAILED")
		} else {
			log.Info().Str("name", t.name).Msg("success!")
		}
	}
}

func Connect(args *Args) (*client.Client, error) {
	var c *client.Client
	var err error
	if args.UseTLS {
		c, err = client.DialTLS(args.URL, nil)
	} else {
		c, err = client.Dial(args.URL)
	}
	return c, err
}

func TestLoginLogout(args *Args) error {
	c, err := Connect(args)
	if err != nil {
		return err
	}
	err = c.Login(testUser1, args.Password)
	if err != nil {
		return err
	}
	return c.Logout()
}

func TestListMailboxes(args *Args) error {
	c, err := Connect(args)
	if err != nil {
		return err
	}

	err = c.Login(testUser1, args.Password)
	if err != nil {
		return err
	}
	defer c.Logout()

	mailboxesChan := make(chan *imap.MailboxInfo, 100)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxesChan)
	}()

	var mailboxes []*imap.MailboxInfo
	for m := range mailboxesChan {
		mailboxes = append(mailboxes, m)
	}

	if err := <-done; err != nil {
		return err
	}

	if len(mailboxes) != 1 {
		return fmt.Errorf("expected one mailbox, got %d", len(mailboxes))
	}

	if mailboxes[0].Name != "INBOX" {
		return fmt.Errorf("expected INBOX, got %s", mailboxes[0].Name)
	}

	return nil
}

func TestSendReceive(args *Args) error {
	c, err := Connect(args)
	if err != nil {
		return err
	}

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(args.Timeout),
	}
	log.Info().Str("url", args.LookupServiceURL).Msg("connecting to lookup service")
	lookupConn, err := grpc.Dial(args.LookupServiceURL, opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to the lookup service")
	}
	defer lookupConn.Close()

	lookupClient := pb.NewLookupServiceClient(lookupConn)

	key1 := easyecc.NewPrivateKeyFromPassword([]byte(args.Password), util.Hash256([]byte(testUser1)))

	ctx := context.Background()
	email := mail.NewMessage(testUser2, testUser1, "integration testing", "This is test message 1")
	err = protoutil.SendMessage(ctx, key1, []byte(email), testUser1, testUser2, lookupClient)
	if err != nil {
		return err
	}

	err = c.Login(testUser2, args.Password)
	if err != nil {
		return err
	}
	defer c.Logout()

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		return err
	}

	// Read all the messages from the inbox.
	from := uint32(1)
	to := mbox.Messages
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messagesChan := make(chan *imap.Message, 100)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messagesChan)
	}()

	var messages []*imap.Message
	for msg := range messagesChan {
		messages = append(messages, msg)
	}

	if err := <-done; err != nil {
		return err
	}

	if len(messages) != 1 {
		return fmt.Errorf("expected one message, got %d", len(messages))
	}

	for i := range messages {
		seqset = new(imap.SeqSet)
		seqset.AddNum(uint32(i + 1))
		err = c.Store(seqset, imap.FormatFlagsOp(imap.AddFlags, true), []interface{}{imap.DeletedFlag}, nil)
		if err != nil {
			return fmt.Errorf("failed to mark message as deleted, %w", err)
		}
	}

	err = c.Expunge(nil)
	if err != nil {
		return fmt.Errorf("failed to expunge messages: %w", err)
	}

	return nil
}
