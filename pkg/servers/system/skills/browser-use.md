---
name: browser-use
description: Automates browser interactions for web testing, form filling, screenshots, and data extraction. Use when the user needs to navigate websites, interact with web pages, fill forms, take screenshots, or extract information from web pages. IMPORTANT - If webfetch fails, try this tool instead.
---

# Browser Automation with browser-use CLI

The `browser-use` command provides fast, persistent browser automation. It maintains browser sessions across commands, enabling complex multi-step workflows.

`browser-use` is already installed in the dedicated Nanobot agent image. Use the built-in CLI directly.

## Quick Start

```bash
browser-use open https://example.com           # Navigate to URL
browser-use state                              # Get page elements with indices
browser-use click 5                            # Click element by index
browser-use type "Hello World"                 # Type text
browser-use screenshot                         # Take screenshot
browser-use close                              # Close browser
```

## Core Workflow

1. **Navigate**: `browser-use open <url>` - Opens URL (starts browser if needed)
2. **Inspect**: `browser-use state` - Returns clickable elements with indices
3. **Interact**: Use indices from state to interact (`browser-use click 5`, `browser-use input 3 "text"`)
4. **Verify**: `browser-use state` or `browser-use screenshot` to confirm actions
5. **Repeat**: Browser stays open between commands

## Browser Modes

```bash
browser-use --browser chromium open <url>      # Default: headless Chromium
browser-use --browser chromium --headed open <url>  # Visible Chromium window
browser-use --browser real open <url>          # User's Chrome with login sessions
```

- **chromium**: Fast, isolated, headless by default
- **real**: Uses your Chrome with cookies, extensions, logged-in sessions

## Browser Viewing with VNC

The VNC server is automatically started and ready to use. You can view and interact with the browser in real-time using VNC. This is **essential** for:

- **Solving CAPTCHAs** - Many sites require human verification
- **Debugging** - See exactly what the browser is doing
- **User interaction** - Manual intervention when automation gets stuck

The VNC server provides:
- X virtual framebuffer (Xvfb) on display :99
- VNC server on port 5900
- an internal WebSocket proxy on port 6080, exposed only through Nanobot's `/browser` path
- a borderless, maximized browser window sized for the BrowserView pane by default

### Accessing the Browser View

The Nanobot UI includes a browser viewer component that embeds VNC access. The backend proxies only the WebSocket transport; the standalone noVNC `vnc.html` page is not exposed.

### Using with browser-use

**Always use `--headed` mode when you need VNC viewing:**

```bash
# WITHOUT VNC (headless, fast, but user can't see it)
browser-use --browser chromium open https://example.com

# WITH VNC (visible in VNC viewer, user can interact)
browser-use --browser chromium --headed open https://example.com
```

The `--headed` flag makes the browser visible on the VNC display (:99).

### CAPTCHA Solving Workflow

**When a CAPTCHA is encountered:**

1. **Detect the CAPTCHA** - Look for verification challenges, "I'm not a robot" checkboxes, or automation blocks
2. **Inform the user immediately** - Tell them a CAPTCHA needs to be solved
3. **Ensure VNC is running** - Verify the browser is running in `--headed` mode
4. **Guide the user** - Tell them to open the BrowserView pane in the Nanobot UI
5. **Wait for completion** - Pause automation until the user confirms they've solved the CAPTCHA
6. **Continue** - Resume automation after manual intervention

**Example:**
```bash
# You're automating a task and hit a CAPTCHA
browser-use --browser chromium --headed open https://example.com
browser-use state
# → Detects CAPTCHA verification challenge

# STOP and inform user:
# "I've encountered a CAPTCHA that needs human verification.
#  Please open the BrowserView pane in the Nanobot UI
#  and solve the CAPTCHA. Let me know when you're done."

# Wait for user confirmation, then continue:
browser-use state  # Verify CAPTCHA is solved
browser-use click 5  # Continue with automation
```

**Important:** CAPTCHAs cannot be automated - they require human interaction. Always use `--headed` mode when you anticipate CAPTCHAs, and be prepared to pause and ask the user for help.

## Commands

### Navigation
```bash
browser-use open <url>                    # Navigate to URL
browser-use back                          # Go back in history
browser-use scroll down                   # Scroll down
browser-use scroll up                     # Scroll up
```

### Page State
```bash
browser-use state                         # Get URL, title, and clickable elements
browser-use screenshot                    # Take screenshot (outputs base64)
browser-use screenshot path.png           # Save screenshot to file
browser-use screenshot --full path.png    # Full page screenshot
```

