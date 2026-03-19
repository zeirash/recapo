package service

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/zeirash/recapo/arion/common/apierr"
)

// uploadImage detects the content type, generates a random filename, and uploads
// the image to R2 (or the local filesystem in dev). pathPrefix is the directory
// segment used both as the R2 object key prefix and the local upload sub-directory
// (e.g. "products", "feedback").
const maxImageSize = 5 * 1024 * 1024 // 5MB

func uploadImage(file io.Reader, pathPrefix string) (string, error) {
	limited := io.LimitReader(file, maxImageSize+1)

	buf := make([]byte, 512)
	n, err := limited.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}
	contentType := http.DetectContentType(buf[:n])

	var ext string
	switch contentType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/webp":
		ext = ".webp"
	default:
		return "", errors.New(apierr.ErrUnsupportedImageType)
	}

	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		return "", err
	}
	filename := fmt.Sprintf("%x%s", randBytes, ext)

	// Buffer the entire file content so the reader is seekable. The AWS SDK v2
	// reads the body once to compute a CRC32 checksum, then seeks back to the
	// start before sending. A non-seekable io.MultiReader causes the second read
	// to return empty data, producing a BadDigest error from Cloudflare R2.
	var bodyBuf bytes.Buffer
	bodyBuf.Write(buf[:n])
	if _, err := io.Copy(&bodyBuf, limited); err != nil {
		return "", err
	}
	if bodyBuf.Len() > maxImageSize {
		return "", errors.New(apierr.ErrImageTooLarge)
	}
	fullReader := bytes.NewReader(bodyBuf.Bytes())

	// Cloud path: upload to Cloudflare R2.
	if cfg.R2BucketName != "" {
		objectKey := pathPrefix + "/" + filename
		if err := r2UploadFunc(objectKey, fullReader, contentType); err != nil {
			return "", err
		}
		return cfg.R2PublicURL + "/" + objectKey, nil
	}

	// Local filesystem path.
	uploadDir := filepath.Join(cfg.UploadDir, pathPrefix)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	filePath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, fullReader); err != nil {
		os.Remove(filePath)
		return "", err
	}

	return "/uploads/" + pathPrefix + "/" + filename, nil
}
