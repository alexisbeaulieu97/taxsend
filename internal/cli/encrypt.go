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
	var recipients []string
	var recipientsFile string
	var output string
	var force bool
	cmd := &cobra.Command{
		Use:   "encrypt --recipient age1... [files/dirs...]",
		Short: "Low-level compatibility: encrypt files/directories into one .tar.age artifact",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p := getPrinter(cmd.Context())
			parsedRecipients, err := crypto.ParseRecipients(recipients, recipientsFile)
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
			if err := crypto.EncryptStream(tmpFile, pr, parsedRecipients...); err != nil {
				_ = tmpFile.Close()
				return err
			}
			if err := tmpFile.Close(); err != nil {
				return err
			}
			if err := fsutil.CommitAtomicForce(tmpPath, output, force); err != nil {
				return err
			}
			p.Success("encryption complete")
			fmt.Printf("files: %d\noutput: %s\nrecipients: %d\n", len(files), output, len(parsedRecipients))
			p.Warn("artifact contents are encrypted, but output filename and size remain visible")
			return nil
		},
	}
	cmd.Flags().StringArrayVar(&recipients, "recipient", nil, "age recipient (repeatable)")
	cmd.Flags().StringVar(&recipientsFile, "recipients-file", "", "path to a file containing one native age recipient per line")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output .tar.age path")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing output file")
	cmd.Example = `taxsend encrypt --recipient age1abc... --recipient age1def... --output 2025-tax-docs.tar.age T4.pdf RL1.pdf`
	return cmd
}
