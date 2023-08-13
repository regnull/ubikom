package protoutil

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/big"
	"time"

	easyeccv1 "github.com/regnull/easyecc"
	"github.com/regnull/easyecc/v2"

	"github.com/regnull/ubikom/mail"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

var (
	ErrFailedToSignMessage         = errors.New("failed to sign message")
	ErrSignatureVerificationFailed = errors.New("signature verification failed")
	ErrTimeDifferenceTooLarge      = errors.New("time difference is too large")
)

// CreateSigned creates a signature for the given content.
func CreateSigned(privateKey *easyecc.PrivateKey, content []byte) (*pb.Signed, error) {
	compressedKey := privateKey.PublicKey().CompressedBytes()
	hash := util.Hash256(content)
	sig, err := privateKey.Sign(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign request, %w", err)
	}

	req := &pb.Signed{
		Content: content,
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
	key, err := easyecc.NewPublicKeyFromCompressedBytes(easyecc.SECP256K1, serializedKey)
	if err != nil {
		log.Printf("invalid serialized compressed key")
		return false
	}

	eccSig := &easyecc.Signature{
		R: new(big.Int).SetBytes(sig.R),
		S: new(big.Int).SetBytes(sig.S)}

	if !eccSig.Verify(key, util.Hash256(content)) {
		log.Printf("signature verification failed")
		return false
	}
	return true
}

// CreateMessages creates a new DMSMessage, signed and encrypted.
func CreateMessage(privateKey *easyecc.PrivateKey, body []byte, sender, receiver string,
	receiverKey *easyecc.PublicKey) (*pb.DMSMessage, error) {
	encryptedBody, err := privateKey.Encrypt(body, receiverKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt message: %w", err)
	}

	hash := util.Hash256(encryptedBody)
	sig, err := privateKey.Sign(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign message, %w", err)
	}

	return &pb.DMSMessage{
		Sender:   sender,
		Receiver: receiver,
		Content:  encryptedBody,
		Signature: &pb.Signature{
			R: sig.R.Bytes(),
			S: sig.S.Bytes(),
		},
		CryptoContext: &pb.CryptoContext{
			EllipticCurve: getProtoEC(privateKey.Curve()),
			EcdhVersion:   2,
			EcdsaVersion:  1,
		},
	}, nil
}

func CreateLegacyMessage(privateKey *easyecc.PrivateKey, body []byte, sender, receiver string,
	receiverKey *easyecc.PublicKey) (*pb.DMSMessage, error) {
	// We must use easyecc v1 for the legacy encryption.
	privateKeyV1 := easyeccv1.CreatePrivateKey(easyeccv1.SECP256K1, privateKey.Secret())
	receiverKeyV1, err := easyeccv1.NewPublicFromSerializedCompressed(receiverKey.CompressedBytes())
	if err != nil {
		return nil, err
	}
	encryptedBody, err := privateKeyV1.Encrypt(body, receiverKeyV1)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt message: %w", err)
	}

	hash := util.Hash256(encryptedBody)
	sig, err := privateKey.Sign(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign message, %w", err)
	}

	return &pb.DMSMessage{
		Sender:   sender,
		Receiver: receiver,
		Content:  encryptedBody,
		Signature: &pb.Signature{
			R: sig.R.Bytes(),
			S: sig.S.Bytes(),
		},
		CryptoContext: &pb.CryptoContext{
			EllipticCurve: getProtoEC(privateKey.Curve()),
			EcdhVersion:   1,
			EcdsaVersion:  1,
		},
	}, nil
}

// SendEmail adds Ubikom headers to the email message and sends it.
func SendEmail(ctx context.Context, privateKey *easyecc.PrivateKey, body []byte,
	sender, receiver string, lookupService pb.LookupServiceClient) error {
	withHeaders, err := mail.AddUbikomHeaders(ctx, string(body), sender, receiver,
		privateKey.PublicKey(), lookupService)
	if err != nil {
		return err
	}
	return SendMessage(ctx, privateKey, []byte(withHeaders), sender, receiver, lookupService)
}

