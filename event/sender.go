package event

import (
	"context"

	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/regnull/ubikom/util"
)

type Sender struct {
	target       string
	lookupClient *pb.LookupServiceClient
}

func NewSender(target string, lookupClient *pb.LookupServiceClient) *Sender {
	return &Sender{target: target, lookupClient: lookupClient}
}

func (s *Sender) KeyRegistered(ctx context.Context, privateKey *easyecc.PrivateKey, sender string, address string) error {
	event := &pb.Event{
		Id:        uuid.New().String(),
		Timestamp: uint64(util.NowMs()),
		EventType: pb.EventType_ET_KEY_REGISTRATION,
		Data1:     address,
		Message:   "New key was registered",
	}
	b, err := proto.Marshal(event)
	if err != nil {
		return err
	}
	err = protoutil.SendMessage(ctx, privateKey, b, sender, s.target, *s.lookupClient)
	if err != nil {
		return err
	}
	return err
}
