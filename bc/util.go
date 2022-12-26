package bc

import (
	"fmt"
	"strings"

	"github.com/regnull/ubikom/globals"
)

func GetNodeURL(network string) (string, error) {
	if strings.HasPrefix(network, "http://") {
		return network, nil
	}
	switch network {
	case "main":
		return globals.InfuraNodeURL, nil
	case "sepolia":
		return globals.InfuraSepoliaNodeURL, nil
	}
	return "", fmt.Errorf("invalid network")
}

func GetContractAddress(network string, contractAddress string) (string, error) {
	if contractAddress != "" {
		return contractAddress, nil
	}
	switch network {
	case "main":
		return globals.MainnetNameRegistryAddress, nil
	case "sepolia":
		return globals.SepoliaNameRegistryAddress, nil
	}
	return "", fmt.Errorf("invalid network")
}
