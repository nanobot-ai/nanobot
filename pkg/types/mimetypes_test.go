package types

import "testing"

func TestResourceContentUseBlob(t *testing.T) {
	docxMIME := "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	fakeDocx := []byte{0x50, 0x4b, 0x03, 0x04, 0xff, 0xfe, 0xfd}
	utf8Text := []byte("hello 世界")

	tests := []struct {
		name     string
		mime     string
		content  []byte
		wantBlob bool
	}{
		{"text plain", "text/plain", utf8Text, false},
		{"json", "application/json", utf8Text, false},
		{"svg", "image/svg+xml", []byte("<svg></svg>"), false},
		{"png", "image/png", fakeDocx, true},
		{"pdf", "application/pdf", fakeDocx, true},
		{"docx mime", docxMIME, fakeDocx, true},
		{"ooxml prefix", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", fakeDocx, true},
		{"zip", "application/zip", fakeDocx, true},
		{"octet utf8 text", "application/octet-stream", utf8Text, false},
		{"octet binary", "application/octet-stream", fakeDocx, true},
		{"application yaml utf8", "application/x-yaml", []byte("key: val\n"), false},
		{"gif", "image/gif", fakeDocx, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResourceContentUseBlob(tt.mime, tt.content); got != tt.wantBlob {
				t.Errorf("ResourceContentUseBlob(%q, ...) = %v, want %v", tt.mime, got, tt.wantBlob)
			}
		})
	}
}
