package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"

	"github.com/dgraph-io/badger/v3"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/db"
	"github.com/regnull/ubikom/protoio"
	"github.com/regnull/ubikom/protoutil"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

type CmdArgs struct {
	KeyLocation string
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	var args CmdArgs
	flag.StringVar(&args.KeyLocation, "key", "", "key location")
	flag.Parse()

	keyFile := args.KeyLocation
	var err error
	if keyFile == "" {
		keyFile, err = util.GetDefaultKeyLocation()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get private key")
		}
	}
	key, err := easyecc.NewPrivateKeyFromFile(keyFile, "")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load private key")
	}
	log.Info().Str("file", keyFile).Msg("private key loaded")

	badger, err := badger.Open(badger.DefaultOptions("db"))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize the database")
	}
	db := db.NewBadgerDB(badger)
	for i := 0; i < 100; i++ {
		k, _ := easyecc.NewRandomPrivateKey()
		db.RegisterKey(k.PublicKey())
	}
	f, err := os.Create("keys")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create file")
	}
	w := bufio.NewWriter(f)
	hashWriter := protoio.NewSha256Writer(w)
	protoWriter := protoio.NewWriter(hashWriter)
	err = db.WriteKeys(protoWriter, math.MaxUint64)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to write keys")
	}
	w.Flush()
	f.Close()

	// The header will include hashes of all files, one line per file, in "name hash\n" format.
	header := fmt.Sprintf("keys %X\n", hashWriter.Hash())

	signed, err := protoutil.CreateSigned(key, []byte(header))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create signature")
	}
	b, err := proto.Marshal(signed)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to marshal signature")
	}
	err = ioutil.WriteFile("snapshot_sig", b, 0440)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to write snapshot signature")
	}
}
