package bc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/regnull/easyecc"
	"github.com/regnull/ubikom/util"
	"github.com/spf13/cobra"
)

func LoadKeyFromFlag(cmd *cobra.Command, keyFlagName string) (*easyecc.PrivateKey, error) {
	keyFile, err := cmd.Flags().GetString(keyFlagName)
	if err != nil {
		return nil, err
	}

	if keyFile == "" {
		keyFile, err = util.GetDefaultKeyLocation()
		if err != nil {
			return nil, err
		}
	}

	encrypted, err := util.IsKeyEncrypted(keyFile)
	if err != nil {
		return nil, err
	}

	passphrase := ""
	if encrypted {
		passphrase, err = util.ReadPassphase()
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

func WaitMinedAndPrintReceipt(client *ethclient.Client, tx *types.Transaction, waitDuration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), waitDuration)
	defer cancel()
	receipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		return err
	}
	jsonBytes, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", string(jsonBytes))
	return nil
}
