package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"taxsend/internal/archive"
	"taxsend/internal/chunk"
	"taxsend/internal/crypto"
	"taxsend/internal/keystore"
)

func newUnsealCmd() *cobra.Command {
	var name string
	var outputDir string
	var force bool
	cmd := &cobra.Command{
		Use:   "unseal <artifact-or-first-part>",
		Short: "Reassemble if needed, decrypt with the local receiver keystore, and extract files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p := getPrinter(cmd.Context())
			receiverName, err := resolveReceiverName(name)
			if err != nil {
				return err
			}
			passphrase, err := promptPassphrase(false)
			if err != nil {
				return err
			}
			identity, err := keystore.LoadIdentity(receiverName, passphrase)
			if err != nil {
				return err
			}
			input, artifactInfo, err := chunk.OpenJoined(args[0])
			if err != nil {
				return err
			}
			defer input.Close()
			reader, err := crypto.DecryptStream(input, identity)
			if err != nil {
				return err
			}
			if outputDir == "" {
				outputDir = defaultUnsealDir(time.Now())
			}
			if err := os.MkdirAll(outputDir, 0o700); err != nil {
				return err
			}
			count, err := archive.ExtractTar(reader, outputDir, force)
			if err != nil {
				return err
			}
			p.Success("unseal complete")
			fmt.Printf("receiver: %s\nfiles: %d\noutput-dir: %s\nparts: %d\n", receiverName, count, outputDir, len(artifactInfo.PartPaths))
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "receiver profile name; required if multiple local receivers exist")
	cmd.Flags().StringVar(&outputDir, "output-dir", "", "directory to extract into; defaults to a private timestamped directory")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing extracted files")
	cmd.Example = `taxsend unseal attachment-20260401-120000.part001.bin`
	return cmd
}
