package protoutil

import (
	"fmt"
	"math/big"

	"github.com/rs/zerolog/log"
	"teralyt.com/ubikom/ecc"
	"teralyt.com/ubikom/pb"
	"teralyt.com/ubikom/pow"
	"teralyt.com/ubikom/util"
)

// CreateSignedWithPOW creates a request signed with the given private key and generates POW of the given strength.
func CreateSignedWithPOW(privateKey *ecc.PrivateKey, content []byte, powStrength int) (*pb.SignedWithPow, error) {
	compressedKey := privateKey.PublicKey().SerializeCompressed()

	log.Debug().Msg("generating POW...")
	reqPow := pow.Compute(content, powStrength)
	log.Debug().Hex("pow", reqPow).Msg("POW found")

	hash := util.Hash256(content)
	sig, err := privateKey.Sign(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign request, %w", err)
	}

	req := &pb.SignedWithPow{
		Content: content,
		Pow:     reqPow,
		Signature: &pb.Signature{
			R: sig.R.Bytes(),
			S: sig.S.Bytes(),
		},
		Key: compressedKey,
	}
	return req, nil
}

// VerifySignature returns true if the provided signature is valid for the given key and content.
func VerifySignature(sig *pb.Signature, serializedKey []byte, content []byte) bool {
	key, err := ecc.NewPublicFromSerializedCompressed(serializedKey)
	if err != nil {
		log.Printf("invalid serialized compressed key")
		return false
	}

	eccSig := &ecc.Signature{
		R: new(big.Int).SetBytes(sig.R),
		S: new(big.Int).SetBytes(sig.S)}

	if !eccSig.Verify(key, util.Hash256(content)) {
		log.Printf("signature verification failed")
		return false
	}
	return true
}
