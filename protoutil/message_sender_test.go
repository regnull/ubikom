package protoutil

import (
	"context"
	"testing"
	"time"

	"github.com/regnull/easyecc/v2"
	bcmocks "github.com/regnull/ubikom/bc/mocks"
	"github.com/regnull/ubikom/pb"
	pbmocks "github.com/regnull/ubikom/pb/mocks"
	pumocks "github.com/regnull/ubikom/protoutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

func Test_MessageSender(t *testing.T) {
	assert := assert.New(t)

	bchain := new(bcmocks.MockBlockchain)
	dscfactory := new(pumocks.MockDumpServiceClientFactory)
	dsclient := new(pbmocks.MockDMSDumpServiceClient)

	privateKey, err := easyecc.NewPrivateKey(easyecc.P256)
	assert.NoError(err)

	receiverPrivateKey, err := easyecc.NewPrivateKey(easyecc.P256)
	assert.NoError(err)

	ctx := context.Background()
	bchain.EXPECT().PublicKeyByCurve(ctx, "bob",
		easyecc.P256).Return(receiverPrivateKey.PublicKey(), nil)
	bchain.EXPECT().Endpoint(ctx, "bob").Return("bob's endpoint", nil)
	dscfactory.EXPECT().CreateDumpServiceClient(ctx, "bob's endpoint", time.Duration(0)).Return(dsclient, nil, nil)
	dsclient.EXPECT().Send(ctx, mock.Anything).RunAndReturn(
		func(ctx context.Context, req *pb.SendRequest,
			opts ...grpc.CallOption) (*pb.SendResponse, error) {
			assert.Equal("alice", req.GetMessage().GetSender())
			assert.Equal("bob", req.GetMessage().GetReceiver())
			assert.True(VerifySignature(req.GetMessage().GetSignature(), privateKey.PublicKey(), req.GetMessage().GetContent()))
			content, err := receiverPrivateKey.Decrypt(req.GetMessage().GetContent(), privateKey.PublicKey())
			assert.NoError(err)
			assert.Equal("the message", string(content))
			return &pb.SendResponse{}, nil
		})
	sender := NewMessageSender(dscfactory, bchain)
	err = sender.Send(ctx, privateKey, []byte("the message"), "alice", "bob")
	assert.NoError(err)

	bchain.AssertExpectations(t)
	dscfactory.AssertExpectations(t)
	dsclient.AssertExpectations(t)
}
