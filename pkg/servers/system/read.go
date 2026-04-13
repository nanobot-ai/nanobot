package system

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
	"github.com/nanobot-ai/nanobot/pkg/types"
)

const (
	defaultReadLimit = 2000
	maxLineLength    = 2000
	maxPDFPages      = 10
	maxImageBytes    = 10_000_000 // 10MB
	// maxReadTextBytes caps the size of a readText result. Beyond this, we return
	// a notice instructing the model to use bash to read relevant portions instead
	// of letting the generic tool-result truncator persist the output to disk.
	// Should be kept in sync with maxToolResultSize in pkg/agents/truncate.go.
	maxReadTextBytes = 50 * 1024 // 50 KiB
)

func readText(p ReadParams) (*mcp.CallToolResult, error) {
	if p.Pages != nil {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("pages is only supported for PDF files")
	}

	var offset int
	if p.Offset != nil {
		offset = *p.Offset
	}
	if offset < 0 {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("offset must be >= 0")
	}

	limit := defaultReadLimit
	if p.Limit != nil {
		limit = *p.Limit
	}
	if limit <= 0 {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("limit must be > 0")
	}

	file, err := os.Open(p.FilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}
	defer file.Close()

	var (
		result    strings.Builder
		linesRead int
	)
	reader := bufio.NewReader(file)
	lineNum := 1

	for {
		if linesRead >= limit {
			break
		}

		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			line = strings.TrimRight(line, "\n\r")

			if lineNum > offset {
				// Truncate long lines
				if len(line) > maxLineLength {
					line = line[:maxLineLength]
				}

				// Format with line number (cat -n style)
				fmt.Fprintf(&result, "%6d\t%s\n", lineNum, line)
				linesRead++
			}

			lineNum++
		}

		if err != nil {
			if err != io.EOF {
				return nil, fmt.Errorf("error reading file: %w", err)
			}
			break
		}
	}

	// Determine whether the loop stopped because of the line limit while file
	// content still remains. We only trip the bash hint for this case when the
	// caller relied on the default limit (no explicit offset or limit) — when
	// the caller is paginating intentionally, hitting the limit is expected.
	hitDefaultLineLimit := linesRead >= limit && p.Limit == nil && p.Offset == nil
	moreFileContent := false
	if hitDefaultLineLimit {
		if _, err := reader.Peek(1); err == nil {
			moreFileContent = true
		}
	}

	if result.Len() > maxReadTextBytes || moreFileContent {
		return tooLargeReadResult(p, result.Len(), linesRead), nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: result.String()}},
	}, nil
}

// tooLargeReadResult returns a CallToolResult that tells the model the read
// output couldn't return the full file — either because the built output
// exceeded maxReadTextBytes or because the line limit was hit with more file
// content remaining — and that it should call read again with offset and
// limit to fetch specific line ranges. The content is marked with
// SkipTruncationMetaKey so the generic truncator does not persist it to disk.
func tooLargeReadResult(p ReadParams, builtBytes, linesReturned int) *mcp.CallToolResult {
	var fileSize int64 = -1
	if info, err := os.Stat(p.FilePath); err == nil {
		fileSize = info.Size()
	}

	var sizeDesc string
	if fileSize >= 0 {
		sizeDesc = fmt.Sprintf("file size %d bytes; ", fileSize)
	}

	notice := fmt.Sprintf(
		"File %s is too large to return in full through the read tool "+
			"(%sreturned %d lines / %d bytes, with more content remaining). "+
			"Call read again with the `offset` and `limit` parameters to fetch a "+
			"specific line range (offset is the number of lines to skip from the "+
			"start of the file, so the first returned line is line offset+1; limit "+
			"is the maximum number of lines to return).",
		p.FilePath, sizeDesc, linesReturned, builtBytes,
	)

	return &mcp.CallToolResult{
		Content: []mcp.Content{{
			Type: "text",
			Text: notice,
			Meta: map[string]any{types.SkipTruncationMetaKey: true},
		}},
	}
}

