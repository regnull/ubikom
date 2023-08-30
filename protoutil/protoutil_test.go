package protoutil

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/regnull/easyecc/v2"
	bcmocks "github.com/regnull/ubikom/bc/mocks"
	"github.com/regnull/ubikom/pb"
	"github.com/stretchr/testify/assert"
)

func Test_CreateSigned(t *testing.T) {
	assert := assert.New(t)

	key, err := easyecc.NewPrivateKey(easyecc.SECP256K1)
	assert.NoError(err)
	content := []byte("something to be signed")
	signed, err := CreateSigned(key, content)
	assert.NoError(err)
	assert.NotNil(signed)

	assert.True(VerifySignature(signed.Signature, key.PublicKey(), content))

	// Let's mess with the content.
	content = []byte("something to be signed xyz")
	assert.False(VerifySignature(signed.Signature, key.PublicKey(), content))
}

func Test_VerifyIdentity(t *testing.T) {
	assert := assert.New(t)

	key, err := easyecc.NewPrivateKey(easyecc.SECP256K1)
	assert.NoError(err)
	ts := time.Now()
	signed, err := IdentityProof(key, ts)
	assert.NoError(err)
	assert.NoError(VerifyIdentity(signed, ts, 10.0, easyecc.SECP256K1))

	ts1 := ts.Add(time.Minute)
	assert.Error(VerifyIdentity(signed, ts1, 10.0, easyecc.SECP256K1))
}

func Test_CurveToProto(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		args easyecc.EllipticCurve
		want pb.EllipticCurve
	}{
		{
			args: easyecc.SECP256K1,
			want: pb.EllipticCurve_EC_SECP256K1,
		},
		{
			args: easyecc.P256,
			want: pb.EllipticCurve_EC_P_256,
		},
		{
			args: easyecc.P384,
			want: pb.EllipticCurve_EC_P_384,
		},
		{
			args: easyecc.P521,
			want: pb.EllipticCurve_EC_P_521,
		},
		{
			args: easyecc.INVALID_CURVE,
			want: pb.EllipticCurve_EC_UNKNOWN,
		},
	}

	for _, tt := range tests {
		assert.Equal(tt.want, CurveToProto(tt.args))
	}
}

func Test_CurveFromProto(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		args pb.EllipticCurve
		want easyecc.EllipticCurve
	}{
		{
			args: pb.EllipticCurve_EC_UNKNOWN,
			want: easyecc.SECP256K1,
		},
		{
			args: pb.EllipticCurve_EC_SECP256K1,
			want: easyecc.SECP256K1,
		},
		{
			args: pb.EllipticCurve_EC_P_256,
			want: easyecc.P256,
		},
		{
			args: pb.EllipticCurve_EC_P_384,
			want: easyecc.P384,
		},
		{
			args: pb.EllipticCurve_EC_P_521,
			want: easyecc.P521,
		},
		{
			args: pb.EllipticCurve(999),
			want: easyecc.INVALID_CURVE,
		},
	}

	for _, tt := range tests {
		assert.Equal(tt.want, CurveFromProto(tt.args))
	}
}

func Test_CreateMessage(t *testing.T) {
	assert := assert.New(t)

	privateKey, err := easyecc.NewPrivateKey(easyecc.P521)
	assert.NoError(err)

	receiverKey, err := easyecc.NewPrivateKey(easyecc.P521)
	assert.NoError(err)

	message := []byte("All experience is preceded by mind")
	msg, err := CreateMessage(privateKey, message, "alice", "bob", receiverKey.PublicKey())
	assert.NoError(err)
	assert.NotNil(msg)

	assert.Equal("alice", msg.GetSender())
	assert.Equal("bob", msg.GetReceiver())

	assert.True(len(msg.GetContent()) > 10)
	assert.True(VerifySignature(msg.GetSignature(), privateKey.PublicKey(), msg.GetContent()))

	assert.Equal(pb.EllipticCurve_EC_P_521, msg.GetCryptoContext().GetEllipticCurve())
	assert.EqualValues(2, msg.GetCryptoContext().GetEcdhVersion())
	assert.EqualValues(1, msg.GetCryptoContext().GetEcdsaVersion())
}

func Test_CreateLegacyMessage(t *testing.T) {
	assert := assert.New(t)

	privateKey, err := easyecc.NewPrivateKey(easyecc.SECP256K1)
	assert.NoError(err)

	receiverKey, err := easyecc.NewPrivateKey(easyecc.SECP256K1)
	assert.NoError(err)

	message := []byte("All experience is preceded by mind")
	msg, err := CreateLegacyMessage(privateKey, message, "alice", "bob", receiverKey.PublicKey())
	assert.NoError(err)
	assert.NotNil(msg)

	assert.Equal("alice", msg.GetSender())
	assert.Equal("bob", msg.GetReceiver())

	assert.True(len(msg.GetContent()) > 10)
	assert.True(VerifySignature(msg.GetSignature(), privateKey.PublicKey(), msg.GetContent()))

	assert.Equal(pb.EllipticCurve_EC_SECP256K1, msg.GetCryptoContext().GetEllipticCurve())
	assert.EqualValues(1, msg.GetCryptoContext().GetEcdhVersion())
	assert.EqualValues(1, msg.GetCryptoContext().GetEcdsaVersion())

	privateKeyP521, err := easyecc.NewPrivateKey(easyecc.P521)
	assert.NoError(err)

	_, err = CreateLegacyMessage(privateKeyP521, message, "alice", "bob", receiverKey.PublicKey())
	assert.Equal(ErrUnsupportedCurve, err)
}

func Test_DecryptMessage(t *testing.T) {
	assert := assert.New(t)

	curves := []easyecc.EllipticCurve{easyecc.SECP256K1, easyecc.P256, easyecc.P384, easyecc.P521}

	ctx := context.Background()
	bchain := new(bcmocks.MockBlockchain)

	message := []byte("All experience is preceded by mind")

	for _, curve := range curves {
		privateKey, err := easyecc.NewPrivateKey(curve)
		assert.NoError(err)

		recipientKey, err := easyecc.NewPrivateKey(curve)
		assert.NoError(err)

		bchain.EXPECT().PublicKeyByCurve(ctx, "alice",
			curve).Return(privateKey.PublicKey(), nil)

		msg, err := CreateMessage(privateKey, message, "alice", "bob", recipientKey.PublicKey())
		assert.NoError(err)

		content, err := DecryptMessage(ctx, bchain, recipientKey, msg)
		assert.NoError(err)
		assert.True(bytes.Equal(message, []byte(content)))

		// Try to mess with the message.
		msg.Content[0] = 0x66
		_, err = DecryptMessage(ctx, bchain, recipientKey, msg)
		// Signature verification must fail.
		assert.Error(err)
	}

	bchain.AssertExpectations(t)
}
