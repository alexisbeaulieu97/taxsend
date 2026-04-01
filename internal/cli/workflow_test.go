package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"taxsend/internal/crypto"
	"taxsend/internal/profile"
	"taxsend/internal/storage"
)

func TestReceiverInitCreatesEncryptedKeystoreAndProfile(t *testing.T) {
	tmp := t.TempDir()
	switchConfig := configSwitcher(t)
	switchConfig(filepath.Join(tmp, "personal-config"))

	profileOut := filepath.Join(tmp, "personal-laptop.public.json")
	restorePrompt := stubPrompts(t, "correct horse battery staple", "correct horse battery staple")
	defer restorePrompt()

	if _, err := runCapture([]string{"receiver", "init", "--name", "personal-laptop", "--profile-out", profileOut}); err != nil {
		t.Fatal(err)
	}

	keystorePath, err := storage.ReceiverKeystorePath("personal-laptop")
	if err != nil {
		t.Fatal(err)
	}
	metadataPath, err := storage.ReceiverMetadataPath("personal-laptop")
	if err != nil {
		t.Fatal(err)
	}
	keystoreBytes, err := os.ReadFile(keystorePath)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(keystoreBytes, []byte("AGE-SECRET-KEY-")) {
		t.Fatal("keystore unexpectedly contains a plaintext identity")
	}

	receiverProfile, err := profile.LoadReceiverMetadata("personal-laptop")
	if err != nil {
		t.Fatal(err)
	}
	publicProfile, err := profile.LoadFile(profileOut)
	if err != nil {
		t.Fatal(err)
	}
	if receiverProfile.Name != "personal-laptop" {
		t.Fatalf("got name %q", receiverProfile.Name)
	}
	if receiverProfile.Recipients[0] != publicProfile.Recipients[0] {
		t.Fatalf("recipient mismatch: %q != %q", receiverProfile.Recipients[0], publicProfile.Recipients[0])
	}
	if _, err := os.Stat(metadataPath); err != nil {
		t.Fatal(err)
	}
}

func TestReceiverMigratePreservesRecipient(t *testing.T) {
	tmp := t.TempDir()
	legacyIdentity := filepath.Join(tmp, "identity.txt")
	if _, err := runCapture([]string{"keygen", "--output", legacyIdentity}); err != nil {
		t.Fatal(err)
	}
	legacy, err := crypto.LoadIdentity(legacyIdentity)
	if err != nil {
		t.Fatal(err)
	}

	switchConfig := configSwitcher(t)
	switchConfig(filepath.Join(tmp, "personal-config"))
	profileOut := filepath.Join(tmp, "migrated.public.json")
	restorePrompt := stubPrompts(t, "new secret", "new secret")
	defer restorePrompt()

	if _, err := runCapture([]string{"receiver", "migrate", "--name", "personal-laptop", "--identity", legacyIdentity, "--profile-out", profileOut}); err != nil {
		t.Fatal(err)
	}

	receiverProfile, err := profile.LoadReceiverMetadata("personal-laptop")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := receiverProfile.Recipients[0], legacy.Recipient().String(); got != want {
		t.Fatalf("got recipient %q, want %q", got, want)
	}
}

