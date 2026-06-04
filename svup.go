// Package svup uploads files to IPFS through Pinata.
package svup

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// UploadResult is returned after an upload attempt.
type UploadResult struct {
	Success   bool      `json:"success"`
	Hash      string    `json:"hash"`
	URL       string    `json:"url"`
	Filename  string    `json:"filename"`
	Size      int64     `json:"size"`
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"error,omitempty"`
}

// PinataConfig stores the API keys and gateway used by PinataUploader.
type PinataConfig struct {
	APIKey    string
	APISecret string
	Gateway   string
}

type pinataResponse struct {
	IpfsHash  string `json:"IpfsHash"`
	PinSize   int    `json:"PinSize"`
	Timestamp string `json:"Timestamp"`
}

// PinataUploader uploads files to Pinata.
type PinataUploader struct {
	config     PinataConfig
	httpClient *http.Client
}

// Upload uploads a file using PINATA_API_KEY and PINATA_API_SECRET from the environment.
func Upload(filePath string) (UploadResult, error) {
	apiKey := os.Getenv("PINATA_API_KEY")
	apiSecret := os.Getenv("PINATA_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		return UploadResult{}, errors.New("pinata API credentials not found: set PINATA_API_KEY and PINATA_API_SECRET")
	}

	uploader := NewPinataUploader(apiKey, apiSecret)

	if err := uploader.TestConnection(); err != nil {
		return UploadResult{}, err
	}

	result, err := uploader.UploadFile(filePath, "")
	if err != nil {
		if result != nil {
			return *result, err
		}
		return UploadResult{}, err
	}

	return *result, nil
}

// NewPinataUploader returns a Pinata uploader using the default Pinata gateway.
func NewPinataUploader(apiKey, apiSecret string) *PinataUploader {
	return &PinataUploader{
		config: PinataConfig{
			APIKey:    apiKey,
			APISecret: apiSecret,
			Gateway:   "https://gateway.pinata.cloud/ipfs/",
		},
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// UploadFile uploads a file from disk.
func (p *PinataUploader) UploadFile(filePath string, filename string) (*UploadResult, error) {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return &UploadResult{
			Success: false,
			Error:   fmt.Sprintf("failed to read file: %v", err),
		}, err
	}

	if filename == "" {
		filename = filepath.Base(filePath)
	}

	return p.uploadFileData(fileData, filename)
}

// UploadPhoto uploads a file from disk. It is kept for compatibility with older callers.
func (p *PinataUploader) UploadPhoto(photoPath string, filename string) (*UploadResult, error) {
	return p.UploadFile(photoPath, filename)
}

// UploadBytes uploads file data from memory.
func (p *PinataUploader) UploadBytes(data []byte, filename string) (*UploadResult, error) {
	if filename == "" {
		filename = "file"
	}

	return p.uploadFileData(data, filename)
}

// UploadPhotoFromBytes uploads file data from memory. It is kept for compatibility with older callers.
func (p *PinataUploader) UploadPhotoFromBytes(data []byte, filename string) (*UploadResult, error) {
	return p.UploadBytes(data, filename)
}

func (p *PinataUploader) uploadFileData(data []byte, filename string) (*UploadResult, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	fileWriter, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return &UploadResult{
			Success: false,
			Error:   fmt.Sprintf("failed to create form file: %v", err),
		}, err
	}

	if _, err := fileWriter.Write(data); err != nil {
		return &UploadResult{
			Success: false,
			Error:   fmt.Sprintf("failed to write file data: %v", err),
		}, err
	}

	metadata := map[string]string{
		"name": filename,
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return &UploadResult{
			Success: false,
			Error:   fmt.Sprintf("failed to create metadata: %v", err),
		}, err
	}
	if err := writer.WriteField("pinataMetadata", string(metadataJSON)); err != nil {
		return &UploadResult{
			Success: false,
			Error:   fmt.Sprintf("failed to write metadata: %v", err),
		}, err
	}

	options := map[string]interface{}{
		"cidVersion": 0,
	}
	optionsJSON, err := json.Marshal(options)
	if err != nil {
		return &UploadResult{
			Success: false,
			Error:   fmt.Sprintf("failed to create options: %v", err),
		}, err
	}
	if err := writer.WriteField("pinataOptions", string(optionsJSON)); err != nil {
		return &UploadResult{
			Success: false,
			Error:   fmt.Sprintf("failed to write options: %v", err),
		}, err
	}

	if err := writer.Close(); err != nil {
		return &UploadResult{
			Success: false,
			Error:   fmt.Sprintf("failed to close upload body: %v", err),
		}, err
	}

	req, err := http.NewRequest("POST", "https://api.pinata.cloud/pinning/pinFileToIPFS", &buf)
	if err != nil {
		return &UploadResult{
			Success: false,
			Error:   fmt.Sprintf("failed to create request: %v", err),
		}, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("pinata_api_key", p.config.APIKey)
	req.Header.Set("pinata_secret_api_key", p.config.APISecret)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return &UploadResult{
			Success: false,
			Error:   fmt.Sprintf("failed to send request: %v", err),
		}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &UploadResult{
			Success: false,
			Error:   fmt.Sprintf("failed to read response: %v", err),
		}, err
	}

	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("pinata API error (%d): %s", resp.StatusCode, string(body))
		return &UploadResult{
			Success: false,
			Error:   errMsg,
		}, errors.New(errMsg)
	}

	var result pinataResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return &UploadResult{
			Success: false,
			Error:   fmt.Sprintf("failed to parse response: %v", err),
		}, err
	}

	url := p.config.Gateway + result.IpfsHash

	return &UploadResult{
		Success:   true,
		Hash:      result.IpfsHash,
		URL:       url,
		Filename:  filename,
		Size:      int64(len(data)),
		Timestamp: time.Now(),
	}, nil
}

// TestConnection checks that the configured Pinata API keys work.
func (p *PinataUploader) TestConnection() error {
	req, err := http.NewRequest("GET", "https://api.pinata.cloud/data/testAuthentication", nil)
	if err != nil {
		return fmt.Errorf("failed to create test request: %v", err)
	}

	req.Header.Set("pinata_api_key", p.config.APIKey)
	req.Header.Set("pinata_secret_api_key", p.config.APISecret)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Pinata API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("pinata API authentication failed (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}
