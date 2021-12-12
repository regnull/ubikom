package gateway

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/event"
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
	PrivateKey       *easyecc.PrivateKey
	LookupClient     pb.LookupServiceClient
	DumpClient       pb.DMSDumpServiceClient
	RateLimitPerHour int
	PollInterval     time.Duration
	ExternalSender   ExternalSender
	EventTarget      string
}

// Sender sends emails to the outside world.
type Sender struct {
	privateKey       *easyecc.PrivateKey
	lookupClient     pb.LookupServiceClient
	dumpClient       pb.DMSDumpServiceClient
	rateLimitPerHour int
	pollInterval     time.Duration
	externalSender   ExternalSender
	rateLimiters     map[string]*rate.Limiter
	eventSender      *event.Sender
}

// NewSender creates and returns a new sender.
func NewSender(opts *SenderOptions) *Sender {
	var eventSender *event.Sender
	if opts.EventTarget != "" {
		eventSender = event.NewSender(opts.EventTarget, "gateway", "gateway", opts.PrivateKey, opts.LookupClient)
	}
	return &Sender{
		privateKey:       opts.PrivateKey,
		lookupClient:     opts.LookupClient,
		dumpClient:       opts.DumpClient,
		rateLimitPerHour: opts.RateLimitPerHour,
		pollInterval:     opts.PollInterval,
		externalSender:   opts.ExternalSender,
		rateLimiters:     make(map[string]*rate.Limiter),
		eventSender:      eventSender,
	}
}

// Run blocks while running receive loop and returns when the context expires, or
// when an unrecoverable error happens.
func (s *Sender) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.pollInterval)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return fmt.Errorf("context done")
		case <-ticker.C:
			// Poll for messages, ignore errors.
			// TODO: Maybe exit if a serious error occurs.
			_ = s.poll(ctx)
		}
	}
}

// Poll the dump server, send messages to the outside world.
func (s *Sender) poll(ctx context.Context) error {
	// Get all the messages ready to send.
	var outgoingMessages []string
	var messages []*pb.DMSMessage
	for {
		res, err := s.dumpClient.Receive(ctx, &pb.ReceiveRequest{IdentityProof: protoutil.IdentityProof(s.privateKey)})
		if err != nil && util.StatusCodeFromError(err) == codes.NotFound {
			if len(messages) != 0 {
				log.Info().Int("count", len(messages)).Msg("messages received")
			}
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
		messages = append(messages, msg)
		outgoingMessages = append(outgoingMessages, content)
	}

	// Now that we have all the messages, send them out one by one.

	for i, content := range outgoingMessages {
		// Parse email.
		rewritten, from, to, err := mail.RewriteFromHeader(content)
		if err != nil {
			log.Error().Err(err).Msg("error re-writting message")
			continue
		}

		msg := messages[i]
		blocked := false
		// Check rate limit.
		for range to {
			// Check rate limit.
			if s.rateLimiters[msg.Sender] == nil {
				s.rateLimiters[msg.Sender] = rate.NewLimiter(rate.Every(time.Hour), s.rateLimitPerHour)
			}
			if !s.rateLimiters[msg.Sender].Allow() {
				blocked = true
			}
		}

		if blocked {
			subject, _ := mail.ExtractSubject(content)
			err = NotifyMessageBlocked(ctx, s.privateKey, s.lookupClient, msg.Sender, strings.Join(to, ", "), subject)
			if err != nil {
				log.Error().Err(err).Msg("failed to send rate limit exceeded notification message")
			}
			log.Warn().Str("user", msg.Sender).Msg("outgoing message blocked due to the rate limit")
			continue
		}

		if s.eventSender != nil {
			for _, target := range to {
				err = s.eventSender.ExternalEmailSend(ctx, from, target)
				if err != nil {
					log.Error().Err(err).Msg("failed to send event")
				}
			}
		}

		// Pipe to sendmail.
		err = s.externalSender.Send(from, rewritten)
		if err != nil {
			log.Error().Err(err).Msg("error sending mail")
			continue
		}
		// TODO: Retry and notify the sender of any errors.

		log.Debug().Strs("to", to).Msg("external mail sent")
	}
	return nil
}
