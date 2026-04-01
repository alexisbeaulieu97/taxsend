package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"taxsend/internal/archive"
	"taxsend/internal/chunk"
	"taxsend/internal/crypto"
	"taxsend/internal/fsutil"
	"taxsend/internal/profile"
	"taxsend/internal/sizeutil"
)

func newSealCmd() *cobra.Command {
	var profileName string
	var outputDir string
	var basename string
	var maxPartSize string
	var force bool
	cmd := &cobra.Command{
		Use:   "seal --to <profile> [files/dirs...]",
		Short: "Encrypt to a saved sender profile and optionally split into email-safe parts",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p := getPrinter(cmd.Context())
			senderProfile, err := profile.LoadSender(profileName)
			if err != nil {
				return err
			}
			partSize := senderProfile.DefaultMaxPartSizeBytes
			if maxPartSize != "" {
				partSize, err = sizeutil.ParseBytes(maxPartSize)
				if err != nil {
					return err
				}
			}
			if basename == "" {
				basename = defaultAttachmentBase(time.Now())
			}
			basename, err = normalizeBasename(basename)
			if err != nil {
				return err
			}
			files, err := archive.CollectInputs(args)
			if err != nil {
				return err
			}
			recipients, err := crypto.ParseRecipients(senderProfile.Recipients, "")
			if err != nil {
				return err
			}
			tmpTarget := filepath.Join(outputDir, basename+senderProfile.DefaultOutputExtension)
			tmpFile, tmpPath, err := fsutil.CreateAtomic(tmpTarget, force)
			if err != nil {
				return err
			}
			defer os.Remove(tmpPath)
			pr, pw := io.Pipe()
			go func() {
				pw.CloseWithError(archive.WriteTar(pw, files))
			}()
			if err := crypto.EncryptStream(tmpFile, pr, recipients...); err != nil {
				_ = tmpFile.Close()
				return err
			}
			if err := tmpFile.Close(); err != nil {
				return err
			}
			outputs, err := chunk.FinalizeEncryptedArtifact(tmpPath, outputDir, basename, senderProfile.DefaultOutputExtension, partSize, force)
			if err != nil {
				return err
			}
			p.Success("seal complete")
			fmt.Printf("profile: %s\nfiles: %d\nparts: %d\n", senderProfile.Name, len(files), len(outputs))
			for _, output := range outputs {
				fmt.Printf("output: %s\n", output)
			}
			p.Warn("email metadata, attachment names, and message timing remain visible to the mail system")
			return nil
		},
	}
	cmd.Flags().StringVar(&profileName, "to", "", "sender profile name")
	cmd.Flags().StringVar(&outputDir, "output-dir", ".", "directory for encrypted output")
	cmd.Flags().StringVar(&basename, "basename", "", "neutral output base name without a path")
	cmd.Flags().StringVar(&maxPartSize, "max-part-size", "", "maximum size per output file, for example 7MiB or 7000000")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing output files")
	_ = cmd.MarkFlagRequired("to")
	cmd.Example = `taxsend seal --to personal-laptop --output-dir . T4.pdf RL1.pdf`
	return cmd
}