// SendMessage creates a new DMSMessage and sends it out to the appropriate address.
func SendMessage(ctx context.Context, privateKey *easyecc.PrivateKey, body []byte,
	sender, receiver string, lookupService pb.LookupServiceClient) error {

	// TODO: Pass timeout as an argument.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Second * 5),
	}

	// Get receiver's public key.
	lookupRes, err := lookupService.LookupName(ctx, &pb.LookupNameRequest{Name: receiver})
	if err != nil {
		return fmt.Errorf("failed to get receiver public key: %w", err)
	}
	receiverKey, err := easyecc.NewPublicKeyFromCompressedBytes(easyecc.SECP256K1, lookupRes.GetKey())
	if err != nil {
		log.Fatal().Err(err).Msg("invalid receiver public key")
	}

	log.Debug().Msg("got receiver's public key")

	// Get receiver's address.
	addressLookupRes, err := lookupService.LookupAddress(ctx,
		&pb.LookupAddressRequest{Name: receiver, Protocol: pb.Protocol_PL_DMS})
	if err != nil {
		return fmt.Errorf("failed to get receiver's address: %w", err)
	}

	log.Debug().Str("address", addressLookupRes.GetAddress()).Msg("got receiver's address")

	dumpConn, err := grpc.Dial(addressLookupRes.GetAddress(), opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to the dump server: %w", err)
	}
	defer dumpConn.Close()

	msg, err := CreateMessage(privateKey, body, sender, receiver, receiverKey)
	if err != nil {
		return err
	}

	client := pb.NewDMSDumpServiceClient(dumpConn)
	_, err = client.Send(ctx, &pb.SendRequest{Message: msg})
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	log.Debug().Msg("sent message successfully")
	return nil
}

func DecryptMessage(ctx context.Context, lookupClient pb.LookupServiceClient,
	privateKey *easyecc.PrivateKey, msg *pb.DMSMessage) (string, error) {
	lookupRes, err := lookupClient.LookupName(ctx, &pb.LookupNameRequest{Name: msg.GetSender()})
	if err != nil {
		return "", fmt.Errorf("failed to get sender public key: %w", err)
	}
	senderKey, err := easyecc.NewPublicKeyFromCompressedBytes(easyecc.SECP256K1, lookupRes.GetKey())
	if err != nil {
		return "", fmt.Errorf("invalid sender public key: %w", err)
	}

	if !VerifySignature(msg.GetSignature(), lookupRes.GetKey(), msg.GetContent()) {
		return "", fmt.Errorf("signature verification failed")
	}

	content, err := privateKey.Decrypt(msg.Content, senderKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt message")
	}
	return string(content), nil
}

// IdentityProof generates an identity proof that can be used in receive requests.
func IdentityProof(key *easyecc.PrivateKey, timestamp time.Time) (*pb.Signed, error) {
	ts := timestamp.UTC().Unix()
	log.Debug().Int64("timestamp", ts).Msg("POI setting timestamp")
	var buf [8]byte
	binary.PutVarint(buf[:], ts)
	hash := util.Hash256(buf[:])
	sig, err := key.Sign(hash)
	if err != nil {
		return nil, ErrFailedToSignMessage
	}

	signed := &pb.Signed{
		Content: buf[:],
		Signature: &pb.Signature{
			R: sig.R.Bytes(),
			S: sig.S.Bytes(),
		},
		Key: key.PublicKey().CompressedBytes(),
	}
	return signed, nil
}

// VerifyIdentity returns no error if the signed has the correct signature
// and if it was signed within 10 seconds from now.
func VerifyIdentity(signed *pb.Signed, now time.Time, allowedDeltaSeconds float64) error {
	if !VerifySignature(signed.Signature, signed.Key, signed.Content) {
		return ErrSignatureVerificationFailed
	}

	ts, err := binary.ReadVarint(bytes.NewReader(signed.Content))
	if err != nil {
		return err
	}
	log.Debug().Int64("timestamp", ts).Msg("POI got timestamp")

	d := now.UTC().Unix() - ts
	log.Debug().Int64("time_delta", d).Msg("time delta")
	if math.Abs(float64(d)) > allowedDeltaSeconds {
		return ErrTimeDifferenceTooLarge
	}
	return nil
}

func getProtoEC(curve easyecc.EllipticCurve) pb.EllipticCurve {
	switch curve {
	case easyecc.SECP256K1:
		return pb.EllipticCurve_EC_SECP256P1
	case easyecc.P256:
		return pb.EllipticCurve_EC_P_256
	case easyecc.P384:
		return pb.EllipticCurve_EC_P_384
	case easyecc.P521:
		return pb.EllipticCurve_EC_P_521
	default:
		return pb.EllipticCurve_EC_UNKNOWN
	}
}
