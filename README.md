# svuplib


<b> svuplib </b> is a cross-platform Go library for uploading files or byte data to IPFS through Pinata and returning the CID, gateway URL, filename, size, timestamp, and error metadata. Available for Linux, Windows and MacOS architectures.

Requires <b> Go 1.21 </b> or newer and a <b> Pinata </b> account.


```go
// Usage Example

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

## Install

Download from repository or add the module to an existing Go project:

```bash
go get github.com/derekhandy/svuplib
go mod tidy
```

## Set Environment Variables

Create Pinata API credentials at:

```text
https://app.pinata.cloud/
```

```bash
# Linux / Mac OS
export PINATA_API_KEY="your_api_key"
export PINATA_API_SECRET="your_secret_api_key"
```

```powershell
# Windows PowerShell
$env:PINATA_API_KEY="your_api_key"
$env:PINATA_API_SECRET="your_secret_api_key"
```

## Use

There are two supported ways to use the library: environment-based upload and configured uploader instances.

```go
// Reads PINATA_API_KEY and PINATA_API_SECRET from the environment.
result, err := svup.Upload("path/to/file.jpg")
```

```go
// Uses API credentials already loaded by the calling application.
uploader := svup.NewPinataUploader(apiKey, apiSecret)
result, err := uploader.UploadFile("path/to/file.jpg", "")
```

## API

```go
// Uploads a file using PINATA_API_KEY and PINATA_API_SECRET.
svup.Upload(filePath string) (svup.UploadResult, error)

// Creates a reusable uploader with explicit credentials.
svup.NewPinataUploader(apiKey, apiSecret string) *svup.PinataUploader

// Uploads a file from disk.
uploader.UploadFile(filePath string, filename string) (*svup.UploadResult, error)

// Uploads file data from memory.
uploader.UploadBytes(data []byte, filename string) (*svup.UploadResult, error)

// Checks the configured Pinata credentials.
uploader.TestConnection() error
```

## Examples

```go
// Upload a file with an explicit stored filename.

result, err := uploader.UploadFile("tmp/upload.bin", "avatar.png")
if err != nil {
	return err
}

fmt.Println(result.Hash)
fmt.Println(result.URL)
```

```go
// Upload bytes from memory.

result, err := uploader.UploadBytes(data, "avatar.png")
if err != nil {
	return err
}

fmt.Println(result.URL)
```

## Info

`UploadPhoto` and `UploadPhotoFromBytes` are retained for compatibility with older callers. New code should prefer `UploadFile` and `UploadBytes`.

The default gateway is:

```text
https://gateway.pinata.cloud/ipfs/
```

## NOTICE

<b> This library sends file contents to Pinata using the credentials provided by the caller or shell environment. Calling applications are responsible for validating file contents, account permissions, gateway behavior, and whether returned URLs should be shared.</b>
