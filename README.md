# Filejitsu

## Global Flags

| Full Name | Short Name | Required | Description | Default |
|-----|-----|-----|-----|-----|
| `--logLevel` | `-l` | N | The log level for the command being run. Default is to not log. Logs are emitted to `stderr`. Valid values are `error`, `warn`, `info`, `debug` and `none` | `none` |

## Commands

|Command|Short Name|Readme Link|Description|
|-----|-----|-----|-----|
|bulk-rename|bkrn|[BulkRename Command Details](./cmd/BULKRENAME.md)|A bulk file rename utility that will let you use regular expressions (with capture groups) and go text templates to leverage rich bulk rename functionality.|
|encrypt|encr|[Encrypt / Decrypt Command](./cmd/ENCRYPT_DECRYPT.md)|Encrypt data with AES-256.|
|decrypt|dcry|[Encrypt / Decrypt Command](./cmd/ENCRYPT_DECRYPT.md)|Decrypt AES-256 encrypted data.|
|base64|b64|[Base64 Encode / Decode](./cmd/BASE64.md)|Base 64 encode and decode input. Supports standard and url |
|space-analyzer|sa|[Space Analyzer](./cmd/SPACEANALYZER.md)|Analyzes files on disk. Can be used for a variety of purposes like seeing what taking up disk space, finding duplicate files (by content or by name), etc...|
