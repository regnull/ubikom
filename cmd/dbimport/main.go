package main

import (
	"bufio"
	"flag"
	"io"
	"os"
	"path"

	"github.com/dgraph-io/badger/v3"
	"github.com/regnull/ubikom/db"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoio"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

type CmdArgs struct {
	DbDir   string
	SnapDir string
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	var args CmdArgs
	flag.StringVar(&args.DbDir, "db", "", "database directory")
	flag.StringVar(&args.SnapDir, "snap-dir", "", "snapshot directory")
	flag.Parse()

	if args.DbDir == "" {
		log.Fatal().Msg("database directory must be specified")
	}

	if args.SnapDir == "" {
		log.Fatal().Msg("snapshot directory must be specified")
	}

	badger, err := badger.Open(badger.DefaultOptions(args.DbDir))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize the database")
	}

	db := db.NewBadgerDB(badger)

	keysPath := path.Join(args.SnapDir, "keys")
	f, err := os.Open(keysPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open keys file")
	}
	reader := bufio.NewReader(f)
	protoReader := protoio.NewReader(reader)
	keysCount := 0
	for {
		key, err := protoReader.Read(func(b []byte) (proto.Message, error) {
			var key pb.ExportKeyRecord
			err := proto.Unmarshal(b, &key)
			if err != nil {
				log.Error().Err(err).Msg("unmarshal error")
				return nil, err
			}
			return &key, nil
		})
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal().Err(err).Msg("failed to read proto")
		}
		err = db.ImportKey(key.(*pb.ExportKeyRecord))
		if err != nil {
			log.Fatal().Err(err).Msg("failed to import key record")
		}
		keysCount++
	}
	log.Info().Int("count", keysCount).Msg("keys imported")
}
