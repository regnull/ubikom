package bc

import (
	"testing"

	"github.com/regnull/ubikom/globals"
	"github.com/stretchr/testify/assert"
)

func TestGetNodeURL(t *testing.T) {
	assert := assert.New(t)

	url, err := GetNodeURL("http://18.223.40.196:8545")
	assert.NoError(err)
	assert.Equal("http://18.223.40.196:8545", url)

	url, err = GetNodeURL("main")
	assert.NoError(err)
	assert.Equal(globals.InfuraNodeURL, url)

	url, err = GetNodeURL("sepolia")
	assert.NoError(err)
	assert.Equal(globals.InfuraSepoliaNodeURL, url)

	_, err = GetNodeURL("foo")
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
