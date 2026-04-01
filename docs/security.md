# Security And Operations

## Threat Model

`taxsend` is designed for this model:

- encryption happens locally before email transport
- the sending machine only needs public recipients
- the receiving machine holds the private key material
- email stores and forwards ciphertext

This is a good fit for "store now, decrypt later."

## What The Tool Protects

`taxsend` protects the payload confidentiality of the bundled files.

It does this by:

- tar-streaming files directly into `age` encryption
- avoiding a plaintext `.tar` artifact on disk
- storing the new receiver identity in a local scrypt-protected keystore
- extracting decrypted files with private owner-only permissions by default

## What The Tool Does Not Protect

`taxsend` does not hide:

- the fact that an email was sent
- attachment size
- attachment filename
- sender, recipient, subject, or timestamps
- work-machine endpoint logs, attachment caches, or backup systems
- corporate DLP, transport rules, malware scanning, or retention rules

If the work machine or mail tenant is monitored, assume the transfer event is visible even when the payload is unreadable.

## Outlook / Exchange Constraints

Practical limits usually come from mail systems, not from `age`.

Important implications:

- high-level `seal` defaults to `.bin` output names instead of `.age`
- high-level `seal` defaults to `7 MiB` maximum part size
- chunking happens after encryption, so the cryptographic model stays the same

Chunking helps when attachment limits are the main problem. It does not help if:

- the tenant blocks all attachments above a stricter message policy
- the tenant blocks unsupported or suspicious attachments
- DLP or compliance policy rejects the message based on attachment handling rules

## Receiver Passphrase

The receiver passphrase protects the local keystore, not the email artifact.

Choose a passphrase that is:

- unique to this workflow
- long enough to resist guessing
- stored somewhere you can recover it later

If you lose both the keystore and the passphrase, you lose access to the encrypted artifacts.

## Backups

Back up both of these:

- the encrypted receiver keystore file
- the passphrase used to unlock it

Do not back up only one of them.

If you also keep a legacy plaintext identity during migration, remove it after you verify the new workflow and your backups.

## Optional Recovery Recipient

If you want a second recovery path, create a second receiver profile on another trusted device or offline backup location:

```bash
taxsend receiver init --name recovery --profile-out recovery.public.json
```

Then add the recovery recipient string from `recovery.public.json` into the primary public profile's `recipients` array before importing it on the sending machine.

That means the sender profile can encrypt to:

- your main personal receiver
- a second recovery receiver

The low-level fallback for one-off use is also available:

```bash
taxsend encrypt \
  --recipient <primary-recipient> \
  --recipient <recovery-recipient> \
  --output bundle.tar.age \
  files...
```

## Operational Recommendations

- Keep the receiver keystore only on personal-controlled machines.
- Never copy the private receiver material to the work machine.
- Use neutral attachment names.
- Prefer the smallest file set necessary for the transfer.
- Verify your recovery path before you depend on it.

`taxsend` can make the payload confidential. It cannot make the transfer invisible.
