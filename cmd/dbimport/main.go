package main

import (
	"bufio"
	"flag"
	"fmt"
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

	// Import keys.
	keysPath := path.Join(args.SnapDir, "keys")
	keysCount, err := importItems(keysPath, func(b []byte) (proto.Message, error) {
		var key pb.ExportKeyRecord
		err := proto.Unmarshal(b, &key)
		if err != nil {
			log.Error().Err(err).Msg("unmarshal error")
			return nil, err
		}
		return &key, nil
	}, func(p proto.Message) error {
		return db.ImportKey(p.(*pb.ExportKeyRecord))
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to import keys")
	}
	log.Info().Int("count", keysCount).Msg("keys imported")

	// Import names.
	namesPath := path.Join(args.SnapDir, "names")
	namesCount, err := importItems(namesPath, func(b []byte) (proto.Message, error) {
		var rec pb.ExportNameRecord
		err := proto.Unmarshal(b, &rec)
		if err != nil {
			log.Error().Err(err).Msg("unmarshal error")
			return nil, err
		}
		return &rec, nil
	}, func(p proto.Message) error {
		return db.ImportName(p.(*pb.ExportNameRecord))
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to import names")
	}
	log.Info().Int("count", namesCount).Msg("names imported")

	// Import addresses.
	addressesPath := path.Join(args.SnapDir, "addresses")
	addressesCount, err := importItems(addressesPath, func(b []byte) (proto.Message, error) {
		var rec pb.ExportAddressRecord
		err := proto.Unmarshal(b, &rec)
		if err != nil {
			log.Error().Err(err).Msg("unmarshal error")
			return nil, err
		}
		return &rec, nil
	}, func(p proto.Message) error {
		return db.ImportAddress(p.(*pb.ExportAddressRecord))
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to import names")
	}
	log.Info().Int("count", addressesCount).Msg("addresses imported")
}

func importItems(filePath string, parseFunc func([]byte) (proto.Message, error), applyFunc func(proto.Message) error) (int, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %w", err)
	}
	reader := bufio.NewReader(f)
	count := 0
	protoReader := protoio.NewReader(reader)
	for {
		p, err := protoReader.Read(parseFunc)
		if err != nil {
			if err == io.EOF {
				break
			}
			return count, fmt.Errorf("failed to read proto: %w", err)
		}
		err = applyFunc(p)
		if err != nil {
			return count, fmt.Errorf("failed to process proto: %w", err)
		}
		count++
	}
	return count, nil
}
