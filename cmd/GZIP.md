# GZIP Commands

These are command to compress and decompress data with gzip.

JWK RFC## GZIP / GUNZIP Input / Output usage

The global `input` and `output` parameters are used in this command.

`input` is the content to be acted on, defaults to `stdin`. `inputText` can be used in lieu of the `input` parameter if you want to pass a string in without using a pipe `|` or other terminal output redirection.

`output` is where the output will go, defaults to `stdout`.

## Commands

* `gzip` (gz) - compress data
* `gunzip` (guz) - decompress gzipped data

### GZIP Parameters

See global parameters for things like `input`, `output` or `logging` [here](../README.md).

| Full Name | Short Name | Required | Description | Default |
|-----|-----|-----|-----|-----|
| `--inputText` | `-t` | N | Text to be used for the gzip compression. If not provided the global `input` parameter is used. | `NONE` |
| `--compressionLevel` | `-q` | N | The compression level to use for gzip. Valid values are [ `NoCompression`, `BestSpeed`, `BestCompression`, `HuffmanOnly`, `DefaultCompression` ] | `DefaultCompression` |
| `--comment` | `-m` | N | The comment to place in the gzip stream header | `NONE` |
| `--name` | `-n` | N | The name to place in the gzip stream header | `NONE` |
| `--modTime` | NA | N | The comment to place in the gzip stream header | `NONE` |
| `--extra` | `-e` | N | The extra data to place in the gzip stream header | `NONE` |

### GUNZIP Parameters

See global parameters for things like `input`, `output` or `logging` [here](../README.md).

> **Note for the current implementation**: decompression does not currently handle the GZIP header. If this is desired I or  some industrious person can add it.

| Full Name | Short Name | Required | Description | Default |
|-----|-----|-----|-----|-----|
| `--inputText` | `-t` | N | Text to be used for the gzip decompression. If not provided the global `input` parameter is used. | `NONE` |
