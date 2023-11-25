# Tar Command

## Commands

* `tar` - package or unpackage a tar archive with optional gzip compression and AES256 encryption

### Input / Output usage

The global `input` and `output` parameters are used in this command.

`input` (ONLY FOR UNPACKING TAR ARCHIVES) is the content to be acted on, defaults to `stdin`.

`output` (ONLY FOR CREATING TAR ARCHIVES) is where the output will go, defaults to `stdout`.

### Parameters

See global parameters for things like `input`, `output` or `logging` [here](../README.md).

| Full Name | Short Name | Required | Description | Default |
|-----|-----|-----|-----|-----|
| `--inputPath` | NA | N* | The input path to tar. Can be file or directory. Can be specified multiple times - (USED ONLY WITH CREATING A TAR ARCHIVE I.E. NO unpackage flag) | `NONE` |
| `--outputPath` | NA | N** | The output path to untar the contents of a tar archive to. Must be a directory - (USED ONLY WITH THE unpackage FLAG)
| `--useGzip` | `-z` | N | If present the contents being packaged will be gzipped or unpackaged will be gunzipped | `false` |
| `--compressionLevel` | `-q` | N | The compression level to use for gzip. Valid values are [ `NoCompression`, `BestSpeed`, `BestCompression`, `HuffmanOnly`, `DefaultCompression` ] | `DefaultCompression` |
| `--unpackage` | `-u` | N | If present the input tar package will be unpacked at the `outputPath` | `false` |
| `--encrypt` | `-e` | N | If present the tar will be encrypted while created, or decrypted while unpacked. Requires a passphrase or passphrase file be provided | `false` |
| `--passphrase` | `-p` | N*** | The passphrase used to encrypt or decrypt the data | `None` |
| `--passphraseFile` | `-f` | N*** | The file which will be read to get the passphrase used for encryption or decryption | `None` |

\* Required only if creating a tar archive (NA for unpacking a tar)
** Required only for unpack a tar archive (NA for creating a tar archive)
*** If `--encrypt` is provided then either `--passphrase` or `--passphraseFile` are required

## Example Commands

### Tar and encrypt the tar

```bash
go run main.go tar -z -o output.tar.gz ./test_files/
go run main.go encr -i output.tar.gz -p test -o output.tar.gz.enc
```

### Decrypt the tar and unpack

```bash
go run main.go dcry -i output.tar.gz.enc -o output.tar.gz.denc -p test
go run main.go tar -u -z -i output.tar.gz.denc ./test2_files
```
