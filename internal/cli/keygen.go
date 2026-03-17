package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"taxsend/internal/crypto"
	"taxsend/internal/fsutil"
)

func newKeygenCmd() *cobra.Command {
	var output string
	var force bool
	cmd := &cobra.Command{
		Use:   "keygen",
		Short: "Generate a new age identity and recipient",
		RunE: func(cmd *cobra.Command, args []string) error {
			p := getPrinter(cmd.Context())
			if err := fsutil.EnsureNotExists(output, force); err != nil {
				return err
			}
			id, err := crypto.GenerateIdentity()
			if err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(output, []byte(id.String()+"\n"), 0o600); err != nil {
				return err
			}
			p.Success("identity generated")
			fmt.Printf("identity: %s\n", output)
			fmt.Printf("recipient: %s\n", id.Recipient())
			p.Warn("store your identity file securely; it is required for decryption")
			return nil
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", filepath.Join(os.Getenv("HOME"), ".config", "taxsend", "identity.txt"), "identity file path")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing identity file")
	return cmd
}
