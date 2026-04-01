# Migration From Legacy Plaintext Identities

Older versions of `taxsend` used a plaintext `identity.txt` file on disk. The new workflow stores the receiver identity encrypted at rest in a local keystore.

## Migration Steps

### 1. Keep a backup of the legacy identity first

Do not delete the old file before you verify the new workflow.

Example legacy location:

```bash
~/.config/taxsend/identity.txt
```

### 2. Import it into the encrypted receiver keystore

```bash
taxsend receiver migrate \
  --name personal-laptop \
  --identity ~/.config/taxsend/identity.txt \
  --profile-out personal-laptop.public.json
```

You will be prompted for a new local passphrase for the keystore.

This command:

- reads the legacy plaintext identity
- stores it in the encrypted local keystore
- writes receiver metadata
- exports a new public sender profile JSON

It does not delete the legacy file automatically.

### 3. Re-import the public profile on the sending machine

On the work machine:

```bash
taxsend profile import personal-laptop.public.json --force
```

### 4. Verify the end-to-end workflow

On the work machine:

```bash
taxsend seal --to personal-laptop sample.txt
```

On the personal machine:

```bash
taxsend unseal --name personal-laptop attachment-YYYYMMDD-HHMMSS.bin
```

Or, if the output was split, pass any one of the chunk files.

### 5. Remove the plaintext identity after verification

Only after you have verified:

- the new receiver keystore works
- the public profile is imported on the sender side
- your backup and passphrase recovery story are in place

Then remove the old plaintext identity file.

## Low-Level Compatibility

The legacy low-level commands still work:

```bash
taxsend recipient --identity ~/.config/taxsend/identity.txt
taxsend decrypt --identity ~/.config/taxsend/identity.txt bundle.tar.age
```

They remain useful during migration and for recovery, but the recommended workflow is now:

- `receiver ...`
- `profile ...`
- `seal`
- `unseal`
