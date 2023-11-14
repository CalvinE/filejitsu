# Encrypt / Decrypt Commands

This is a command to encrypt or decrypt data. It operates on streams of bytes, so it can be a file, or piped in from stdin. Encryption is currently using `AES-256`. Might be open to adding others later if there is a good reason.

## Commands

* `encrypt` (encr) - encrypt data
* `decrypt` (dcry) - decrypt data

The parameters for both commands are identical.

## Input / Output usage

The global `input` and `output` parameters are used in this command.

`input` is the content to be acted on, defaults to `stdin`. `inputText` can be used in lieu of the `input` parameter if you want to pass a string in without using a pipe `|` or other terminal output redirection.

`output` is where the output will go, defaults to `stdout`.

## Parameters

See global parameters for things like `input`, `output` or `logging` [here](../README.md).

| Full Name | Short Name | Required | Description | Default |
|-----|-----|-----|-----|-----|
| `--inputText` | `-t` | N | Text to be used for the base64 encode / decode. If not provided the global `input` parameter is used. | NONE |
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
