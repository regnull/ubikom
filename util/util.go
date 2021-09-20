package util

import (
	"context"
	"crypto/sha256"
	"fmt"
	"hash"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/pb"
	"golang.org/x/crypto/ripemd160"
	"golang.org/x/term"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v2"
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

func GetConfigFileLocation(location string) (string, error) {
	defaultFileName := "ubikom.conf"

	// If the location was explicitly specified, just use that.
	if location != "" {
		s, err := os.Stat(location)
		if err != nil {
			return "", fmt.Errorf("config file doesn't exist: %w", err)
		}
		if s.IsDir() {
			return "", fmt.Errorf("config file location points to a directory")
		}
		return location, nil
	}

	// Try current directory.
	if stat, err := os.Stat(defaultFileName); err == nil && !stat.IsDir() {
		return defaultFileName, nil
	}

	// Try the executable file location.
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err == nil {
		configPath := path.Join(dir, defaultFileName)
		if stat, err := os.Stat(configPath); err == nil && !stat.IsDir() {
			return configPath, nil
		}
	}

	if strings.HasSuffix(os.Args[0], "main") || strings.HasSuffix(os.Args[0], "__debug_bin") {
		// This process was likely started with "go run".
		// Check the config directory in the source directory tree.
		configPath := path.Join(dir, "..", "..", "config", defaultFileName)
		if stat, err := os.Stat(configPath); err == nil && !stat.IsDir() {
			return configPath, nil
		}

		wd, err := os.Getwd()
		if err == nil {
			configPath = path.Join(wd, "..", "..", "config", defaultFileName)
			if stat, err := os.Stat(configPath); err == nil && !stat.IsDir() {
				return configPath, nil
			}
		}
	}

	return "", fmt.Errorf("config file not found")
}

// FindAndParseConfigFile locates the config file and parses it, storing the result in out.
func FindAndParseConfigFile(configFile string, out interface{}) error {
	configFile, err := GetConfigFileLocation(configFile)
	if err != nil {
		return fmt.Errorf("config file not found: %w", err)
	}
	fmt.Printf("using config file: %s\n", configFile)
	config, err := ioutil.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("config file not found")
	}
	err = yaml.Unmarshal(config, out)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}
	return nil
}

// SerializedCompressedToAddress is a convenience function which converts
// serialized compressed representation of the private key to its address (which is shorter).
// If the key is invalid, the return string will contain an error message.
func SerializedCompressedToAddress(key []byte) string {
	publicKey, err := easyecc.NewPublicFromSerializedCompressed(key)
	if err != nil {
		return "**invalid key**"
	}
	return publicKey.Address()
}

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
	fi, err := os.Stat(filePath)
	if err != nil {
		return false, fmt.Errorf("file not found")
	}
	// get the size
	size := fi.Size()
	return size > 32, nil
}

// GetKeyFromNameAndPassword attempts to construct a private key from name and password and verify it with
// key lookup service.
func GetKeyFromNamePassword(ctx context.Context, name string, pass string,
	lookupClient pb.LookupServiceClient) (*easyecc.PrivateKey, error) {
	n := strings.Trim(name, " ")

	// Strip @domain from the string, if any.
	if i := strings.Index(n, "@"); i != -1 {
		n = n[:i]
	}

	// Try hash of the user name as salt first.
	privateKey := easyecc.NewPrivateKeyFromPassword([]byte(pass), Hash256([]byte(n)))
	res, err := lookupClient.LookupKey(ctx, &pb.LookupKeyRequest{
		Key: privateKey.PublicKey().SerializeCompressed()})
	if err == nil {
		if res.Disabled {
			return nil, fmt.Errorf("the key is disabled")
		}
		return privateKey, nil
	}

	// We used to have user name as Base 58 representation of a random 8 byte number,
	// try that.
	salt := base58.Decode(name)
	if len(salt) != 8 {
		return nil, fmt.Errorf("invalid user name or password")
	}
	privateKey = easyecc.NewPrivateKeyFromPassword([]byte(pass), salt)

	// Confirm that this key is registered.
	res, err = lookupClient.LookupKey(ctx, &pb.LookupKeyRequest{
		Key: privateKey.PublicKey().SerializeCompressed()})
	if err != nil {
		return nil, err
	}
	if res.Disabled {
		return nil, fmt.Errorf("the key is disabled")
	}
	return privateKey, nil
}
