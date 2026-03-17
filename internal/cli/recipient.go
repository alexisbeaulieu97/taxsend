package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"taxsend/internal/crypto"
)

func newRecipientCmd() *cobra.Command {
	var identity string
	cmd := &cobra.Command{
		Use:   "recipient",
		Short: "Display the recipient for an age identity",
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := crypto.LoadIdentity(identity)
			if err != nil {
				return err
			}
			fmt.Println(id.Recipient().String())
			return nil
		},
	}
	cmd.Flags().StringVar(&identity, "identity", "", "path to age identity file")
	_ = cmd.MarkFlagRequired("identity")
	return cmd
}
