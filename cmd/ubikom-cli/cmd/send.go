package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/regnull/ubikom/bc"
	"github.com/regnull/ubikom/cmd/ubikom-cli/cmd/cmdutil"
	"github.com/regnull/ubikom/protoutil"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	sendCmd.PersistentFlags().String("network", "main", "mode, either live or prod")
	sendCmd.PersistentFlags().String("node-url", "", "blockchain node location")
	sendCmd.PersistentFlags().String("contract-address", "", "registry contract address")

	sendMessageCmd.Flags().String("receiver", "", "receiver's address")
	sendMessageCmd.Flags().String("sender", "", "sender's address")
	sendMessageCmd.Flags().String("key", "", "Location for the private key file")
	sendCmd.AddCommand(sendMessageCmd)

	rootCmd.AddCommand(sendCmd)
}

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send stuff",
	Long:  "Send stuff",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var sendMessageCmd = &cobra.Command{
	Use:   "message",
	Short: "Send message",
	Long:  "Send message",
	Run: func(cmd *cobra.Command, args []string) {
		nodeURL, err := cmdutil.GetNodeURL(cmd.Flags())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get node URL")
		}
		log.Debug().Str("node-url", nodeURL).Msg("using node")
		contractAddress, err := cmdutil.GetContractAddress(cmd.Flags())
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load contract address")
		}
		log.Debug().Str("contract-address", contractAddress).Msg("using contract addresss")

		privateKey, err := cmdutil.LoadKeyFromFlag(cmd, "key")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to load encryption key")
		}

		sender, err := cmd.Flags().GetString("sender")
		if err != nil || sender == "" {
			log.Fatal().Err(err).Msg("sender's address must be specified")
		}

		receiver, err := cmd.Flags().GetString("receiver")
		if err != nil || receiver == "" {
			log.Fatal().Err(err).Msg("receiver's address must be specified")
		}

		ctx := context.Background()

		bchain, err := bc.NewBlockchain(nodeURL, contractAddress)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create lookup service")
		}

		var lines []string
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter your message, dot on an empty line to finish: \n")
		for {
			text, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			if text == ".\n" {
				break
			}
			lines = append(lines, strings.ReplaceAll(text, "\n", ""))
		}
		body := strings.Join(lines, "\n")

		err = protoutil.SendMessage(ctx, privateKey, []byte(body), sender, receiver, bchain)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to send message")
		}
	},
}
