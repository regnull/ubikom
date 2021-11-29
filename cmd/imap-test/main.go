package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
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
	{"TestAppend", TestAppend},
	{"TestDeleteMessage", TestDeleteMessage},
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})

	args := &Args{}
	flag.StringVar(&args.URL, "url", "alpha.ubikom.cc:1993", "server url")
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

	if len(mailboxes) != 3 {
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

	messages, err := readAllMessages(c, mbox)
	if err != nil {
		return err
	}

	if len(messages) != 1 {
		return fmt.Errorf("expected one message, got %d", len(messages))
	}

	err = deleteAllMessages(c, mbox)
	if err != nil {
		return err
	}

	return nil
}

func TestAppend(args *Args) error {
	c, err := Connect(args)
	if err != nil {
		return err
	}

	err = c.Login(testUser1, args.Password)
	if err != nil {
		return err
	}
	defer c.Logout()

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		return err
	}

	email := mail.NewMessage(testUser1, testUser1, "note to self", "Don't forget to take over the world")

	err = deleteAllMessages(c, mbox)
	if err != nil {
		return err
	}

	messages, err := readAllMessages(c, mbox)
	if err != nil {
		return err
	}

	if len(messages) != 0 {
		return fmt.Errorf("expected no messages, got %d", len(messages))
	}

	// Append it to INBOX, with two flags
	flags := []string{imap.FlaggedFlag, "foobar"}
	if err := c.Append("INBOX", flags, time.Now(), strings.NewReader(email)); err != nil {
		return fmt.Errorf("failed to append message to inbox: %w", err)
	}

	messages, err = readAllMessages(c, mbox)
	if err != nil {
		return err
	}

	if len(messages) != 1 {
		return fmt.Errorf("expected one message, got %d", len(messages))
	}

	return deleteAllMessages(c, mbox)
}

func TestDeleteMessage(args *Args) error {
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
	// Send a bunch of test messages.
	for i := 0; i < 100; i++ {
		email := mail.NewMessage(testUser2, testUser1,
			fmt.Sprintf("integration testing message [%d]", i),
			fmt.Sprintf("This is test message %d", i))
		err = protoutil.SendMessage(ctx, key1, []byte(email), testUser1, testUser2, lookupClient)
		if err != nil {
			return err
		}
	}

	// Make sure all messages are received.
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

	defer func() {
		// Clean up, even if the test fails.
		_ = deleteAllMessages(c, mbox)
	}()

	messages, err := readAllMessages(c, mbox)
	if err != nil {
		return err
	}
	if len(messages) != 100 {
		return fmt.Errorf("expected 100 messages, got %d", len(messages))
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(uint32(77))
	err = c.Store(seqset, imap.FormatFlagsOp(imap.AddFlags, true), []interface{}{imap.DeletedFlag}, nil)
	if err != nil {
		return fmt.Errorf("failed to mark message as deleted, %w", err)
	}

	err = c.Expunge(nil)
	if err != nil {
		return fmt.Errorf("failed to expunge messages: %w", err)
	}

	messages, err = readAllMessages(c, mbox)
	if err != nil {
		return err
	}
	if len(messages) != 99 {
		return fmt.Errorf("expected 100 messages, got %d", len(messages))
	}

	for i := 0; i < 100; i++ {
		found := false
		for _, message := range messages {
			if strings.Contains(message.Envelope.Subject,
				fmt.Sprintf("message [%d]", i)) {
				found = true
				break
			}
		}
		if found {
			fmt.Printf("message %d found\n", i)
		} else {
			fmt.Printf("message %d NOT FOUND\n", i)
		}

	}

	// found := false
	// for _, message := range messages {
	// 	if strings.Contains(message.Envelope.Subject, "message 76") {
	// 		found = true
	// 		break
	// 	}
	// }
	// if found {
	// 	return fmt.Errorf("found message that was supposed to be deleted")
	// }

	return nil
}

func readAllMessages(c *client.Client, mbox *imap.MailboxStatus) ([]*imap.Message, error) {
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
		return nil, err
	}

	return messages, nil
}

func deleteAllMessages(c *client.Client, mbox *imap.MailboxStatus) error {
	messages, err := readAllMessages(c, mbox)
	if err != nil {
		return err
	}

	for i := range messages {
		seqset := new(imap.SeqSet)
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
