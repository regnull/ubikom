package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"
	"github.com/ethereum/go-ethereum/ethclient"
	_ "github.com/go-sql-driver/mysql"

	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/bc"
	"github.com/regnull/ubikom/globals"
	"github.com/regnull/ubikom/newscache"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	articleTTL       = 24 * time.Hour
	headlinesSubject = "Последние новости о войне"
	header           = `Новости Си-Эн-Эн

Каждая статья имеет номер. Пошлите сообщение с этим номером в теме чтобы получить статью полностью. 

Если вы пользуетесь зашифрованной почтой Ubikom, то ваше взаимодействие с war-info@ubikom.cc не регистрируется и
не отслеживается. Метаинформация о ваших сообщениях всегда зашифрована. Обслуживающие серверы находятся
за пределами РФ. Регестрируйтесь здесь: https://ubikom.cc/ru/index.html.

`
	footer = `
`
)

type CmdArgs struct {
	DumpServiceURL         string
	LookupServiceURL       string
	Key                    string
	DBPassword             string
	BlockchainNodeURL      string
	UseLegacyLookupService bool
	UbikomName             string
}

type CacheEntry struct {
	Url      string
	Added    time.Time
	Headline string
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05", NoColor: true})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	args := &CmdArgs{}
	flag.StringVar(&args.DumpServiceURL, "dump-service-url", globals.PublicDumpServiceURL, "dump service URL")
	flag.StringVar(&args.LookupServiceURL, "lookup-service-url", globals.PublicLookupServiceURL, "lookup service URL")
	flag.StringVar(&args.Key, "key", "/Users/regnull/.ubikom/war-info.key", "key location")
	flag.StringVar(&args.BlockchainNodeURL, "blockchain-node-url", globals.BlockchainNodeURL, "blockchain node url")
	flag.BoolVar(&args.UseLegacyLookupService, "use-legacy-lookup-service", false, "use legacy lookup service")
	flag.StringVar(&args.UbikomName, "ubikom-name", "war-info", "ubikom name")
	flag.Parse()

	//os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/Users/regnull/gcloud/clear-talent-299521-9a3e9ed59bf1.json")

	var err error
	keyFile := args.Key
	if keyFile == "" {
		keyFile, err = util.GetDefaultKeyLocation()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get private key")
		}
	}

	if args.UbikomName == "" {
		log.Fatal().Msg("ubikom-name must be specified")
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

	signed, err := protoutil.IdentityProof(privateKey, time.Now())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to generate identity proof")
	}

	ctx := context.Background()
	client := pb.NewDMSDumpServiceClient(dumpConn)
	lookupConn, err := grpc.Dial(args.LookupServiceURL, opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to the lookup server")
	}
	defer lookupConn.Close()
	lookupService := pb.NewLookupServiceClient(lookupConn)

	log.Info().Str("url", args.BlockchainNodeURL).Msg("connecting to blockchain node")
	blockchainClient, err := ethclient.Dial(args.BlockchainNodeURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to blockchain node")
	}
	blockchain := bc.NewBlockchain(blockchainClient, globals.KeyRegistryContractAddress,
		globals.NameRegistryContractAddress, globals.ConnectorRegistryContractAddress, nil)

	var combinedLookupClient pb.LookupServiceClient
	if args.UseLegacyLookupService {
		log.Info().Msg("using legacy lookup service")
		combinedLookupClient = lookupService
	} else {
		combinedLookupClient = bc.NewLookupServiceClient(blockchain, lookupService, false)
	}

	cache := newscache.New()
	err = cache.Refresh()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get headlines")
	}

	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		for range ticker.C {
			err := cache.Refresh()
			if err != nil {
				log.Error().Err(err).Msg("error refreshing headlines")
			}
		}
	}()

	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
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

			lookupRes, err := combinedLookupClient.LookupName(ctx, &pb.LookupNameRequest{Name: msg.GetSender()})
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

			r := bytes.NewReader(filterMalformedHeaders(content))
			e, err := message.Read(r)
			if err != nil {
				log.Error().Err(err).Msg("failed to read message")
				continue
			}
			h := mail.HeaderFromMap(e.Header.Map())
			al, err := h.AddressList("From")
			if err != nil {
				log.Error().Err(err).Msg("failed to get from address")
				continue
			}
			if len(al) != 1 {
				log.Error().Err(err).Msg("more than one address in from address list")
				continue
			}
			to := al[0].Address

			subj, err := h.Subject()
			if err != nil {
				log.Error().Err(err).Msg("failed to get subject")
				continue
			}

			var articleId int64
			if subj != "" {
				articleId, _ = strconv.ParseInt(subj, 10, 32)
			}

			log.Debug().Str("to", to).Msg("got address")
			if articleId != 0 {
				log.Debug().Int("id", int(articleId)).Msg("getting article")
				headline, text, err := cache.GetArticle(int(articleId))
				if err != nil {
					log.Error().Int("id", int(articleId)).Err(err).Msg("error retrieving article")
				} else {
					err = sendArticle(ctx, text, headline, to, args.UbikomName, msg.Sender, privateKey, lookupService)
					if err != nil {
						log.Error().Int("id", int(articleId)).Err(err).Msg("error sending the article")
					}
				}
				continue
			}
			headlines := cache.GetHeadlines()

			buf := new(bytes.Buffer)
			buf.WriteString(header)
			for _, h := range headlines {
				buf.WriteString(fmt.Sprintf("[%d] %s\n\n", h.ID, h.Title))
			}
			buf.WriteString(footer)

			resp, err := CreateTextEmail(&Email{
				From: &mail.Address{
					Name:    "Ubikom War Info",
					Address: "war-info@ubikom.cc",
				},
				To: []*mail.Address{
					{
						Address: to,
					},
				},
				Subject: headlinesSubject,
				Date:    time.Now(),
				Body:    buf.String(),
			})
			if err != nil {
				log.Error().Err(err).Msg("failed to create email")
			}

			respReceiver := "gateway"
			if msg.Sender != "gateway" {
				respReceiver = msg.Sender
			}

			err = protoutil.SendEmail(ctx, privateKey, resp, args.UbikomName, respReceiver, lookupService)
			if err != nil {
				log.Error().Err(err).Msg("failed to send response")
			} else {
				log.Info().Str("to", respReceiver).Msg("message sent")
			}
		}
	}
}

