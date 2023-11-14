# space-analyzer Command

This command enumerates the contents of a directory for analysis. The output is currently JSON and is intended to be analyzed in another application (or sub commands on this command)

## Input / Output usage

The global `input` and `output` parameters not used with this command.

## Parameters

See global parameters for things like `input`, `output` or `logging` [here](./GLOBAL.md).

| Full Name | Short Name | Required | Description | Default |
|-----|-----|-----|-----|-----|
| `--rootPath` | `-p` | N | The directory to perform analysis on. | `.` |
| `--maxRecursion` | `-m` | N | The max depth allowed in analysis. `-1` indicates that there is no limit. | `-1` |
| `--calculateFileHashes` | `-c` | N | If provided SHA512 hashes are calculated on all regular files. | not enabled |
| `--outputFormat` | `-f` | N | The desired output format. Supported values are `json` and `sjson` | `json` |

## Output Schema

## TODO

* Write the analysis feature (some kind of UI for reviewing the output of this) like a CUI.
* Add an ignore feature to pass over files and enumerating directories based on black list / regex?
