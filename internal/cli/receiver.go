package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"taxsend/internal/crypto"
	"taxsend/internal/keystore"
	"taxsend/internal/profile"
)

func newReceiverCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "receiver",
		Short: "Manage encrypted local receiver identities",
	}
	cmd.AddCommand(newReceiverInitCmd(), newReceiverMigrateCmd(), newReceiverRecipientCmd())
	return cmd
}

func newReceiverInitCmd() *cobra.Command {
	var name string
	var profileOut string
	var force bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create an encrypted local receiver identity and export a public profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			p := getPrinter(cmd.Context())
			passphrase, err := promptPassphrase(true)
			if err != nil {
				return err
			}
			identity, err := crypto.GenerateIdentity()
			if err != nil {
				return err
			}
			senderProfile, err := profile.Default(name, []string{identity.Recipient().String()})
			if err != nil {
				return err
			}
			if profileOut == "" {
				profileOut = defaultPublicProfilePath(senderProfile.Name)
			}
			keystorePath, err := keystore.SaveIdentity(senderProfile.Name, identity, passphrase, force)
			if err != nil {
				return err
			}
			metadataPath, err := profile.SaveReceiverMetadata(senderProfile, force)
			if err != nil {
				return err
			}
			if err := profile.SaveFile(profileOut, senderProfile, force); err != nil {
				return err
			}
			p.Success("receiver initialized")
			fmt.Printf("name: %s\nkeystore: %s\nmetadata: %s\nprofile-out: %s\nrecipient: %s\n", senderProfile.Name, keystorePath, metadataPath, profileOut, senderProfile.Recipients[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "receiver profile name")
	cmd.Flags().StringVar(&profileOut, "profile-out", "", "path to write the exported public profile JSON")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing receiver files")
	_ = cmd.MarkFlagRequired("name")
	cmd.Example = `taxsend receiver init --name personal-laptop --profile-out personal-laptop.public.json`
	return cmd
}

func newReceiverMigrateCmd() *cobra.Command {
	var name string
	var identityPath string
	var profileOut string
	var force bool
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Import a legacy plaintext identity into the encrypted receiver keystore",
		RunE: func(cmd *cobra.Command, args []string) error {
			p := getPrinter(cmd.Context())
			passphrase, err := promptPassphrase(true)
			if err != nil {
				return err
			}
			identity, keystorePath, err := keystore.SaveMigratedIdentity(name, identityPath, passphrase, force)
			if err != nil {
				return err
			}
			senderProfile, err := profile.Default(name, []string{identity.Recipient().String()})
			if err != nil {
				return err
			}
			if profileOut == "" {
				profileOut = defaultPublicProfilePath(senderProfile.Name)
			}
			metadataPath, err := profile.SaveReceiverMetadata(senderProfile, force)
			if err != nil {
				return err
			}
			if err := profile.SaveFile(profileOut, senderProfile, force); err != nil {
				return err
			}
			p.Success("receiver migrated")
			fmt.Printf("name: %s\nkeystore: %s\nmetadata: %s\nprofile-out: %s\nrecipient: %s\n", senderProfile.Name, keystorePath, metadataPath, profileOut, senderProfile.Recipients[0])
			p.Warn("legacy identity file was not deleted; remove it manually after verifying the new workflow")
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "receiver profile name")
	cmd.Flags().StringVar(&identityPath, "identity", "", "path to the legacy plaintext age identity file")
	cmd.Flags().StringVar(&profileOut, "profile-out", "", "path to write the exported public profile JSON")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing receiver files")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("identity")
	cmd.Example = `taxsend receiver migrate --name personal-laptop --identity ~/.config/taxsend/identity.txt`
	return cmd
}

func newReceiverRecipientCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "recipient",
		Short: "Display recipient strings from stored receiver metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			metadata, err := profile.LoadReceiverMetadata(name)
			if err != nil {
				return err
			}
			for _, recipient := range metadata.Recipients {
				fmt.Println(recipient)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "receiver profile name")
	_ = cmd.MarkFlagRequired("name")
	cmd.Example = `taxsend receiver recipient --name personal-laptop`
	return cmd
}
