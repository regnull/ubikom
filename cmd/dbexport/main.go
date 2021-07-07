package main

import (
	"bufio"
	"math"
	"os"

	"github.com/dgraph-io/badger/v3"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/db"
	"github.com/regnull/ubikom/protoio"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
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
	protoWriter := protoio.NewWriter(w)
	err = db.WriteKeys(protoWriter, math.MaxUint64)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to write keys")
	}
	w.Flush()
	f.Close()
}
