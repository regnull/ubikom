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
	sender       string
	component    string
	lookupClient pb.LookupServiceClient
}

func NewSender(target string, sender string, component string, lookupClient pb.LookupServiceClient) *Sender {
	return &Sender{target: target, sender: sender, component: component, lookupClient: lookupClient}
}

func (s *Sender) KeyRegistered(ctx context.Context, privateKey *easyecc.PrivateKey, keyAddress string) error {
	event := &pb.Event{
		Id:        uuid.New().String(),
		Timestamp: uint64(util.NowMs()),
		EventType: pb.EventType_ET_KEY_REGISTRATION,
		Data1:     keyAddress,
		Message:   "New key was registered",
		Component: s.component}
	b, err := proto.Marshal(event)
	if err != nil {
		return err
	}
	err = protoutil.SendMessage(ctx, privateKey, b, s.sender, s.target, s.lookupClient)
	if err != nil {
		return err
	}
	return err
}

func (s *Sender) NameRegistered(ctx context.Context, privateKey *easyecc.PrivateKey, keyAddress string, name string) error {
	event := &pb.Event{
		Id:        uuid.New().String(),
		Timestamp: uint64(util.NowMs()),
		EventType: pb.EventType_ET_NAME_REGISTRATION,
		User1:     name,
		Data1:     keyAddress,
		Message:   "New name was registered",
		Component: s.component}
	b, err := proto.Marshal(event)
	if err != nil {
		return err
	}
	err = protoutil.SendMessage(ctx, privateKey, b, s.sender, s.target, s.lookupClient)
	if err != nil {
		return err
	}
	return err
}

func (s *Sender) AddressRegistered(ctx context.Context, privateKey *easyecc.PrivateKey, address string, name string) error {
	event := &pb.Event{
		Id:        uuid.New().String(),
		Timestamp: uint64(util.NowMs()),
		EventType: pb.EventType_ET_ADDRESS_REGISTRATION,
		User1:     name,
		Data1:     address,
		Message:   "New address was registered",
		Component: s.component}
	b, err := proto.Marshal(event)
	if err != nil {
		return err
	}
	err = protoutil.SendMessage(ctx, privateKey, b, s.sender, s.target, s.lookupClient)
	if err != nil {
		return err
	}
	return err
}
