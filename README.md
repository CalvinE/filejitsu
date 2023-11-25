# Filejitsu

## Note

Not all commands use the global parameters. Specifically `input` and `output`. The context of the command will indicate if these are used. Also the documentation for each command SHOULD indicate if either of these are used.

## Global Flags

| Full Name | Short Name | Required | Description | Default |
|-----|-----|-----|-----|-----|
| `--input` | `-i` | N | Some commands can receive input from outside sources. Files or `stdin`. This parameter sets the source of the input | `stdin` |
| `--logLevel` | `-l` | N | The log level to use for the command. `none` means no logs. Other levels are `debug`, `info`, `warn`, `error` | `none` |
| `--logOutput` | NA  | N| The destinations where the logs will be written. A file or something like `stderr` | `stderr` |
| `--output` | `-o`  | N| The destination for the output of the command. A file or something like `stdout` | `stdout` |

## Commands

|Command|Short Name|Readme Link|Description|
|-----|-----|-----|-----|
|bulk-rename|bkrn|[BulkRename Command Details](./cmd/BULKRENAME.md)|A bulk file rename utility that will let you use regular expressions (with capture groups) and go text templates to leverage rich bulk rename functionality.|
|encrypt|encr|[Encrypt / Decrypt Command](./cmd/ENCRYPT_DECRYPT.md)|Encrypt data with AES-256.|
|decrypt|dcry|[Encrypt / Decrypt Command](./cmd/ENCRYPT_DECRYPT.md)|Decrypt AES-256 encrypted data.|
|base64|b64|[Base64 Encode / Decode](./cmd/BASE64.md)|Base 64 encode and decode input. Supports standard and url |
|space-analyzer|sa|[Space Analyzer](./cmd/SPACEANALYZER.md)|Analyzes files on disk. Can be used for a variety of purposes like seeing what taking up disk space, finding duplicate files (by content or by name), etc...|
|gzip|gz|[GZIP Compress](./cmd/GZIP.md)|Gzip compression tool|
|gunzip|guz|[GZIP Decompress](./cmd/GZIP.md)|Gzip decompression tool|
|tar||[TAR utility](./cmd/TAR.md)|A tool for creating and unpacking TAR files. Also supports compression with gzip and encryption with AES-256|
