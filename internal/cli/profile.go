package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"taxsend/internal/profile"
)

func newProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage sender profiles used by the high-level workflow",
	}
	cmd.AddCommand(newProfileImportCmd(), newProfileListCmd(), newProfileRemoveCmd())
	return cmd
}

func newProfileImportCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "import public-profile.json",
		Short: "Import a public sender profile onto the sending machine",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p := getPrinter(cmd.Context())
			senderProfile, savedPath, err := profile.Import(args[0], force)
			if err != nil {
				return err
			}
			p.Success("profile imported")
			fmt.Printf("name: %s\nsaved: %s\nrecipients: %d\n", senderProfile.Name, savedPath, len(senderProfile.Recipients))
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "overwrite an existing sender profile")
	cmd.Example = `taxsend profile import personal-laptop.public.json`
	return cmd
}

func newProfileListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List locally stored sender profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			names, err := profile.ListSenderNames()
			if err != nil {
				return err
			}
			for _, name := range names {
				fmt.Println(name)
			}
			return nil
		},
	}
	cmd.Example = `taxsend profile list`
	return cmd
}

func newProfileRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a locally stored sender profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p := getPrinter(cmd.Context())
			path, err := profile.RemoveSender(args[0])
			if err != nil {
				return err
			}
			p.Success("profile removed")
			fmt.Printf("removed: %s\n", path)
			return nil
		},
	}
	cmd.Example = `taxsend profile remove personal-laptop`
	return cmd
}
