package server

import (
	"context"
	"testing"
	"time"

	"github.com/regnull/easyecc/v2"
	bcmocks "github.com/regnull/ubikom/bc/mocks"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/protoutil"
	"github.com/regnull/ubikom/store"
	"github.com/stretchr/testify/assert"
)

func Test_DumpServer_SendReceive(t *testing.T) {
	assert := assert.New(t)

	dumpStore := store.NewMemory()
	bchain := new(bcmocks.MockBlockchain)
	ctx := context.Background()
	dumpServer := NewDumpServer(dumpStore, bchain)

	aliceKey, err := easyecc.NewPrivateKey(easyecc.P256)
	assert.NoError(err)
	bobKey, err := easyecc.NewPrivateKey(easyecc.P256)
	assert.NoError(err)

	bchain.EXPECT().PublicKeyByCurve(ctx, "alice",
		easyecc.P256).Return(aliceKey.PublicKey(), nil)
	bchain.EXPECT().PublicKeyByCurve(ctx, "bob",
		easyecc.P256).Return(bobKey.PublicKey(), nil)

	msg, err := protoutil.CreateMessage(aliceKey, []byte("hi bob"), "alice", "bob", bobKey.PublicKey())
	assert.NoError(err)

	sendRes, err := dumpServer.Send(ctx, &pb.SendRequest{Message: msg})
	assert.NoError(err)
	assert.NotNil(sendRes)

	identityProof, err := protoutil.IdentityProof(bobKey, time.Now())
	assert.NoError(err)
	req := &pb.ReceiveRequest{
		IdentityProof: identityProof,
		CryptoContext: &pb.CryptoContext{
			EllipticCurve: pb.EllipticCurve(easyecc.P256),
			EcdhVersion:   2,
			EcdsaVersion:  1,
		},
	}
	receiveRes, err := dumpServer.Receive(ctx, req)
	assert.NoError(err)
	assert.NotNil(receiveRes)
	assert.Equal("alice", receiveRes.GetMessage().GetSender())
	assert.Equal("bob", receiveRes.GetMessage().GetReceiver())

	content, err := protoutil.DecryptMessage(ctx, bchain, bobKey, receiveRes.GetMessage())
	assert.NoError(err)
	assert.Equal("hi bob", content)

	bchain.AssertExpectations(t)
}

func Test_DumpServer_SendReceiveLegacy(t *testing.T) {
	assert := assert.New(t)

	dumpStore := store.NewMemory()
	bchain := new(bcmocks.MockBlockchain)
	ctx := context.Background()
	dumpServer := NewDumpServer(dumpStore, bchain)

	aliceKey, err := easyecc.NewPrivateKey(easyecc.SECP256K1)
	assert.NoError(err)
	bobKey, err := easyecc.NewPrivateKey(easyecc.SECP256K1)
	assert.NoError(err)

	bchain.EXPECT().PublicKeyByCurve(ctx, "alice",
		easyecc.SECP256K1).Return(aliceKey.PublicKey(), nil)
	bchain.EXPECT().PublicKeyByCurve(ctx, "bob",
		easyecc.SECP256K1).Return(bobKey.PublicKey(), nil)

	msg, err := protoutil.CreateLegacyMessage(aliceKey, []byte("hi bob"), "alice", "bob", bobKey.PublicKey())
	assert.NoError(err)

	sendRes, err := dumpServer.Send(ctx, &pb.SendRequest{Message: msg})
	assert.NoError(err)
	assert.NotNil(sendRes)

	identityProof, err := protoutil.IdentityProof(bobKey, time.Now())
	assert.NoError(err)
	req := &pb.ReceiveRequest{
		IdentityProof: identityProof,
	}
	receiveRes, err := dumpServer.Receive(ctx, req)
	assert.NoError(err)
	assert.NotNil(receiveRes)
	assert.Equal("alice", receiveRes.GetMessage().GetSender())
	assert.Equal("bob", receiveRes.GetMessage().GetReceiver())

	content, err := protoutil.DecryptLegacyMessage(ctx, bchain, bobKey, receiveRes.GetMessage())
	assert.NoError(err)
	assert.Equal("hi bob", content)

	bchain.AssertExpectations(t)
}

func Test_DumpServer_IdentityFallback(t *testing.T) {
	// This test should go away after the identity proof is working as expected.
	assert := assert.New(t)

	dumpStore := store.NewMemory()
	bchain := new(bcmocks.MockBlockchain)
	ctx := context.Background()
	dumpServer := NewDumpServer(dumpStore, bchain)

	aliceKey, err := easyecc.NewPrivateKey(easyecc.P256)
	assert.NoError(err)
	bobKey, err := easyecc.NewPrivateKey(easyecc.P256)
	assert.NoError(err)

	bchain.EXPECT().PublicKeyByCurve(ctx, "alice",
		easyecc.P256).Return(aliceKey.PublicKey(), nil)
	bchain.EXPECT().PublicKeyByCurve(ctx, "bob",
		easyecc.P256).Return(bobKey.PublicKey(), nil)

	msg, err := protoutil.CreateMessage(aliceKey, []byte("hi bob"), "alice", "bob", bobKey.PublicKey())
	assert.NoError(err)

	sendRes, err := dumpServer.Send(ctx, &pb.SendRequest{Message: msg})
	assert.NoError(err)
	assert.NotNil(sendRes)

	// Intentionally send old proof of identity.
	identityProof, err := protoutil.IdentityProof(bobKey, time.Now().Add(-time.Second*100))
	assert.NoError(err)
	req := &pb.ReceiveRequest{
		IdentityProof: identityProof,
		CryptoContext: &pb.CryptoContext{
			EllipticCurve: pb.EllipticCurve(easyecc.P256),
			EcdhVersion:   2,
			EcdsaVersion:  1,
		},
	}
	receiveRes, err := dumpServer.Receive(ctx, req)
	assert.NoError(err)
	assert.NotNil(receiveRes)
	assert.Equal("alice", receiveRes.GetMessage().GetSender())
	assert.Equal("bob", receiveRes.GetMessage().GetReceiver())

	content, err := protoutil.DecryptMessage(ctx, bchain, bobKey, receiveRes.GetMessage())
	assert.NoError(err)
	assert.Equal("hi bob", content)

	bchain.AssertExpectations(t)
}
