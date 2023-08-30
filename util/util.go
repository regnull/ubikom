package util

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"os"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/pb"
	"golang.org/x/crypto/ripemd160"
	"golang.org/x/term"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	allowedChars      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"
	defaultHomeSubDir = ".ubikom"
	defaultKeyFile    = "key"
)

// NowMs returns current time as milliseconds from epoch.
func NowMs() int64 {
	return time.Now().UnixNano() / 1000000
}

// TimeFromMs returns time given milliseconds from epoch.
func TimeFromMs(ts int64) time.Time {
	return time.Unix(ts/1000, (ts%1000)*1000000)
}

// Hash256 does two rounds of SHA256 hashing.
func Hash256(data []byte) []byte {
	h := sha256.Sum256(data)
	h1 := sha256.Sum256(h[:])
	return h1[:]
}

// Calculate the hash of hasher over buf.
func calcHash(buf []byte, hasher hash.Hash) []byte {
	hasher.Write(buf)
	return hasher.Sum(nil)
}

// Hash160 calculates the hash ripemd160(sha256(b)).
func Hash160(buf []byte) []byte {
	return calcHash(calcHash(buf, sha256.New()), ripemd160.New())
}

// ValidateName returns true if the name is valid.
func ValidateName(name string) bool {
	if len(name) < 3 || len(name) > 64 {
		return false
	}

	for _, c := range name {
		if !strings.ContainsRune(allowedChars, c) {
			return false
		}
	}
	return true
}

func GetDefaultKeyLocation() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory, %w", err)
	}
	dir := path.Join(homeDir, defaultHomeSubDir)
	keyFile := path.Join(dir, defaultKeyFile)
	return keyFile, nil
}

// StatusCodeFromError returns gRPC status code from error, or codes.Unknown if the error does
// not contain gRPC code.
func StatusCodeFromError(err error) codes.Code {
	if err == nil {
		return codes.OK
	}
	s, ok := status.FromError(err)
	if !ok {
		return codes.Unknown
	}
	return s.Code()
}

// ErrEqualCode returns true if the given error has the specified code (gRPC status code).
func ErrEqualCode(err error, code codes.Code) bool {
	return StatusCodeFromError(err) == code
}

// GetConfigFromArgs returns the value of --config flag from the arguments.
func GetConfigFromArgs(args []string) string {
	for i, arg := range args {
		if arg == "--config" {
			if len(args) <= i+1 {
				return ""
			}
			return args[i+1]
		} else if strings.HasPrefix(arg, "--config=") {
			parts := strings.Split(arg, "=")
			if len(parts) != 2 {
				return ""
			}
			return parts[1]
		}
	}
	return ""
}

func EnterPassphrase() (string, error) {
	fmt.Print("Passphrase (enter for none): ")
	bytePassphrase, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("failed to read passphrase")
	}
	passphrase1 := string(bytePassphrase)

	fmt.Print("\nConfirm passphrase (enter for none): ")
	bytePassphrase, err = term.ReadPassword(int(syscall.Stdin))
	fmt.Print("\n")
	if err != nil {
		return "", fmt.Errorf("failed to read passphrase")
	}
	passphrase2 := string(bytePassphrase)
	if passphrase1 != passphrase2 {
		return "", fmt.Errorf("passphrase mismatch")
	}
	return passphrase1, nil
}

func ReadPassphase() (string, error) {
	fmt.Print("Passphrase: ")
	bytePassphrase, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Print("\n")
	if err != nil {
		return "", fmt.Errorf("failed to read passphrase")
	}
	passphrase := string(bytePassphrase)
	return passphrase, nil
}

func IsKeyEncrypted(filePath string) (bool, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}
	var i interface{}
	err = json.Unmarshal(b, &i)
	if err != nil {
		return false, err
	}
	m, ok := i.(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("invalid content")
	}
	_, ok = m["ciphertext"]
	if ok {
		return true, nil
	}
	return false, nil
}

func StripDomainName(s string) string {
	n := strings.Trim(s, " ")

	// Strip @domain from the string, if any.
	if i := strings.Index(n, "@"); i != -1 {
		n = n[:i]
	}
	return n
}

func NowUint32() uint32 {
	return uint32(time.Now().UnixNano() / 1000000000)
}

// ContainsFlag return true if the flags slice contains the specified flag.
func ContainsFlag(flags []string, flag string) bool {
	for _, f := range flags {
		if f == flag {
			return true
		}
	}
	return false
}

// Clear flag clears the specified flag from the flag set.
func ClearFlag(flags []string, flag string) []string {
	var newFlags []string
	for _, f := range flags {
		if f == flag {
			continue
		}
		newFlags = append(newFlags, f)
	}
	return newFlags
}

// FileNameNoExtension returns file name without extension.
func FileNameNoExtension(filePath string) string {
	b := path.Base(filePath)
	if pos := strings.LastIndexByte(b, '.'); pos != -1 {
		return b[:pos]
	}
	return b
}

func HexStringToBytes(s string) ([]byte, error) {
	return hex.DecodeString(strings.TrimPrefix(s, "0x"))
}

func EqualHexStrings(s1, s2 string) bool {
	return strings.EqualFold(strings.TrimPrefix(s1, "0x"), strings.TrimPrefix(s2, "0x"))
}

// FixName processes user-provided name to obtain a valid blockchain-registered name.
func FixName(name string) string {
	n := strings.TrimSpace(name)
	n = StripDomainName(n)
	n = strings.ToLower(n)
	return n
}

// CheckUserNameAndPassword verifies user name and password by generating a public key and comparing it with
// the key in the identity registry.
func CheckUserNameAndPassword(ctx context.Context, name, password string, lookupClient pb.LookupServiceClient) error {
	privateKey := GetPrivateKeyFromNameAndPassword(name, password)
	res, err := lookupClient.LookupName(ctx, &pb.LookupNameRequest{Name: FixName(name)})
	if err != nil {
		return fmt.Errorf("failed to lookup name: %w", err)
	}
	if !bytes.Equal(res.GetKey(), privateKey.PublicKey().SerializeCompressed()) {
		return fmt.Errorf("keys do not match")
	}
	return nil
}

func GetPrivateKeyFromNameAndPassword(name, password string) *easyecc.PrivateKey {
	n := strings.TrimSpace(name)
	n = StripDomainName(n)
	n = strings.ToLower(n)
	return easyecc.NewPrivateKeyFromPassword([]byte(password),
		Hash256([]byte(strings.ToLower(n))))
}

// LoadKey reads a private key from file, asking for the passphrase, if needed.
func LoadKey(keyFile string) (*easyecc.PrivateKey, error) {
	var err error
	if keyFile == "" {
		keyFile, err = GetDefaultKeyLocation()
		if err != nil {
			return nil, err
		}
	}

	encrypted, err := IsKeyEncrypted(keyFile)
	if err != nil {
		return nil, err
	}

	passphrase := ""
	if encrypted {
		passphrase, err = ReadPassphase()
		if err != nil {
			return nil, err
		}
	}

	privateKey, err := easyecc.NewPrivateKeyFromFile(keyFile, passphrase)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}
