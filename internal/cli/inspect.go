package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newInspectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "inspect artifact.tar.age",
		Short: "Inspect minimal artifact metadata without decryption",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := os.Open(args[0])
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
			if strings.HasPrefix(line, "age-encryption.org/") {
				fmt.Println("type: age-encrypted payload")
				fmt.Println("format: encrypted tar stream (.tar.age)")
			} else {
				fmt.Println("type: unknown (not recognized as age) ")
			}
			fmt.Println("note: plaintext file metadata is not inspected or exposed")
			return nil
		},
	}
}