### Interactions (use indices from `browser-use state`)
```bash
browser-use click <index>                 # Click element
browser-use type "text"                   # Type text into focused element
browser-use input <index> "text"          # Click element, then type text
browser-use keys "Enter"                  # Send keyboard keys
browser-use keys "Control+a"              # Send key combination
browser-use select <index> "option"       # Select dropdown option
```

### Tab Management
```bash
browser-use switch <tab>                  # Switch to tab by index
browser-use close-tab                     # Close current tab
browser-use close-tab <tab>               # Close specific tab
```

### JavaScript & Data
```bash
browser-use eval "document.title"         # Execute JavaScript, return result
```

### Cookies
```bash
browser-use cookies get                   # Get all cookies
browser-use cookies get --url <url>       # Get cookies for specific URL
browser-use cookies set <name> <value>    # Set a cookie
browser-use cookies set name val --domain .example.com --secure --http-only
browser-use cookies clear                 # Clear all cookies
browser-use cookies clear --url <url>     # Clear cookies for specific URL
browser-use cookies export <file>         # Export all cookies to JSON file
browser-use cookies export <file> --url <url>  # Export cookies for specific URL
browser-use cookies import <file>         # Import cookies from JSON file
```

### Wait Conditions
```bash
browser-use wait selector "h1"            # Wait for element to be visible
browser-use wait selector ".loading" --state hidden  # Wait for element to disappear
browser-use wait selector "#btn" --state attached    # Wait for element in DOM
browser-use wait text "Success"           # Wait for text to appear
browser-use wait selector "h1" --timeout 5000  # Custom timeout in ms
```

### Additional Interactions
```bash
browser-use hover <index>                 # Hover over element (triggers CSS :hover)
browser-use dblclick <index>              # Double-click element
browser-use rightclick <index>            # Right-click element (context menu)
```

### Information Retrieval
```bash
browser-use get title                     # Get page title
browser-use get html                      # Get full page HTML
browser-use get html --selector "h1"      # Get HTML of specific element
browser-use get text <index>              # Get text content of element
browser-use get value <index>             # Get value of input/textarea
browser-use get attributes <index>        # Get all attributes of element
browser-use get bbox <index>              # Get bounding box (x, y, width, height)
```

### Python Execution (Persistent Session)
```bash
browser-use python "x = 42"               # Set variable
browser-use python "print(x)"             # Access variable (outputs: 42)
browser-use python "print(browser.url)"   # Access browser object
browser-use python --vars                 # Show defined variables
browser-use python --reset                # Clear Python namespace
browser-use python --file script.py       # Execute Python file
```

The Python session maintains state across commands. The `browser` object provides:
- `browser.url` - Current page URL
- `browser.title` - Page title
- `browser.goto(url)` - Navigate
- `browser.click(index)` - Click element
- `browser.type(text)` - Type text
- `browser.screenshot(path)` - Take screenshot
- `browser.scroll()` - Scroll page
- `browser.html` - Get page HTML

### Session Management
```bash
browser-use sessions                      # List active sessions
browser-use close                         # Close current session
browser-use close --all                   # Close all sessions
```

### Profile Management
```bash
browser-use profile list-local            # List local Chrome profiles
```

**Before opening a real browser (`--browser real`)**, always ask the user if they want to use a specific Chrome profile or no profile. Use `profile list-local` to show available profiles:

```bash
browser-use profile list-local
# Output: Default: Person 1 (user@gmail.com)
#         Profile 1: Work (work@company.com)

# With a specific profile (has that profile's cookies/logins)
browser-use --browser real --profile "Profile 1" open https://gmail.com

# Without a profile (fresh browser, no existing logins)
browser-use --browser real open https://gmail.com

# Headless mode (no visible window) - useful for cookie export
browser-use --browser real --profile "Default" cookies export /tmp/cookies.json
```

Each Chrome profile has its own cookies, history, and logged-in sessions. Choosing the right profile determines whether sites will be pre-authenticated.

## Global Options

| Option | Description |
|--------|-------------|
| `--session NAME` | Use named session (default: "default") |
| `--browser MODE` | Browser mode: chromium, real |
| `--headed` | Show browser window (chromium mode) |
| `--profile NAME` | Chrome profile (real mode only) |
| `--json` | Output as JSON |

**Session behavior**: All commands without `--session` use the same "default" session. The browser stays open and is reused across commands. Use `--session NAME` to run multiple browsers in parallel.

## Usage Notes

- Always run `browser-use state` before interacting so you have current element indices.
- Prefer `--browser chromium --headed` when the user needs to see the browser or may need to step in.
- Use `browser-use screenshot` or `browser-use state` to confirm each critical action.
- Close the session with `browser-use close` when the browser task is complete.
