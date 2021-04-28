package cmd

import (
	"context"
	"log"
	"os"
	"path"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"teralyt.com/ubikom/ecc"
	"teralyt.com/ubikom/pb"
	"teralyt.com/ubikom/pow"
)

const (
	leadingZeros = 10
)

func init() {
	registerCmd.AddCommand(registerKeyCmd)
	rootCmd.AddCommand(registerCmd)
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register various things",
	Long:  "Register various things",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var registerKeyCmd = &cobra.Command{
	Use:   "key",
	Short: "Register public key",
	Long:  "Register public key",
	Run: func(cmd *cobra.Command, args []string) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		dir := path.Join(homeDir, defaultHomeSubDir)
		keyFile := path.Join(dir, defaultKeyFile)
		privateKey, err := ecc.LoadPrivateKey(keyFile)
		if err != nil {
			log.Fatal(err)
		}
		opts := []grpc.DialOption{
			grpc.WithInsecure(),
		}
		conn, err := grpc.Dial("localhost:8825", opts...)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		compressedKey := privateKey.PublicKey().SerializeCompressed()
		log.Printf("generating POW...")
		pow := pow.Compute(compressedKey, leadingZeros)
		log.Printf("POW found: %x", pow)

		client := pb.NewIdentityServiceClient(conn)
		res, err := client.RegisterKey(context.TODO(),
			&pb.KeyRegistrationRequest{
				Key: compressedKey,
				Pow: pow,
			})
		if err != nil {
			log.Fatal(err)
		}
		if res.Result != pb.ResultCode_OK {
			log.Fatalf("got response code: %d", res.Result)
		}
		log.Printf("key registered successfully")
	},
}
