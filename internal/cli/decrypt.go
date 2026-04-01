package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"taxsend/internal/archive"
	"taxsend/internal/crypto"
)

func newDecryptCmd() *cobra.Command {
	var identity string
	var outDir string
	var force bool
	cmd := &cobra.Command{
		Use:   "decrypt --identity ~/.config/taxsend/identity.txt artifact.tar.age",
		Short: "Low-level compatibility: decrypt a .tar.age artifact using a plaintext identity file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p := getPrinter(cmd.Context())
			id, err := crypto.LoadIdentity(identity)
			if err != nil {
				return err
			}
			f, err := os.Open(args[0])
			if err != nil {
				return err
			}
			defer f.Close()
			r, err := crypto.DecryptStream(f, id)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(outDir, 0o700); err != nil {
				return err
			}
			count, err := archive.ExtractTar(r, outDir, force)
			if err != nil {
				return err
			}
			p.Success("decryption complete")
			fmt.Printf("files: %d\noutput-dir: %s\nidentity: %s\n", count, outDir, identity)
			return nil
		},
	}
	cmd.Flags().StringVar(&identity, "identity", "", "path to age identity file")
	cmd.Flags().StringVar(&outDir, "output-dir", ".", "directory to extract into")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing files in output directory")
	_ = cmd.MarkFlagRequired("identity")
	cmd.Example = `taxsend decrypt --identity ~/.config/taxsend/identity.txt --output-dir ./out 2025-tax-docs.tar.age`
	return cmd
}
