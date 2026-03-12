package session

import "testing"

func TestTrimBrowserPrefix(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"/browser":         "/",
		"/browser/":        "/",
		"/browser/resize":  "/resize",
		"/browser/healthz": "/healthz",
		"/browser/foo/bar": "/foo/bar",
		"/something-else":  "/something-else",
		"":                 "",
	}

	for input, expected := range cases {
		if actual := trimBrowserPrefix(input); actual != expected {
			t.Fatalf("trimBrowserPrefix(%q) = %q, want %q", input, actual, expected)
		}
	}
}

func TestParseResolution(t *testing.T) {
	t.Parallel()

	width, height, err := parseResolution("1920x1080")
	if err != nil {
		t.Fatalf("parseResolution returned error: %v", err)
	}
	if width != 1920 || height != 1080 {
		t.Fatalf("parseResolution returned %dx%d, want 1920x1080", width, height)
	}

	if _, _, err := parseResolution("invalid"); err == nil {
		t.Fatal("parseResolution should reject invalid input")
	}
}

func TestBrowserNormalizeSize(t *testing.T) {
	t.Parallel()

	handler := &browserProxyHandler{
		maxWidth:  1600,
		maxHeight: 1200,
	}

	width, height := handler.normalizeSize(319, 241)
	if width != 640 || height != 480 {
		t.Fatalf("normalizeSize should enforce minimums, got %dx%d", width, height)
	}

	width, height = handler.normalizeSize(1611, 1401)
	if width != 1600 || height != 1200 {
		t.Fatalf("normalizeSize should clamp to max resolution, got %dx%d", width, height)
	}

	width, height = handler.normalizeSize(801, 601)
	if width != 808 || height != 608 {
		t.Fatalf("normalizeSize should round to 8px increments, got %dx%d", width, height)
	}
}
