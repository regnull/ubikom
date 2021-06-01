package main

import (
	"flag"
	"fmt"
	"html"
	"net/http"
	"os"
	"time"

	"github.com/regnull/ubikom/globals"
	"github.com/regnull/ubikom/pb"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type CmdArgs struct {
	Port             int
	LookupServiceURL string
	Timeout          time.Duration
}

type Server struct {
	lookupClient pb.LookupServiceClient
}

func (s *Server) HandleNameLookup(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	res, err := s.lookupClient.LookupName(r.Context(), &pb.LookupNameRequest{
		Name: name,
	})
	w.Header().Add("Content-Type", "application/json")
	if err != nil {
		log.Error().Err(err).Msg("name lookup request failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if res.GetResult().GetResult() == pb.ResultCode_RC_RECORD_NOT_FOUND {
		fmt.Fprintf(w, `{
	"name": "%s", 
	"available": true
}`, name)
		return
	}
	if res.GetResult().GetResult() == pb.ResultCode_RC_INVALID_REQUEST {
		log.Warn().Str("name", name).Msg("invalid request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if res.GetResult().GetResult() != pb.ResultCode_RC_OK {
		log.Error().Str("result", res.GetResult().GetResult().String()).Msg("server returned error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// If we got here, the name record was found.
	fmt.Fprintf(w, `{
	"name": "%s", 
	"available": false
}`, name)
}

func (s *Server) HandleEasySetup(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

func (s *Server) HandleGetKey(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	var args CmdArgs
	flag.IntVar(&args.Port, "port", 8088, "HTTP port")
	flag.StringVar(&args.LookupServiceURL, "lookup-service-url", globals.PublicLookupServiceURL, "lookup service url")
	flag.DurationVar(&args.Timeout, "timeout", 5*time.Second, "timeout when connecting to the lookup service")
	flag.Parse()

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(args.Timeout),
	}
	conn, err := grpc.Dial(args.LookupServiceURL, opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to the server")
	}
	defer conn.Close()

	lookupClient := pb.NewLookupServiceClient(conn)

	server := &Server{lookupClient: lookupClient}

	http.HandleFunc("/lookupName", server.HandleNameLookup)
	http.HandleFunc("/easySetup", server.HandleEasySetup)
	http.HandleFunc("/getKey", server.HandleGetKey)
	log.Info().Int("port", args.Port).Msg("listening...")
	log.Fatal().Err(http.ListenAndServe(fmt.Sprintf(":%d", args.Port), nil))
}
