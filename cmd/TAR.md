# Tar Command

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
