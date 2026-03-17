package cli

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"
	"taxsend/internal/archive"
	"taxsend/internal/crypto"
	"taxsend/internal/fsutil"
)

func newEncryptCmd() *cobra.Command {
	var recipient string
	var output string
	var force bool
	cmd := &cobra.Command{
		Use:   "encrypt --recipient age1... [files/dirs...]",
		Short: "Encrypt files/directories into one .tar.age artifact",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p := getPrinter(cmd.Context())
			r, err := crypto.ParseRecipient(recipient)
			if err != nil {
				return err
			}
			if output == "" {
				output = fmt.Sprintf("bundle-%s.tar.age", time.Now().Format("20060102-150405"))
			}
			files, err := archive.CollectInputs(args)
			if err != nil {
				return err
			}
			tmpFile, tmpPath, err := fsutil.CreateAtomic(output, force)
			if err != nil {
				return err
			}
			defer os.Remove(tmpPath)
			pr, pw := io.Pipe()
			go func() {
				pw.CloseWithError(archive.WriteTar(pw, files))
			}()
			if err := crypto.EncryptStream(tmpFile, pr, r); err != nil {
				_ = tmpFile.Close()
				return err
			}
			if err := tmpFile.Close(); err != nil {
				return err
			}
			if err := fsutil.CommitAtomic(tmpPath, output); err != nil {
				return err
			}
			p.Success("encryption complete")
			fmt.Printf("files: %d\noutput: %s\nrecipient: %s\n", len(files), output, recipient)
			p.Warn("artifact contents are encrypted, but output filename and size remain visible")
			return nil
		},
	}
	cmd.Flags().StringVar(&recipient, "recipient", "", "age recipient (age1...)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output .tar.age path")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing output file")
	_ = cmd.MarkFlagRequired("recipient")
	cmd.Example = `taxsend encrypt --recipient age1abc... --output 2025-tax-docs.tar.age T4.pdf RL1.pdf`
	return cmd
}