func sendArticle(ctx context.Context, text string, headline string, to string, ubikomName string, sender string,
	privateKey *easyecc.PrivateKey, lookupService pb.LookupServiceClient) error {
	respMsg, err := CreateTextEmail(&Email{
		From: &mail.Address{
			Name:    "Ubikom War Info",
			Address: "war-info@ubikom.cc",
		},
		To: []*mail.Address{
			{
				Address: to,
			},
		},
		Subject: headline,
		Date:    time.Now(),
		Body:    text,
	})
	if err != nil {
		return err
	}

	respReceiver := "gateway"
	if sender != "gateway" {
		respReceiver = sender
	}

	err = protoutil.SendEmail(ctx, privateKey, []byte(respMsg), ubikomName, respReceiver, lookupService)
	if err != nil {
		log.Error().Err(err).Msg("failed to send response")
	} else {
		log.Info().Str("to", respReceiver).Msg("message sent")
	}
	return nil
}

func filterMalformedHeaders(body []byte) []byte {
	bodyStr := string(body)
	lines := strings.Split(bodyStr, "\n")
	var newLines []string
	headers := true
	for _, line := range lines {
		if headers && (line == "" || line == "\r") {
			// Done with headers.
			headers = false
		}
		if headers &&
			(strings.HasPrefix(line, ">From") || strings.HasPrefix(line, "From") &&
				!strings.HasPrefix(line, "From:")) {
			continue
		}
		newLines = append(newLines, line)
	}
	newBody := strings.Join(newLines, "\n")
	return []byte(newBody)
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
	h.Set("Content-Language", "ru")
	h.Set("Content-Type", "text/plain; charset=utf-8; format=flowed")
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
