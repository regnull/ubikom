package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/emersion/go-message/mail"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/globals"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"google.golang.org/grpc"
)

const body = `Hello, Ubikom users.

First, we apologize for getting in touch with you without your permission. We are going to add our twitter handle to the registration page, so that people may opt into receiving messages from the Ubikom team. It is a learning process for us in these very early days of the platform. 

We want to ask you a favor and fill out a short anonymous survey about Ubikom. It will help all of us to improve the product and offer the benefits of the free encrypted email platform to many users. 

This is the link to the survey: https://docs.google.com/forms/d/1saorAZCxaRuUMpyDKPkH4AJtWtvYWCEcGBvmVSk6mTA/edit?usp=sharing

You can also subscribe to our twitter feed: https://twitter.com/UbikomProject, join our subreddit at https://www.reddit.com/r/ubikom/, or check new developments on our blog page (https://www.ubikom.cc/blogs.html). For example, in the last two weeks we migrated the avatar registry to a custom blockchain and enabled webmail. 

We are planning to transition to a new blockchain in the future and distribute tokens to the early adopters. Stay connected!

You can always reach us with questions, suggestions, and proposals to collaborate at lgx@ubikom.cc, sasha@ubikom.cc. 
`

type CmdArgs struct {
	KeyFile         string
	From            string
	LookupServerURL string
	UsersFile       string
}

func main() {
	var args CmdArgs
	flag.StringVar(&args.KeyFile, "key", "", "key file")
	flag.StringVar(&args.From, "from", "", "from")
	flag.StringVar(&args.LookupServerURL, "lookup-server-url", globals.PublicLookupServiceURL, "URL of the lookup server")
	flag.StringVar(&args.UsersFile, "users", "", "users file")
	flag.Parse()

	if args.KeyFile == "" {
		log.Fatal("--key must be specified")
	}

	if args.From == "" {
		log.Fatal("--from must be specified")
	}

	if args.UsersFile == "" {
		log.Fatal("--users file must be specified")
	}

	privateKey, err := easyecc.NewPrivateKeyFromFile(args.KeyFile, "")
	if err != nil {
		log.Fatal(err)
	}

	lookupService, conn, err := connectToLookupService(args.LookupServerURL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	usersBytes, err := ioutil.ReadFile(args.UsersFile)
	if err != nil {
		log.Fatal(err)
	}

	usersLines := strings.Split(string(usersBytes), "\n")
	for _, user := range usersLines {
		if len(user) == 0 {
			continue
		}
		//fmt.Printf("%s\n", user)
	}

	//	recs := []string{"sasha", "lgx"}
	recs := usersLines

	for _, r := range recs {
		if len(r) == 0 {
			continue
		}
		email, err := CreateTextEmail(&Email{
			From: &mail.Address{
				Address: fmt.Sprintf("%s@ubikom.cc", args.From),
			},
			To: []*mail.Address{
				{
					Address: fmt.Sprintf("%s@ubikom.cc", r),
				},
			},
			Subject: "Please help us to improve Ubikom",
			Date:    time.Now(),
			Body:    body,
		})
		if err != nil {
			log.Fatal(err)
		}
		err = protoutil.SendEmail(context.Background(), privateKey, email, args.From, r, lookupService)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("sent email to %s", r)
	}
}

type Email struct {
	From    *mail.Address
	To      []*mail.Address
	Cc      []*mail.Address
	Subject string
	Date    time.Time
	Body    string
}

func CreateTextEmail(email *Email) ([]byte, error) {
	var h mail.Header
	h.SetDate(email.Date)
	h.SetAddressList("From", []*mail.Address{email.From})
	h.SetAddressList("To", email.To)
	if email.Cc != nil {
		h.SetAddressList("Cc", email.Cc)
	}
	h.SetSubject(email.Subject)

	var b bytes.Buffer

	w, err := mail.CreateSingleInlineWriter(&b, h)
	if err != nil {
		return nil, err
	}
	_, err = w.Write([]byte(email.Body))
	if err != nil {
		return nil, err
	}
	w.Close()
	return b.Bytes(), nil
}

func connectToLookupService(url string) (pb.LookupServiceClient, *grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Second * 5),
	}

	conn, err := grpc.Dial(url, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to lookup service: %w", err)
	}

	return pb.NewLookupServiceClient(conn), conn, nil
}