func TestProfileImportListAndRemove(t *testing.T) {
	tmp := t.TempDir()
	identity, err := crypto.GenerateIdentity()
	if err != nil {
		t.Fatal(err)
	}
	publicProfile, err := profile.Default("work-target", []string{identity.Recipient().String()})
	if err != nil {
		t.Fatal(err)
	}
	publicPath := filepath.Join(tmp, "work-target.public.json")
	if err := profile.SaveFile(publicPath, publicProfile, true); err != nil {
		t.Fatal(err)
	}

	switchConfig := configSwitcher(t)
	switchConfig(filepath.Join(tmp, "work-config"))

	if _, err := runCapture([]string{"profile", "import", publicPath}); err != nil {
		t.Fatal(err)
	}
	out, err := runCapture([]string{"profile", "list"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "work-target") {
		t.Fatalf("expected imported profile in output: %s", out)
	}
	if _, err := runCapture([]string{"profile", "remove", "work-target"}); err != nil {
		t.Fatal(err)
	}
	out, err = runCapture([]string{"profile", "list"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "work-target") {
		t.Fatalf("unexpected removed profile in output: %s", out)
	}
}

func TestSealUnsealChunkedRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	personalConfig := filepath.Join(tmp, "personal-config")
	workConfig := filepath.Join(tmp, "work-config")
	switchConfig := configSwitcher(t)

	profileOut := filepath.Join(tmp, "personal.public.json")
	switchConfig(personalConfig)
	restorePrompt := stubPrompts(t, "swordfish", "swordfish")
	if _, err := runCapture([]string{"receiver", "init", "--name", "personal", "--profile-out", profileOut}); err != nil {
		restorePrompt()
		t.Fatal(err)
	}
	restorePrompt()

	docsDir := filepath.Join(tmp, "docs")
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docsDir, "T4.txt"), bytes.Repeat([]byte("income-"), 128), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docsDir, "RL1.txt"), bytes.Repeat([]byte("tax-"), 128), 0o644); err != nil {
		t.Fatal(err)
	}

	switchConfig(workConfig)
	if _, err := runCapture([]string{"profile", "import", profileOut}); err != nil {
		t.Fatal(err)
	}
	sealedDir := filepath.Join(tmp, "sealed")
	if _, err := runCapture([]string{"seal", "--to", "personal", "--output-dir", sealedDir, "--max-part-size", "512", docsDir}); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(sealedDir)
	if err != nil {
		t.Fatal(err)
	}
	var partPaths []string
	for _, entry := range entries {
		partPaths = append(partPaths, filepath.Join(sealedDir, entry.Name()))
	}
	if len(partPaths) < 2 {
		t.Fatalf("expected chunked output, got %d files", len(partPaths))
	}

	switchConfig(personalConfig)
	restorePrompt = stubPrompts(t, "swordfish")
	defer restorePrompt()
	outDir := filepath.Join(tmp, "unsealed")
	if _, err := runCapture([]string{"unseal", "--name", "personal", "--output-dir", outDir, partPaths[1]}); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(outDir, filepath.Base(docsDir), "T4.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, bytes.Repeat([]byte("income-"), 128)) {
		t.Fatal("unexpected decrypted content")
	}
}

func TestUnsealMissingPartFails(t *testing.T) {
	tmp := t.TempDir()
	personalConfig := filepath.Join(tmp, "personal-config")
	workConfig := filepath.Join(tmp, "work-config")
	switchConfig := configSwitcher(t)
	profileOut := filepath.Join(tmp, "personal.public.json")

	switchConfig(personalConfig)
	restorePrompt := stubPrompts(t, "swordfish", "swordfish")
	if _, err := runCapture([]string{"receiver", "init", "--name", "personal", "--profile-out", profileOut}); err != nil {
		restorePrompt()
		t.Fatal(err)
	}
	restorePrompt()

	docsDir := filepath.Join(tmp, "docs")
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docsDir, "T4.txt"), bytes.Repeat([]byte("income-"), 128), 0o644); err != nil {
		t.Fatal(err)
	}

	switchConfig(workConfig)
	if _, err := runCapture([]string{"profile", "import", profileOut}); err != nil {
		t.Fatal(err)
	}
	sealedDir := filepath.Join(tmp, "sealed")
	if _, err := runCapture([]string{"seal", "--to", "personal", "--output-dir", sealedDir, "--max-part-size", "256", docsDir}); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(sealedDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) < 2 {
		t.Fatalf("expected chunked output, got %d", len(entries))
	}
	if err := os.Remove(filepath.Join(sealedDir, entries[1].Name())); err != nil {
		t.Fatal(err)
	}

	switchConfig(personalConfig)
	restorePrompt = stubPrompts(t, "swordfish")
	defer restorePrompt()
	_, err = runCapture([]string{"unseal", "--name", "personal", filepath.Join(sealedDir, entries[0].Name())})
	if err == nil || !strings.Contains(err.Error(), "missing artifact part") {
		t.Fatalf("got err %v", err)
	}
}

