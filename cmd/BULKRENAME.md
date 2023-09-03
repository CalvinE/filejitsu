# bulkrename Command

This command renames files based on the parameters provided.

## Parameters

| Full Name | Short Name | Required | Description | Default |
|-----|-----|-----|-----|-----|
| `--rootPath` | `-p` | Y | The path in which to perform the rename operation. | `None` |
| `--targetRegex` | `-r` | Y | A regular expression to use when selecting files to rename. This can also contain named capture groups, and those named capture groups can be used in the `DestinationTemplate`. Anything in the file name will be altered to match the destination template. See examples below. | `None` |
| `--destinationTemplate` | `-d` | Y | A go text template to rename the part of file names that match the `TargetRegex`. Named capture groups in the `TargetRegex` can be used as values in the `DestinationTemplate`. | `None` |
| `--recursive` | `-s` | N | If true the application will also look into directories in the `RootPath` for files to rename. | `false` |
| `--test` | `-t` | N | If true the rename operation will not be performed. This can be used to do a dry run to make sure what you think is going to happen is what happens... | `false` |

## Example Usage

### Partial File Rename

Say you have some files named like `file_23819.txt` where the 23819 represents `2023-08-19` in a directory located `../test_files`. If you wanted to change the part of the name where the date is to a proper length format like `20230819` you can fun the following command

```bash
go run . bkrn -p ../test_files -r "(?P<year>\d{2})(?P<month>[0-1]?\\d)(?P<day>\\d{1,2})" -d "20{{ .year }}{{padLeft .month \"0\" 2}}{{padLeft .day \"0\" 2}}"
```

Which will rename any file matching the regex in the command in the following way `file_23819.txt => file_20230819.txt`

### Total File Rename

Now say with the same file you wanted to rewrite the files name completely.

```bash
go run . bkrn -p ../test_files -r ".*_(?P<year>\\d{2})(?P<month>[0-1]?\\d)(?P<day>\\d{1,2})(?P<extension>.*)" -d "20{{ .year }}{{padLeft .month \"0\" 2}}{{padLeft .day \"0\" 2}}_file{{ .extension }}"
```

Which will rename any file matching the regex in the command in the following way `file_23819.txt => 20230819_file.txt`
