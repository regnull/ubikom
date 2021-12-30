package bc

import (
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
