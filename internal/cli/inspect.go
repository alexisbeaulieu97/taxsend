package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"taxsend/internal/chunk"
)

func newInspectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect artifact.tar.age",
		Short: "Inspect minimal artifact metadata without decryption",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			artifactInfo, err := chunk.Collect(args[0])
			if err != nil {
				return err
			}
			inspectPath := artifactInfo.PartPaths[0]
			f, err := os.Open(inspectPath)
			if err != nil {
				return err
			}
			defer f.Close()
			r := bufio.NewReader(f)
			line, err := r.ReadString('\n')
			if err != nil {
				return fmt.Errorf("unable to inspect artifact: %w", err)
			}
			fmt.Printf("file: %s\n", args[0])
			if artifactInfo.Chunked {
				fmt.Printf("chunked: yes\nparts: %d\n", len(artifactInfo.PartPaths))
				fmt.Printf("base-name: %s%s\n", artifactInfo.BaseName, artifactInfo.Ext)
			} else {
				fmt.Println("chunked: no")
			}
			if strings.HasPrefix(line, "age-encryption.org/") {
				fmt.Println("type: age-encrypted payload")
				if artifactInfo.Chunked {
					fmt.Println("format: encrypted tar stream split across multiple parts")
				} else {
					fmt.Println("format: encrypted tar stream")
				}
			} else {
				fmt.Println("type: unknown (not recognized as age) ")
			}
			fmt.Println("note: plaintext file metadata is not inspected or exposed")
			return nil
		},
	}
	cmd.Example = `taxsend inspect attachment-20260401-120000.part001.bin`
	return cmd
}
