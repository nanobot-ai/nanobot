package completions

import (
	"testing"

	"github.com/obot-platform/nanobot/pkg/mcp"
	"github.com/obot-platform/nanobot/pkg/types"
)

func TestToRequestIncludesVideoResources(t *testing.T) {
	tests := []struct {
		name         string
		mimeType     string
		expectedMIME string
	}{
		{name: "MP4", mimeType: "video/mp4", expectedMIME: "video/mp4"},
		{name: "MOV", mimeType: "video/quicktime", expectedMIME: "video/mov"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := toRequest(&types.CompletionRequest{
				Model: "video-capable-model",
				Input: []types.Message{
					{
						Role: "user",
						Items: []types.CompletionItem{
							{
								Content: &mcp.Content{
									Type: "resource",
									Resource: &mcp.EmbeddedResource{
										URI:      "file:///clip",
										MIMEType: tt.mimeType,
										Blob:     "dmlkZW8=",
										Annotations: &mcp.ResourceAnnotations{
											Audience: []string{"assistant"},
										},
									},
								},
							},
						},
					},
				},
			})
			if err != nil {
				t.Fatalf("toRequest failed: %v", err)
			}

			parts := req.Messages[0].Content.ContentParts
			if len(parts) != 1 {
				t.Fatalf("expected one content part, got %d", len(parts))
			}
			if parts[0].Type != "video_url" || parts[0].VideoURL == nil {
				t.Fatalf("expected video_url content part, got %+v", parts[0])
			}
			expectedURL := "data:" + tt.expectedMIME + ";base64,dmlkZW8="
			if parts[0].VideoURL.URL != expectedURL {
				t.Fatalf("expected video URL %q, got %q", expectedURL, parts[0].VideoURL.URL)
			}
		})
	}
}