func readImage(p ReadParams, mimeType string) (*mcp.CallToolResult, error) {
	if p.Offset != nil || p.Limit != nil || p.Pages != nil {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("offset, limit, and pages are not supported for image files")
	}

	info, err := os.Stat(p.FilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}
	if info.Size() > int64(maxImageBytes) {
		return nil, fmt.Errorf("file size %d B exceeds maximum allowed size %d B", info.Size(), maxImageBytes)
	}

	data, err := os.ReadFile(p.FilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{{
			Type:     "image",
			Data:     base64.StdEncoding.EncodeToString(data),
			MIMEType: mimeType,
			Meta:     map[string]any{types.SkipTruncationMetaKey: true},
		}},
	}, nil
}

func readPDF(ctx context.Context, p ReadParams) (*mcp.CallToolResult, error) {
	if p.Offset != nil || p.Limit != nil {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("offset and limit are not supported for PDF files")
	}

	if _, err := exec.LookPath("pdftoppm"); err != nil {
		return nil, fmt.Errorf("pdftoppm not found on PATH. Install poppler to read PDF files (e.g., brew install poppler, apt-get install poppler-utils)")
	}

	totalPages, err := pdfPageCount(ctx, p.FilePath)
	if err != nil {
		return nil, fmt.Errorf("could not determine PDF page count (install poppler-utils for pdfinfo): %w", err)
	}

	first, last, err := parsePagesRange(p.Pages, totalPages)
	if err != nil {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("%v", err)
	}

	content := []mcp.Content{
		{Type: "text", Text: fmt.Sprintf("PDF: %s (pages %d-%d of %d)", filepath.Base(p.FilePath), first, last, totalPages)},
	}
	for page := first; page <= last; page++ {
		data, err := exec.CommandContext(ctx, "pdftoppm",
			"-jpeg", "-jpegopt", "quality=85",
			"-f", strconv.Itoa(page), "-l", strconv.Itoa(page),
			"-scale-to", "1024", "-singlefile",
			p.FilePath,
		).Output()
		if err != nil {
			return nil, fmt.Errorf("failed to render page %d: %w", page, err)
		}
		content = append(content, mcp.Content{
			Type:     "image",
			Data:     base64.StdEncoding.EncodeToString(data),
			MIMEType: "image/jpeg",
			Meta:     map[string]any{types.SkipTruncationMetaKey: true},
		})
	}

	return &mcp.CallToolResult{Content: content}, nil
}

func pdfPageCount(ctx context.Context, filePath string) (int, error) {
	output, err := exec.CommandContext(ctx, "pdfinfo", filePath).Output()
	if err != nil {
		return 0, fmt.Errorf("failed to run pdfinfo: %w", err)
	}

	for line := range strings.SplitSeq(string(output), "\n") {
		if strings.HasPrefix(line, "Pages:") {
			count, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "Pages:")))
			if err != nil {
				return 0, fmt.Errorf("failed to parse page count: %w", err)
			}
			return count, nil
		}
	}

	return 0, fmt.Errorf("could not find page count in pdfinfo output")
}

func parsePagesRange(pages *string, totalPages int) (int, int, error) {
	if pages == nil {
		if totalPages > 10 {
			return 0, 0, fmt.Errorf(
				"PDF has %d pages, please specify a pages parameter (e.g., pages: \"1-5\"), maximum %d pages per request",
				totalPages, maxPDFPages)
		}
		return 1, totalPages, nil
	}

	parts := strings.SplitN(strings.TrimSpace(*pages), "-", 2)
	first, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid page number %q: %w", parts[0], err)
	}

	last := first
	if len(parts) == 2 {
		last, err = strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return 0, 0, fmt.Errorf("invalid page number %q: %w", parts[1], err)
		}
	}

	if first < 1 {
		return 0, 0, fmt.Errorf("page numbers must be >= 1, got %d", first)
	}
	if last < first {
		return 0, 0, fmt.Errorf("last page (%d) must be >= first page (%d)", last, first)
	}
	if first > totalPages {
		return 0, 0, fmt.Errorf("first page %d exceeds PDF page count %d", first, totalPages)
	}
	last = min(last, totalPages)
	if last-first+1 > maxPDFPages {
		return 0, 0, fmt.Errorf("requested %d pages, maximum is %d per request", last-first+1, maxPDFPages)
	}

	return first, last, nil
}
