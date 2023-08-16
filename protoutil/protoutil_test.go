package protoutil

import (
	"testing"
	"time"

	"github.com/regnull/easyecc/v2"
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
			want: pb.EllipticCurve_EC_SECP256P1,
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
