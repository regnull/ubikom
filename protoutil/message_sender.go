package protoutil

import (
	"context"
	"fmt"

	"github.com/regnull/easyecc/v2"
	"github.com/regnull/ubikom/bc"
	"github.com/regnull/ubikom/pb"
	"github.com/rs/zerolog/log"
)

type MessageSender interface {
	Send(ctx context.Context, privateKey *easyecc.PrivateKey, body []byte,
		sender, receiver string) error
}

type messageSenderImpl struct {
	bchain                   bc.Blockchain
	dumpServiceClientFactory DumpServiceClientFactory
}

func NewMessageSender(dumpServiceClientFactory DumpServiceClientFactory, bchain bc.Blockchain) MessageSender {
	return &messageSenderImpl{
		bchain:                   bchain,
		dumpServiceClientFactory: dumpServiceClientFactory}
}

func (s *messageSenderImpl) Send(ctx context.Context, privateKey *easyecc.PrivateKey, body []byte,
	sender, receiver string) error {
	// Get receiver's public key.
	receiverKey, err := s.bchain.PublicKeyByCurve(ctx, receiver, privateKey.Curve())
	if err != nil {
		return fmt.Errorf("failed to get receiver public key: %w", err)
	}
	log.Debug().Msg("got receiver's public key")

	// Get receiver's address.
	endpoint, err := s.bchain.Endpoint(ctx, receiver)
	if err != nil {
		return fmt.Errorf("failed to get receiver's address: %w", err)
	}
	log.Debug().Str("address", endpoint).Msg("got receiver's address")

	msg, err := CreateMessage(privateKey, body, sender, receiver, receiverKey)
	if err != nil {
		return err
	}
	client, cleanup, err := s.dumpServiceClientFactory.CreateDumpServiceClient(ctx, endpoint, 0)
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}
	_, err = client.Send(ctx, &pb.SendRequest{Message: msg})
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	log.Debug().Msg("sent message successfully")
	return nil
}
