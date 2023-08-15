package bc

import (
	"fmt"
	"os"
	"testing"

	"github.com/regnull/ubikom/globals"
	"github.com/stretchr/testify/assert"
)

func TestGetNodeURL(t *testing.T) {
	assert := assert.New(t)

	url, err := GetNodeURL("http://18.223.40.196:8545", "")
	assert.NoError(err)
	assert.Equal("http://18.223.40.196:8545", url)

	url, err = GetNodeURL("main", "123456")
	assert.NoError(err)
	assert.Equal(fmt.Sprintf(globals.InfuraNodeURL, "123456"), url)

	url, err = GetNodeURL("sepolia", "123456")
	assert.NoError(err)
	assert.Equal(fmt.Sprintf(globals.InfuraSepoliaNodeURL, "123456"), url)

	_, err = GetNodeURL("foo", "123456")
	assert.Error(err)

	// Test using environment variable for the project ID.
	err = os.Setenv("INFURA_PROJECT_ID", "123456")
	assert.NoError(err)

	url, err = GetNodeURL("main", "")
	assert.Equal(fmt.Sprintf(globals.InfuraNodeURL, "123456"), url)

	// If the environment variable is not set, we should get an error.
	err = os.Unsetenv("INFURA_PROJECT_ID")
	assert.NoError(err)
	_, err = GetNodeURL("main", "")
	assert.Error(err)
}

func TestGetContract(t *testing.T) {
	assert := assert.New(t)

	contract, err := GetContractAddress("whatever", "0xcc8650c9cd8d99b62375c22f270a803e7abf0de9")
	assert.NoError(err)
	assert.Equal("0xcc8650c9cd8d99b62375c22f270a803e7abf0de9", contract)

	contract, err = GetContractAddress("main", "")
	assert.NoError(err)
	assert.Equal(globals.MainnetNameRegistryAddress, contract)

	contract, err = GetContractAddress("sepolia", "")
	assert.NoError(err)
	assert.Equal(globals.SepoliaNameRegistryAddress, contract)

	_, err = GetContractAddress("foo", "")
	assert.Error(err)
}
