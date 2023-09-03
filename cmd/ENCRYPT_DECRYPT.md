# Encrypt / Decrypt Commands

This is a command to encrypt or decrypt data. It operates on streams of bytes, so it can be a file, or piped in from stdin. Encryption is currently using `AES-256`. Might be open to adding others later if there is a good reason.

## Commands

* `encrypt` (encr) - encrypt data
* `decrypt` (dcry) - decrypt data

The parameters for both commands are identical.

## Parameters

| Full Name | Short Name | Required | Description | Default |
|-----|-----|-----|-----|-----|
| `--input` | `-i` | N | The input for the encryption or decryption process. Can be a file, or not provided stdin is used. | `_stdin` |
| `--output` | `-o` | N | The output for the encryption or decryption process. Can be a file, or not provided stdout is used. | `_stdout` |
| `--passphrase` | `-p` | Y** | A passphrase for the encryption process. | `NONE` |
| `--passphraseFile` | `-f` | Y** | The path to a file that will be used as the passphrase. | `NONE` |

** Either `--passphrase` or `--passphraseFile` is required

## Example command

Encrypt data from `stdin` and output to `stdout`

```bash
echo "this is a test" | go run ./... encr -p "test"
```

Encrypt data from `stdin` and then decrypt it.

```bash
echo "this is a test" | go run ./... encr -p "test" | go run ./... dcry -p "test"
```

Encrypt file from `stdin` and have output put into file from `stdout`.

```bash
echo "this is a test" | go run ./... encr -p "test" > file.enc
```

Cat encrypted file into decrypt and output decrypted data into file from `stdout`.

```bash
cat file.enc | go run ./... dcry -p "test" > file.clear
```
