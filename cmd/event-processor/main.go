package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/golang/protobuf/proto"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/globals"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CmdArgs struct {
	DumpServiceURL   string
	LookupServiceURL string
	Key              string
	DBPassword       string
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05", NoColor: true})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	args := &CmdArgs{}
	flag.StringVar(&args.DumpServiceURL, "dump-service-url", globals.PublicDumpServiceURL, "dump service URL")
	flag.StringVar(&args.LookupServiceURL, "lookup-service-url", globals.PublicLookupServiceURL, "lookup service URL")
	flag.StringVar(&args.Key, "key", "", "key location")
	flag.StringVar(&args.DBPassword, "db-password", "", "db password")
	flag.Parse()

	var err error
	keyFile := args.Key
	if keyFile == "" {
		keyFile, err = util.GetDefaultKeyLocation()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get private key")
		}
	}

	encrypted, err := util.IsKeyEncrypted(keyFile)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot find key file")
	}

	passphrase := ""
	if encrypted {
		passphrase, err = util.ReadPassphase()
		if err != nil {
			log.Fatal().Err(err).Msg("cannot read passphrase")
		}
	}

	privateKey, err := easyecc.NewPrivateKeyFromFile(keyFile, passphrase)
	if err != nil {
		log.Fatal().Err(err).Str("location", keyFile).Msg("cannot load private key")
	}

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Second * 5),
	}

	dumpConn, err := grpc.Dial(args.DumpServiceURL, opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to the dump server")
	}
	defer dumpConn.Close()

	db, err := sql.Open("mysql", fmt.Sprintf("ubikom:%s@/ubikom", args.DBPassword))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open database connection")
	}
	// See "Important settings" section.
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	signed := protoutil.IdentityProof(privateKey, time.Now())

	ctx := context.Background()
	client := pb.NewDMSDumpServiceClient(dumpConn)
	lookupConn, err := grpc.Dial(args.LookupServiceURL, opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to the lookup server")
	}
	defer lookupConn.Close()
	lookupService := pb.NewLookupServiceClient(lookupConn)

	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		count := 0
		for {
			res, err := client.Receive(ctx, &pb.ReceiveRequest{IdentityProof: signed})
			if err != nil {
				st, ok := status.FromError(err)
				if !ok {
					log.Fatal().Err(err).Msg("error receiving messages")
				}
				if st.Code() == codes.NotFound {
					// This is expected - not new messages.
					break
				}
				log.Fatal().Err(err).Msg("error receiving messages")
			}
			msg := res.GetMessage()

			lookupRes, err := lookupService.LookupName(ctx, &pb.LookupNameRequest{Name: msg.GetSender()})
			if err != nil {
				log.Fatal().Err(err).Msg("failed to get receiver public key")
			}
			senderKey, err := easyecc.NewPublicFromSerializedCompressed(lookupRes.GetKey())
			if err != nil {
				log.Fatal().Err(err).Msg("invalid receiver public key")
			}

			if !protoutil.VerifySignature(msg.GetSignature(), lookupRes.GetKey(), msg.GetContent()) {
				log.Fatal().Msg("signature verification failed")
			}

			content, err := privateKey.Decrypt(msg.Content, senderKey)
			if err != nil {
				log.Fatal().Msg("failed to decode message")
			}

			event := &pb.Event{}
			err = proto.Unmarshal(content, event)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to unmarshal event")
			}

			ts := util.TimeFromMs(int64(event.Timestamp))
			tsStr := ts.Format("2006-01-02 15:04:05")
			stmt := "INSERT INTO events (id, timestamp, event_type, user1, user2, message, data1, component, flags) VALUES " +
				fmt.Sprintf("('%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', 0);", event.GetId(), tsStr, event.GetEventType().String(),
					strings.ToLower(event.GetUser1()), strings.ToLower(event.GetUser2()), event.GetMessage(), event.GetData1(), event.GetComponent())
			_, err = db.Exec(stmt)
			if err != nil {
				log.Fatal().Err(err).Msg("error executing insert statement")
			}
			count++
		}
		log.Info().Int("count", count).Msg("events imported")
	}
}
