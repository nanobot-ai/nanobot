# @nanobot-ai/coder

Coder package for NanoMCP that provides enhanced workspace client functionality with document conversion support and advanced todo management tools.

## Features

### Enhanced readTextFile with Markitdown Conversion

The `readTextFile` function supports optional conversion of various file formats to markdown using the [markitdown](https://github.com/microsoft/markitdown) CLI tool.

#### Supported File Formats

When `convert: true` is specified, the following file formats will be automatically converted to markdown:

- **Documents**: PDF (`.pdf`), Word (`.docx`), PowerPoint (`.pptx`)
- **Spreadsheets**: Excel (`.xlsx`, `.xls`), CSV (`.csv`)
- **Images**: JPEG (`.jpg`, `.jpeg`), PNG (`.png`) - with OCR support
- **Audio**: WAV (`.wav`), MP3 (`.mp3`)
- **Web**: HTML (`.html`)
- **Archives**: ZIP (`.zip`)

#### Installation

To use the conversion feature, you need to install the markitdown CLI:

```bash
pip install markitdown
```

#### Usage

```typescript
import { readTextFile } from "@nanobot-ai/coder";

// Read a PDF and convert to markdown
const pdfContent = await readTextFile(
  "workspace-id",
  "/path/to/document.pdf",
  { convert: true }
);

// Read specific lines from a converted Excel file
const excelContent = await readTextFile(
  "workspace-id",
  "/path/to/spreadsheet.xlsx",
  {
    convert: true,
    line: 10,    // Start from line 10
    limit: 20    // Read 20 lines
  }
);

// Read a regular text file (no conversion)
const textContent = await readTextFile(
  "workspace-id",
  "/path/to/file.txt",
  {
    line: 0,
    limit: 100
  }
);
```

#### API Reference

```typescript
export async function readTextFile(
  workspaceId: string,
  path: string,
  options?: ReadTextFileOptions,
  url?: string
): Promise<string>
```

**Parameters:**

- `workspaceId` - The workspace identifier
- `path` - Path to the file to read
- `options` - Optional reading options:
  - `line?: number` - Line number to start reading from (0-based)
  - `limit?: number` - Maximum number of lines to read
  - `convert?: boolean` - If true, convert file to markdown using markitdown CLI
- `url` - Optional workspace server URL

**Returns:** Promise<string> - The file contents as a string

#### Behavior

1. If `convert: true` is specified and the file extension is supported:
   - The file is converted to markdown using markitdown CLI
   - The `line` and `limit` options are applied to the converted markdown
   - If conversion fails, it falls back to regular file reading

2. If `convert: false` or not specified, or the file extension is not supported:
   - The file is read normally using the workspace client
   - The `line` and `limit` options are applied directly

3. The conversion is done on the client side, allowing for flexible processing before the content is returned.

#### Error Handling

If markitdown conversion fails (e.g., markitdown is not installed or the file cannot be converted), the function will:
1. Log a warning to the console
2. Automatically fall back to regular file reading
3. Continue execution without throwing an error

This ensures graceful degradation when markitdown is not available.

## Examples

See `examples/markitdown-usage.ts` for comprehensive usage examples.

## Todo Management Tools

The coder package includes three advanced todo management tools for organizing and tracking tasks with hierarchy support:

### TodoCreate

Create new tasks with optional parent-child relationships:

```json
{
  "tasks": [
    {
      "title": "Implement feature X",
      "description": "Add new feature with error handling",
      "status": "pending"
    }
  ]
}
```

### TodoUpdate

Update existing tasks by ID:

```json
{
  "updates": [
    {
      "id": "task_123",
      "status": "in_progress",
      "activeForm": "Implementing feature X"
    }
  ]
}
```

### TodoList

List and filter tasks by status, parent, or plan:

```json
{
  "status": "in_progress"
}
```

### Features

- **Unique task IDs** for reliable updates
- **Parent-child hierarchies** for organizing complex work
- **Multiple plans** for different projects
- **Four statuses**: pending, in_progress, completed, blocked
- **Task ordering** with explicit order field
- **Detailed descriptions** separate from titles

See [docs/TODO_TOOLS.md](docs/TODO_TOOLS.md) for comprehensive documentation.

## Other Workspace Client Functions

This package also re-exports the standard workspace client functions:

- `getWorkspaceClient(workspaceId, url?)` - Get or create a workspace client
- `ensureConnected(workspaceId, url?)` - Ensure client is connected
- `closeWorkspaceClient(workspaceId)` - Close a specific client
- `closeAllClients()` - Close all clients

## License

MIT
