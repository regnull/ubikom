package util

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/regnull/easyecc"
	"golang.org/x/crypto/ripemd160"
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
