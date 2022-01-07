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

	"github.com/regnull/easyecc"

	"github.com/regnull/ubikom/mail"
	"github.com/regnull/ubikom/pb"
	"github.com/regnull/ubikom/pow"
	"github.com/regnull/ubikom/util"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

var (
	ErrFailedToSignMessage         = errors.New("failed to sign message")
	ErrSignatureVerificationFailed = errors.New("signature verification failed")
	ErrTimeDifferenceTooLarge      = errors.New("time difference is too large")
)

// CreateSignedWithPOW creates a request signed with the given private key and generates POW of the given strength.
func CreateSignedWithPOW(privateKey *easyecc.PrivateKey, content []byte, powStrength int) (*pb.SignedWithPow, error) {
	//compressedKey := privateKey.PublicKey().SerializeCompressed()

	log.Debug().Msg("generating POW...")
	reqPow := pow.Compute(content, powStrength)
	log.Debug().Hex("pow", reqPow).Msg("POW found")

	signed, err := CreateSigned(privateKey, content)
	if err != nil {
		return nil, err
	}

	req := &pb.SignedWithPow{
		Content:   content,
		Pow:       reqPow,
		Signature: signed.Signature,
		Key:       signed.Key,
	}
	return req, nil
}

// CreateSigned creates a signature for the given content.
func CreateSigned(privateKey *easyecc.PrivateKey, content []byte) (*pb.Signed, error) {
	compressedKey := privateKey.PublicKey().SerializeCompressed()
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
	key, err := easyecc.NewPublicFromSerializedCompressed(serializedKey)
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
	receiverKey, err := easyecc.NewPublicFromSerializedCompressed(lookupRes.GetKey())
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

func DecryptMessage(ctx context.Context, lookupClient pb.LookupServiceClient, privateKey *easyecc.PrivateKey, msg *pb.DMSMessage) (string, error) {
	lookupRes, err := lookupClient.LookupName(ctx, &pb.LookupNameRequest{Name: msg.GetSender()})
	if err != nil {
		return "", fmt.Errorf("failed to get sender public key: %w", err)
	}
	senderKey, err := easyecc.NewPublicFromSerializedCompressed(lookupRes.GetKey())
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
	ts := timestamp.Unix()
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
		Key: key.PublicKey().SerializeCompressed(),
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

	d := now.Unix() - ts
	if math.Abs(float64(d)) > allowedDeltaSeconds {
		return ErrTimeDifferenceTooLarge
	}
	return nil
}
