# Global parameters

## Note

Not all commands use the global parameters. Specifically `input` and `output`. The context of the command will indicate if these are used. Also the documentation for each command SHOULD indicate if either of these are used.

## parameters

| Full Name | Short Name | Required | Description | Default |
|-----|-----|-----|-----|-----|
| `--input` | `-i` | N | Some commands can receive input from outside sources. Files or `stdin`. This parameter sets the source of the input | `stdin` |
| `--logLevel` | `-l` | The log level to use for the command. `none` means no logs. Other levels are `debug`, `info`, `warn`, `error` | `none` |
| `--logOutput` | NA | The destinations where the logs will be written. A file or something like `stderr` | `stderr` |
| `--output` | `-o` | The destination for the output of the command. A file or something like `stdout` | `stdout` |