func TestUnsealWrongPassphraseFails(t *testing.T) {
	tmp := t.TempDir()
	switchConfig := configSwitcher(t)
	personalConfig := filepath.Join(tmp, "personal-config")
	workConfig := filepath.Join(tmp, "work-config")
	profileOut := filepath.Join(tmp, "personal.public.json")

	switchConfig(personalConfig)
	restorePrompt := stubPrompts(t, "correct", "correct")
	if _, err := runCapture([]string{"receiver", "init", "--name", "personal", "--profile-out", profileOut}); err != nil {
		restorePrompt()
		t.Fatal(err)
	}
	restorePrompt()

	docsDir := filepath.Join(tmp, "docs")
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docsDir, "T4.txt"), []byte("income"), 0o644); err != nil {
		t.Fatal(err)
	}

	switchConfig(workConfig)
	if _, err := runCapture([]string{"profile", "import", profileOut}); err != nil {
		t.Fatal(err)
	}
	sealedDir := filepath.Join(tmp, "sealed")
	if _, err := runCapture([]string{"seal", "--to", "personal", "--output-dir", sealedDir, "--max-part-size", "1000000", docsDir}); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(sealedDir)
	if err != nil {
		t.Fatal(err)
	}

	switchConfig(personalConfig)
	restorePrompt = stubPrompts(t, "wrong")
	defer restorePrompt()
	outDir := filepath.Join(tmp, "wrong-passphrase-out")
	_, err = runCapture([]string{"unseal", "--name", "personal", "--output-dir", outDir, filepath.Join(sealedDir, entries[0].Name())})
	if err == nil {
		t.Fatal("expected passphrase failure")
	}
	if _, statErr := os.Stat(outDir); !os.IsNotExist(statErr) {
		t.Fatalf("unexpected output dir state: %v", statErr)
	}
}

func TestLowLevelEncryptSupportsMultipleRecipientsAndRecipientsFile(t *testing.T) {
	tmp := t.TempDir()
	docsDir := filepath.Join(tmp, "docs")
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docsDir, "T4.txt"), []byte("income"), 0o644); err != nil {
		t.Fatal(err)
	}
	idOne := filepath.Join(tmp, "one.txt")
	idTwo := filepath.Join(tmp, "two.txt")
	if _, err := runCapture([]string{"keygen", "--output", idOne}); err != nil {
		t.Fatal(err)
	}
	if _, err := runCapture([]string{"keygen", "--output", idTwo}); err != nil {
		t.Fatal(err)
	}
	identityOne, err := crypto.LoadIdentity(idOne)
	if err != nil {
		t.Fatal(err)
	}
	identityTwo, err := crypto.LoadIdentity(idTwo)
	if err != nil {
		t.Fatal(err)
	}
	recipientsFile := filepath.Join(tmp, "recipients.txt")
	if err := os.WriteFile(recipientsFile, []byte(identityTwo.Recipient().String()+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	encrypted := filepath.Join(tmp, "bundle.tar.age")
	if _, err := runCapture([]string{"encrypt", "--recipient", identityOne.Recipient().String(), "--recipients-file", recipientsFile, "--output", encrypted, docsDir}); err != nil {
		t.Fatal(err)
	}
	outOne := filepath.Join(tmp, "out-one")
	outTwo := filepath.Join(tmp, "out-two")
	if _, err := runCapture([]string{"decrypt", "--identity", idOne, "--output-dir", outOne, encrypted}); err != nil {
		t.Fatal(err)
	}
	if _, err := runCapture([]string{"decrypt", "--identity", idTwo, "--output-dir", outTwo, encrypted}); err != nil {
		t.Fatal(err)
	}
	for _, outDir := range []string{outOne, outTwo} {
		got, err := os.ReadFile(filepath.Join(outDir, filepath.Base(docsDir), "T4.txt"))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != "income" {
			t.Fatalf("unexpected content %q", got)
		}
	}
}

func stubPrompts(t *testing.T, prompts ...string) func() {
	t.Helper()
	old := promptSecret
	index := 0
	promptSecret = func(prompt string) (string, error) {
		if index >= len(prompts) {
			return "", fmt.Errorf("unexpected prompt: %s", prompt)
		}
		value := prompts[index]
		index++
		return value, nil
	}
	return func() { promptSecret = old }
}

func configSwitcher(t *testing.T) func(string) {
	t.Helper()
	prev, hadPrev := os.LookupEnv(storage.ConfigDirEnv)
	t.Cleanup(func() {
		if hadPrev {
			_ = os.Setenv(storage.ConfigDirEnv, prev)
		} else {
			_ = os.Unsetenv(storage.ConfigDirEnv)
		}
	})
	return func(dir string) {
		t.Helper()
		if err := os.Setenv(storage.ConfigDirEnv, dir); err != nil {
			t.Fatal(err)
		}
	}
}
