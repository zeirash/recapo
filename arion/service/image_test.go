package service

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/zeirash/recapo/arion/common/apierr"
)

func Test_uploadImage(t *testing.T) {
	// Magic bytes for each supported type.
	// Note: Go's http.DetectContentType does not recognise webp, so there is no webp success case.
	jpegBytes := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01}
	pngBytes := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

	tmpDir := t.TempDir()
	oldCfg := cfg
	cfg.UploadDir = tmpDir
	defer func() { cfg = oldCfg }()

	tests := []struct {
		name         string
		file         []byte
		pathPrefix   string
		r2BucketName string // non-empty enables R2 path
		r2UploadErr  error
		wantURLPfx   string
		wantExt      string
		wantErr      bool
		wantErrMsg   string
	}{
		{
			name:       "jpeg saved to local filesystem",
			file:       jpegBytes,
			pathPrefix: "feedback",
			wantURLPfx: "/uploads/feedback/",
			wantExt:    ".jpg",
		},
		{
			name:       "png saved to local filesystem",
			file:       pngBytes,
			pathPrefix: "feedback",
			wantURLPfx: "/uploads/feedback/",
			wantExt:    ".png",
		},
		{
			name:       "uses pathPrefix in local URL",
			file:       jpegBytes,
			pathPrefix: "products",
			wantURLPfx: "/uploads/products/",
			wantExt:    ".jpg",
		},
		{
			name:       "returns error for unsupported file type",
			file:       []byte("hello plain text"),
			pathPrefix: "feedback",
			wantErr:    true,
			wantErrMsg: apierr.ErrUnsupportedImageType,
		},
		{
			name:         "uploads to R2 and returns R2 URL",
			file:         jpegBytes,
			pathPrefix:   "feedback",
			r2BucketName: "test-bucket",
			r2UploadErr:  nil,
			wantURLPfx:   "https://pub-test.r2.dev/feedback/",
			wantExt:      ".jpg",
		},
		{
			name:         "returns error when R2 upload fails",
			file:         jpegBytes,
			pathPrefix:   "feedback",
			r2BucketName: "test-bucket",
			r2UploadErr:  errors.New("R2 unavailable"),
			wantErr:      true,
			wantErrMsg:   "R2 unavailable",
		},
		{
			name:       "returns error when file reader fails",
			file:       nil, // triggers read error via errReader
			pathPrefix: "feedback",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.r2BucketName != "" {
				cfg.R2BucketName = tt.r2BucketName
				cfg.R2PublicURL = "https://pub-test.r2.dev"
				defer func() {
					cfg.R2BucketName = ""
					cfg.R2PublicURL = ""
				}()

				uploadErr := tt.r2UploadErr
				old := r2UploadFunc
				r2UploadFunc = func(key string, body io.Reader, contentType string) error {
					return uploadErr
				}
				defer func() { r2UploadFunc = old }()
			}

			var reader io.Reader
			if tt.file != nil {
				reader = bytes.NewReader(tt.file)
			} else {
				reader = &errReader{err: errors.New("read error")}
			}

			got, gotErr := uploadImage(reader, tt.pathPrefix)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("uploadImage() unexpected error: %v", gotErr)
				}
				if tt.wantErrMsg != "" && !strings.Contains(gotErr.Error(), tt.wantErrMsg) {
					t.Errorf("uploadImage() error = %v, want containing %q", gotErr, tt.wantErrMsg)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("uploadImage() succeeded unexpectedly")
			}
			if !strings.HasPrefix(got, tt.wantURLPfx) {
				t.Errorf("uploadImage() url = %q, want prefix %q", got, tt.wantURLPfx)
			}
			if tt.wantExt != "" && !strings.HasSuffix(got, tt.wantExt) {
				t.Errorf("uploadImage() url = %q, want suffix %q", got, tt.wantExt)
			}
		})
	}
}

// errReader is an io.Reader that always returns an error.
type errReader struct{ err error }

func (e *errReader) Read(p []byte) (int, error) { return 0, e.err }
