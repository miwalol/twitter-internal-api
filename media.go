package twitterinternalapi

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

const mediaUploadURL = "https://upload.x.com/i/media/upload.json"

// MediaUploadResponse represents the response from media upload operations
type MediaUploadResponse struct {
	MediaID       int64      `json:"media_id"`
	MediaIDString string     `json:"media_id_string"`
	MediaKey      string     `json:"media_key"`
	ExpiresAfter  int        `json:"expires_after_secs"`
	Size          int        `json:"size,omitempty"`
	Image         *ImageInfo `json:"image,omitempty"`
}

// ImageInfo contains image metadata
type ImageInfo struct {
	ImageType string `json:"image_type"`
	Width     int    `json:"w"`
	Height    int    `json:"h"`
}

// UploadMedia uploads a media file from a file path and returns the media ID
func (c *Client) UploadMedia(filePath string, mediaType string) (string, error) {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return c.UploadMediaBytes(fileData, mediaType)
}

// UploadMediaBytes uploads media from a byte slice and returns the media ID
func (c *Client) UploadMediaBytes(data []byte, mediaType string) (string, error) {
	// Step 1: INIT
	mediaID, err := c.uploadInit(int64(len(data)), mediaType)
	if err != nil {
		return "", fmt.Errorf("failed to init upload: %w", err)
	}

	// Step 2: APPEND
	if err = c.uploadAppend(mediaID, data); err != nil {
		return "", fmt.Errorf("failed to append media: %w", err)
	}

	// Step 3: FINALIZE
	hash := md5.Sum(data)
	md5Hash := fmt.Sprintf("%x", hash)
	if err = c.uploadFinalize(mediaID, md5Hash); err != nil {
		return "", fmt.Errorf("failed to finalize upload: %w", err)
	}

	return strconv.FormatInt(mediaID, 10), nil
}

// uploadInit initializes a media upload
func (c *Client) uploadInit(totalBytes int64, mediaType string) (int64, error) {
	params := url.Values{}
	params.Set("command", "INIT")
	params.Set("total_bytes", strconv.FormatInt(totalBytes, 10))
	params.Set("media_type", mediaType)
	params.Set("media_category", "tweet_image")

	uploadURL := fmt.Sprintf("%s?%s", mediaUploadURL, params.Encode())

	req, err := http.NewRequest("POST", uploadURL, nil)
	if err != nil {
		return 0, err
	}

	c.prepareRequest(req)
	c.applyCommonHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var result MediaUploadResponse
	if err = json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("failed to parse init response: %w", err)
	}

	return result.MediaID, nil
}

// uploadAppend appends media data to the upload
func (c *Client) uploadAppend(mediaID int64, fileData []byte) error {
	params := url.Values{}
	params.Set("command", "APPEND")
	params.Set("media_id", strconv.FormatInt(mediaID, 10))
	params.Set("segment_index", "0")

	uploadURL := fmt.Sprintf("%s?%s", mediaUploadURL, params.Encode())

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("media", "blob")
	if err != nil {
		return err
	}
	if _, err = part.Write(fileData); err != nil {
		return err
	}
	writer.Close()

	req, err := http.NewRequest("POST", uploadURL, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.prepareRequest(req)
	c.applyCommonHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("append failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// uploadFinalize finalizes the media upload
func (c *Client) uploadFinalize(mediaID int64, md5Hash string) error {
	params := url.Values{}
	params.Set("command", "FINALIZE")
	params.Set("media_id", strconv.FormatInt(mediaID, 10))
	params.Set("original_md5", md5Hash)

	uploadURL := fmt.Sprintf("%s?%s", mediaUploadURL, params.Encode())

	req, err := http.NewRequest("POST", uploadURL, nil)
	if err != nil {
		return err
	}

	c.prepareRequest(req)
	c.applyCommonHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("finalize failed (status %d): %s", resp.StatusCode, string(body))
	}

	var result MediaUploadResponse
	if err = json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse finalize response: %w", err)
	}

	return nil
}
