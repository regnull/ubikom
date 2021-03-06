package event

import (
	"context"
	"fmt"

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
	privateKey   *easyecc.PrivateKey
	lookupClient pb.LookupServiceClient
}

func NewSender(target string, sender string, component string,
	privateKey *easyecc.PrivateKey, lookupClient pb.LookupServiceClient) *Sender {
	return &Sender{
		target:       target,
		sender:       sender,
		component:    component,
		privateKey:   privateKey,
		lookupClient: lookupClient}
}

func (s *Sender) KeyRegistered(ctx context.Context, keyAddress string) error {
	event := &pb.Event{
		Id:        uuid.New().String(),
		Timestamp: uint64(util.NowMs()),
		EventType: pb.EventType_ET_KEY_REGISTRATION,
		User1:     keyAddress,
		Message:   "New key was registered",
		Component: s.component}
	return marshalAndSend(ctx, s.privateKey, s.sender, s.target, s.lookupClient, event)
}

func (s *Sender) NameRegistered(ctx context.Context, keyAddress string, name string) error {
	event := &pb.Event{
		Id:        uuid.New().String(),
		Timestamp: uint64(util.NowMs()),
		EventType: pb.EventType_ET_NAME_REGISTRATION,
		User1:     keyAddress,
		Data1:     name,
		Message:   "New name was registered",
		Component: s.component}
	return marshalAndSend(ctx, s.privateKey, s.sender, s.target, s.lookupClient, event)
}

func (s *Sender) AddressRegistered(ctx context.Context, address string, name string) error {
	event := &pb.Event{
		Id:        uuid.New().String(),
		Timestamp: uint64(util.NowMs()),
		EventType: pb.EventType_ET_ADDRESS_REGISTRATION,
		User1:     name,
		Data1:     address,
		Message:   "New address was registered",
		Component: s.component}
	return marshalAndSend(ctx, s.privateKey, s.sender, s.target, s.lookupClient, event)
}

func (s *Sender) POPLogin(ctx context.Context, name string) error {
	event := &pb.Event{
		Id:        uuid.New().String(),
		Timestamp: uint64(util.NowMs()),
		EventType: pb.EventType_ET_PROXY_POP_LOGIN,
		User1:     name,
		Message:   "User logged in via POP",
		Component: s.component}
	return marshalAndSend(ctx, s.privateKey, s.sender, s.target, s.lookupClient, event)
}

func (s *Sender) IMAPLogin(ctx context.Context, name string) error {
	event := &pb.Event{
		Id:        uuid.New().String(),
		Timestamp: uint64(util.NowMs()),
		EventType: pb.EventType_ET_PROXY_IMAP_LOGIN,
		User1:     name,
		Message:   "User logged in via IMAP",
		Component: s.component}
	return marshalAndSend(ctx, s.privateKey, s.sender, s.target, s.lookupClient, event)
}

func (s *Sender) SMTPLogin(ctx context.Context, name string) error {
	event := &pb.Event{
		Id:        uuid.New().String(),
		Timestamp: uint64(util.NowMs()),
		EventType: pb.EventType_ET_PROXY_SMTP_LOGIN,
		User1:     name,
		Message:   "User logged in via SMTP",
		Component: s.component}
	return marshalAndSend(ctx, s.privateKey, s.sender, s.target, s.lookupClient, event)
}

func (s *Sender) SMTPSend(ctx context.Context, name string, toInternal, toExternal string) error {
	event := &pb.Event{
		Id:        uuid.New().String(),
		Timestamp: uint64(util.NowMs()),
		EventType: pb.EventType_ET_PROXY_SMTP_MESSAGE_SENT,
		User1:     name,
		Message:   "User sent an SMTP message",
		Component: s.component,
		User2:     toInternal,
		Data1:     toExternal}
	return marshalAndSend(ctx, s.privateKey, s.sender, s.target, s.lookupClient, event)
}

func (s *Sender) WebMailLogin(ctx context.Context, name string) error {
	event := &pb.Event{
		Id:        uuid.New().String(),
		Timestamp: uint64(util.NowMs()),
		EventType: pb.EventType_ET_PROXY_WEBMAIL_LOGIN,
		User1:     name,
		Message:   "User logged in via web mail",
		Component: s.component}
	return marshalAndSend(ctx, s.privateKey, s.sender, s.target, s.lookupClient, event)
}

func (s *Sender) WebMailSend(ctx context.Context, name string, toInternal, toExternal string) error {
	event := &pb.Event{
		Id:        uuid.New().String(),
		Timestamp: uint64(util.NowMs()),
		EventType: pb.EventType_ET_PROXY_WEBMAIL_MESSAGE_SENT,
		User1:     name,
		Message:   "User sent a message via web mail",
		Component: s.component,
		User2:     toInternal,
		Data1:     toExternal}
	return marshalAndSend(ctx, s.privateKey, s.sender, s.target, s.lookupClient, event)
}

func (s *Sender) ExternalEmailSend(ctx context.Context, name string, toExternal string) error {
	event := &pb.Event{
		Id:        uuid.New().String(),
		Timestamp: uint64(util.NowMs()),
		EventType: pb.EventType_ET_GATEWAY_EMAIL_MESSAGE_SENT,
		User1:     name,
		Message:   "User sent an external email",
		Component: s.component,
		User2:     toExternal}
	return marshalAndSend(ctx, s.privateKey, s.sender, s.target, s.lookupClient, event)
}

func (s *Sender) GatewayUbikomMessageSend(ctx context.Context, from, to string) error {
	event := &pb.Event{
		Id:        uuid.New().String(),
		Timestamp: uint64(util.NowMs()),
		EventType: pb.EventType_ET_GATEWAY_UBIKOM_MESSAGE_SENT,
		User1:     from,
		Message:   "Gateway sent message to internal user",
		Component: s.component,
		User2:     to}
	return marshalAndSend(ctx, s.privateKey, s.sender, s.target, s.lookupClient, event)
}

func (s *Sender) WebPageServed(ctx context.Context, page string, userName string, userAddress string, userAgent string) error {
	event := &pb.Event{
		Id:        uuid.New().String(),
		Timestamp: uint64(util.NowMs()),
		EventType: pb.EventType_ET_PAGE_SERVED,
		User1:     userName,
		Message:   "Web page was served",
		Component: fmt.Sprintf("%s/%s", s.component, page),
		User2:     userAddress,
		Data1:     userAgent}
	return marshalAndSend(ctx, s.privateKey, s.sender, s.target, s.lookupClient, event)
}

func marshalAndSend(ctx context.Context, privateKey *easyecc.PrivateKey, sender, target string,
	lookupClient pb.LookupServiceClient, event *pb.Event) error {
	b, err := proto.Marshal(event)
	if err != nil {
		return err
	}
	err = protoutil.SendMessage(ctx, privateKey, b, sender, target, lookupClient)
	if err != nil {
		return err
	}
	return err
}
