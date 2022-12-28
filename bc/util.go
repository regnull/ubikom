package bc

import (
	"fmt"
	"os"
	"strings"

	"github.com/regnull/ubikom/globals"
)

func GetNodeURL(network string, projectId string) (string, error) {
	if strings.HasPrefix(network, "http://") {
		return network, nil
	}
	if projectId == "" {
		projectId = os.Getenv("INFURA_PROJECT_ID")
	}
	if projectId == "" {
		return "", fmt.Errorf("invalid project id")
	}
	switch network {
	case "main":
		return fmt.Sprintf(globals.InfuraNodeURL, projectId), nil
	case "sepolia":
		return fmt.Sprintf(globals.InfuraSepoliaNodeURL, projectId), nil
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
