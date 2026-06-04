# svuplib

`svuplib` is the library version of `svup`.

It uploads a file to IPFS through Pinata and returns the CID, gateway URL, file name, size, and upload time.

## Install

```bash
go get github.com/derekhandy/svuplib
```

## Use

Set the Pinata keys before running your program:

```bash
export PINATA_API_KEY="your_api_key"
export PINATA_API_SECRET="your_secret_api_key"
```

On Windows PowerShell:

```powershell
$env:PINATA_API_KEY="your_api_key"
$env:PINATA_API_SECRET="your_secret_api_key"
```

Import the library as svup and call `Upload`. It reads `PINATA_API_KEY` and `PINATA_API_SECRET` from the environment.

```go
package main

import (
	"fmt"
	"log"

	svup "github.com/derekhandy/svuplib"
)

func main() {
	result, err := svup.Upload("path/to/file.jpg")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result.URL)
}
```

## Use without environment variables

If your app already has the keys in config, create an uploader directly:

```go
uploader := svup.NewPinataUploader(apiKey, apiSecret)

result, err := uploader.UploadFile("path/to/file.jpg", "")
if err != nil {
	return err
}

fmt.Println(result.Hash)
fmt.Println(result.URL)
```

Pass a second argument to `UploadFile` if you want to store the file under a different name:

```go
result, err := uploader.UploadFile("tmp/upload.bin", "avatar.png")
```

You can also upload bytes:

```go
result, err := uploader.UploadBytes(data, "avatar.png")
```