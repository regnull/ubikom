package gateway

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/mail"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
	"google.golang.org/grpc/codes"
)

// SendOptions is used to pass options when creating a Sender.
type SenderOptions struct {
	PrivateKey             *easyecc.PrivateKey
	LookupClient           pb.LookupServiceClient
	DumpClient             pb.DMSDumpServiceClient
	GlobalRateLimitPerHour int
	PollInterval           time.Duration
}

// Sender sends emails to the outside world.
type Sender struct {
	privateKey             *easyecc.PrivateKey
	lookupClient           pb.LookupServiceClient
	dumpClient             pb.DMSDumpServiceClient
	globalRateLimitPerHour int
	pollInterval           time.Duration
}

// NewSender creates and returns a new sender.
func NewSender(opts *SenderOptions) *Sender {
	return &Sender{
		privateKey:             opts.PrivateKey,
		lookupClient:           opts.LookupClient,
		dumpClient:             opts.DumpClient,
		globalRateLimitPerHour: opts.GlobalRateLimitPerHour,
		pollInterval:           opts.PollInterval}
}

// Run blocks while running receive loop and returns when the context expires, or
// when an unrecoverable error happens.
func (s *Sender) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.pollInterval)
	globalRateLimiter := rate.NewLimiter(rate.Every(time.Hour), s.globalRateLimitPerHour)
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context done")
		case <-ticker.C:
			// Poll for messages, ignore errors.
			// TODO: Maybe exit if a serious error occurs.
			_ = s.poll(ctx, globalRateLimiter)
		}
	}
}

// Poll the dump server, send messages to the outside world.
func (s *Sender) poll(ctx context.Context, rateLimiter *rate.Limiter) error {
	// Get all the messages ready to send.
	var messages []string
	for {
		res, err := s.dumpClient.Receive(ctx, &pb.ReceiveRequest{IdentityProof: protoutil.IdentityProof(s.privateKey)})
		if err != nil && util.StatusCodeFromError(err) == codes.NotFound {
			log.Info().Msg("no new messages")
			break
		}
		if err != nil {
			log.Error().Err(err).Msg("failed to receive message")
			return fmt.Errorf("failed to receive message")
		}
		msg := res.GetMessage()
		content, err := protoutil.DecryptMessage(ctx, s.lookupClient, s.privateKey, msg)
		if err != nil {
			log.Error().Err(err).Msg("failed to decrypt message")
			continue
		}
		messages = append(messages, content)
	}

	// Now that we have all the messages, send them out one by one.

	for _, content := range messages {
		// Check rate limit.

		if !rateLimiter.Allow() {
			log.Warn().Msg("external send is blocked by the global rate limiter")
			continue
		}

		// Parse email.

		rewritten, from, to, err := mail.RewriteFromHeader(content)
		if err != nil {
			log.Error().Err(err).Msg("error re-writting message")
			continue
		}

		// Pipe to sendmail.

		cmd := exec.Command("sendmail", "-t", "-f", from)
		cmd.Stdin = bytes.NewReader([]byte(rewritten))
		err = cmd.Run()
		if err != nil {
			log.Error().Err(err).Msg("error running sendmail")
			continue
		}

		log.Debug().Strs("to", to).Msg("external mail sent")
	}
	return nil
}
