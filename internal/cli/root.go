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
		Short: "Securely bundle and encrypt tax documents with age",
		Long:  "TaxSend is a local-first CLI that creates encrypted .tar.age artifacts from files and directories.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmd.SetContext(withPrinter(cmd.Context(), output.New(verbose)))
		},
	}
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
	cmd.AddCommand(newKeygenCmd(), newRecipientCmd(), newEncryptCmd(), newDecryptCmd(), newInspectCmd(), newVersionCmd())
	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{Use: "version", Short: "Show version", Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Version)
	}}
}
