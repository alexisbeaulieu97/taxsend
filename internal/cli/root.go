package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"taxsend/internal/output"
	"taxsend/internal/version"
)

func NewRootCmd() *cobra.Command {
	var verbose bool
	cmd := &cobra.Command{
		Use:   "taxsend",
		Short: "Encrypt sensitive files for store-now, decrypt-later delivery over email",
		Long:  "TaxSend is a local-first CLI that encrypts file bundles with age and supports a higher-level sender/receiver workflow for email-safe delivery.",
		Example: `  # On your personal computer
  taxsend receiver init --name personal-laptop --profile-out personal-laptop.public.json

  # On your work computer
  taxsend profile import personal-laptop.public.json
  taxsend seal --to personal-laptop T4.pdf RL1.pdf

  # Back on your personal computer
  taxsend unseal attachment-20260401-120000.part001.bin

  # Low-level compatibility commands still exist
  taxsend encrypt --recipient age1abc... --output 2025-tax-docs.tar.age T4.pdf RL1.pdf
  taxsend decrypt --identity ~/.config/taxsend/identity.txt --output-dir ./out 2025-tax-docs.tar.age`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmd.SetContext(withPrinter(cmd.Context(), output.New(verbose)))
		},
	}
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
	cmd.AddCommand(
		newReceiverCmd(),
		newProfileCmd(),
		newSealCmd(),
		newUnsealCmd(),
		newKeygenCmd(),
		newRecipientCmd(),
		newEncryptCmd(),
		newDecryptCmd(),
		newInspectCmd(),
		newVersionCmd(),
	)
	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{Use: "version", Short: "Show version", Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Version)
	}}
}
