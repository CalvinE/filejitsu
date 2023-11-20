# GZIP Commands

These are command to compress and decompress data with gzip.

## Commands

* `gzip` (gz) - compress data
* `gunzip` (guz) - decompress gzipped data

### GZIP Input / Output usage

The global `input` and `output` parameters are used in this command.

`input` is the content to be acted on, defaults to `stdin`. `inputText` can be used in lieu of the `input` parameter if you want to pass a string in without using a pipe `|` or other terminal output redirection.

`output` is where the output will go, defaults to `stdout`.

## GZIP Parameters

See global parameters for things like `input`, `output` or `logging` [here](../README.md).

| Full Name | Short Name | Required | Description | Default |
|-----|-----|-----|-----|-----|
| `--inputText` | `-t` | N | Text to be used for the base64 encode / decode. If not provided the global `input` parameter is used. | NONE |
| `--passphrase` | `-p` | Y** | A passphrase for the encryption process. | `NONE` |
| `--passphraseFile` | `-f` | Y** | The path to a file that will be used as the passphrase. | `NONE` |
