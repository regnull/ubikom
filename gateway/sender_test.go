package gateway

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/gateway/mocks"
	"github.com/regnull/ubikom/pb"
	pbmocks "github.com/regnull/ubikom/pb/mocks"
	"github.com/regnull/ubikom/protoutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_Sender_Run(t *testing.T) {
	assert := assert.New(t)

	externalSender := new(mocks.ExternalSender)
	defer externalSender.AssertExpectations(t)
	lookupClient := new(pbmocks.LookupServiceClient)
	defer lookupClient.AssertExpectations(t)
	dumpClient := new(pbmocks.DMSDumpServiceClient)
	defer dumpClient.AssertExpectations(t)

	spongebobKey, err := easyecc.NewRandomPrivateKey()
	assert.NoError(err)
	gatewayKey, err := easyecc.NewRandomPrivateKey()
	assert.NoError(err)

	ctx, cancel := context.WithCancel(context.Background())

	content1 := "To: Patrick Star <patrick@gmail.com>\n" +
		"From: Spongebob <spongebob@x>\n" +
		"Subject: testing\n\n" +
		"Hello, this is message 1"

	content2 := "To: Sandy <sandy@gmail.com>\n" +
		"From: Spongebob <spongebob@x>\n" +
		"Subject: testing\n\n" +
		"Hello, this is message 2"

	msg1, err := protoutil.CreateMessage(
		spongebobKey,
		[]byte(content1),
		"spongebob",
		"gateway",
		gatewayKey.PublicKey())
	assert.NoError(err)

	msg2, err := protoutil.CreateMessage(
		spongebobKey,
		[]byte(content2),
		"spongebob",
		"gateway",
		gatewayKey.PublicKey())
	assert.NoError(err)

	lookupClient.On("LookupName", ctx, &pb.LookupNameRequest{
		Name: "spongebob"}).
		Return(&pb.LookupNameResponse{
			Key: spongebobKey.PublicKey().SerializeCompressed(),
		}, nil).Twice()

	// Return two messages, then return NotFound.
	dumpClient.On("Receive", ctx, mock.Anything).
		Run(func(args mock.Arguments) {
			verifyRequest(args, gatewayKey, assert)
		}).
		Return(
			&pb.ReceiveResponse{
				Message: msg1,
			}, nil).Once()

	dumpClient.On("Receive", ctx, mock.Anything).
		Run(func(args mock.Arguments) {
			verifyRequest(args, gatewayKey, assert)
		}).
		Return(
			&pb.ReceiveResponse{
				Message: msg2,
			}, nil).Once()

	dumpClient.On("Receive", ctx, mock.Anything).
		Run(func(args mock.Arguments) {
			verifyRequest(args, gatewayKey, assert)
			// Cancel the context to exit the receive loop.
			cancel()
		}).
		Return(
			nil, status.Error(codes.NotFound, "no more messages")).Once()

	// Make sure the right emails get sent out.
	externalSender.On("Send", "spongebob@ubikom.cc",
		strings.Replace(content1, "spongebob@x", "spongebob@ubikom.cc", 1)).Return(nil).Once()
	externalSender.On("Send", "spongebob@ubikom.cc",
		strings.Replace(content2, "spongebob@x", "spongebob@ubikom.cc", 1)).Return(nil).Once()

	senderOpts := &SenderOptions{
		PrivateKey:             gatewayKey,
		LookupClient:           lookupClient,
		DumpClient:             dumpClient,
		GlobalRateLimitPerHour: 10000,
		PollInterval:           100 * time.Millisecond,
		ExternalSender:         externalSender,
	}
	sender := NewSender(senderOpts)
	err = sender.Run(ctx)
	assert.Equal(fmt.Errorf("context done"), err)
}

func verifyRequest(args mock.Arguments, gatewayKey *easyecc.PrivateKey, assert *assert.Assertions) {
	req := args[1].(*pb.ReceiveRequest)
	assert.True(protoutil.VerifySignature(req.IdentityProof.Signature,
		gatewayKey.PublicKey().SerializeCompressed(),
		req.IdentityProof.Content))
}
