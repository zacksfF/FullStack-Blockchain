package cmd

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
	"github.com/zacksfF/FullStack-Blockchain/blockchain/database"
)

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Print account for the specific wallet",
	Run:   accountRun,
}

func init() {
	rootCmd.AddCommand(accountCmd)
}

func accountRun(cmd *cobra.Command, args []string) {
	privateKey, err := crypto.LoadECDSA(getPrivateKeyPath())
	if err != nil {
		log.Fatal(err)
	}

	accountID := database.PublicKeyToAccountID(privateKey.PublicKey)
	fmt.Println(accountID)
}
