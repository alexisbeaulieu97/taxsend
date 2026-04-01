# taxsend

`taxsend` is a local-first CLI for encrypting sensitive files into `age` artifacts that can be emailed and decrypted later on a different machine.

The recommended workflow is:

1. Personal machine: create an encrypted local receiver keystore and export a public profile.
2. Work machine: import that public profile once.
3. Work machine: `seal` files into one neutral attachment or multiple email-safe parts.
4. Personal machine: `unseal` any part or the full artifact and extract the files locally.

`taxsend` does not rely on email for security. Email is only the transport for an encrypted blob.

## Build

```bash
go build ./cmd/taxsend
```

Or install it into your Go bin directory:

```bash
go install ./cmd/taxsend
```

## Recommended Workflow

### 1. Personal machine: initialize a receiver

This creates:

- an encrypted local keystore
- local receiver metadata
- a public profile JSON you can copy to the sending machine

```bash
taxsend receiver init \
  --name personal-laptop \
  --profile-out personal-laptop.public.json
```

You will be prompted once for a local passphrase and once to confirm it.

If you already have a legacy plaintext identity file, migrate it instead:

```bash
taxsend receiver migrate \
  --name personal-laptop \
  --identity ~/.config/taxsend/identity.txt \
  --profile-out personal-laptop.public.json
```

### 2. Work machine: import the public profile

```bash
taxsend profile import personal-laptop.public.json
```

List imported sender profiles:

```bash
taxsend profile list
```

### 3. Work machine: seal files for email

```bash
taxsend seal --to personal-laptop T4.pdf RL1.pdf
```

By default, `seal`:

- uses a neutral name like `attachment-20260401-120000.bin`
- splits automatically above `7 MiB`
- writes chunk files like `attachment-20260401-120000.part001.bin`

Useful flags:

```bash
taxsend seal \
  --to personal-laptop \
  --output-dir ./outbound \
  --basename 2025-packet \
  --max-part-size 7MiB \
  T4.pdf RL1.pdf slips/
```

### 4. Personal machine: unseal the artifact

You can pass a full artifact or any one chunk file. `taxsend` will discover and reassemble sibling parts automatically.

```bash
taxsend unseal \
  --name personal-laptop \
  --output-dir ./inbox \
  attachment-20260401-120000.part001.bin
```

You will be prompted once for the receiver passphrase. The identity is decrypted in memory only.

## Commands

High-level workflow commands:

- `receiver init`
- `receiver migrate`
- `receiver recipient`
- `profile import`
- `profile list`
- `profile remove`
- `seal`
- `unseal`

Low-level compatibility commands:

- `keygen`
- `recipient`
- `encrypt`
- `decrypt`
- `inspect`

The low-level commands remain available for manual workflows and migration, but the README and help output now favor `receiver` / `profile` / `seal` / `unseal`.

## Storage

`taxsend` stores local state under the platform user config directory in a `taxsend/` subdirectory.

It uses:

- `receivers/<name>.age` for encrypted local receiver keystores
- `receivers/<name>.json` for receiver metadata
- `profiles/<name>.json` for imported sender profiles

For tests or isolated runs, you can override the config root with `TAXSEND_CONFIG_DIR`.

## What Is Protected

Protected:

- file contents
- filenames inside the encrypted tar stream
- directory structure inside the encrypted tar stream
- receiver private key at rest, when using the new receiver keystore workflow

Not protected:

- the fact that an email was sent
- attachment name, size, and send time
- sender, recipient, subject, and other email metadata
- copies created by the work machine, mail client, or corporate controls
- employer monitoring, DLP, retention, or policy enforcement

## Outlook / Exchange Notes

- `taxsend` uses neutral `.bin` attachment names in the high-level workflow.
- The default chunk size is `7 MiB`, chosen to be conservative for environments with `10 MB` attachment limits and MIME overhead.
- Chunking helps with message-size limits. It does not bypass tenant attachment rules, unsupported-file rules, or DLP.

## More Docs

- [Security and operations](docs/security.md)
- [Migration from legacy plaintext identities](docs/migration.md)

## Development

Run tests with:

```bash
GOCACHE=$(pwd)/.gocache go test ./...
```
