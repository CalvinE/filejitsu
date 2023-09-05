# base64 Command

This command will encode or decode base64 data and print it on `stdout`.

| Full Name | Short Name | Required | Description | Default |
|-----|-----|-----|-----|-----|
|`--input`|`-i`|Y|The input to base64 encode / decode as a string. If not provided `stdin` is used|`stdin`|
|`--decode`|`-d`|N|If provided the the command will attempt to base64 decode the input|`false`|
|`--useUrlEncoding`|`-u`|N|If provided The URL safe base64 encoding will be used instead of standard base64 encoding|`false`|
|`--omitPadding`|`-n`|N|If provided the base64 encoding / decoding will not either add padding (encoding) or parse padding (decoding).|`false`|
|`--omitEndingNewLine`|`-e`|N|If provided no new line character will be added to the base64 encoded / decoded output.|`false`|

## Examples

Base64 encode input text with standard base64 encoding

```bash
filejitsu b64 -i "hello there"
```

Base64 decode input text with standard base64 encoding

```bash
filejitsu b64 -d -i "aGVsbG8gdGhlcmU="
```

Base64 encode input text with url base64 encoding

```bash
filejitsu b64 -i "hello there" -u
```

Base64 decode input text with url base64 encoding

```bash
filejitsu b64 -d -i "aGVsbG8gdGhlcmU=" -u
```

Base64 input from `stdin` with standard base64 encoding

```bash
echo "oh snap" | filejitsu b64
```

Base64 input from `stdin` with standard base64 encoding and put it into a file

```bash
echo "oh snap" | filejitsu b64 > base64.txt
```
