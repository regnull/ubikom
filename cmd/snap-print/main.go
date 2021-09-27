package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoio"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type CmdArgs struct {
	KeysFile string
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	var args CmdArgs
	flag.StringVar(&args.KeysFile, "keys-file", "", "keys file location")
	flag.Parse()

	if args.KeysFile != "" {
		f, err := os.OpenFile(args.KeysFile, os.O_RDONLY, 0)
		if err != nil {
			log.Error().Err(err).Msg("failed to open file")
		}
		reader := protoio.NewReader(f)
		for {
			msg, err := reader.Read(func(b []byte) (proto.Message, error) {
				var key pb.ExportKeyRecord
				err := proto.Unmarshal(b, &key)
				if err != nil {
					return nil, err
				}
				return &key, nil
			})
			if err != nil {
				break
			}
			opts := protojson.MarshalOptions{
				Multiline: true,
				Indent:    "  ",
			}
			json, err := opts.Marshal(msg)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to marshal to JSON")
			}
			fmt.Printf("%s\n", json)
		}
		f.Close()
	}
}
