---
title: Getting Started
---

This guide will help you quickly get MCP OAuth Proxy up and running.

## Prerequisites

- Docker installed on your system
- An OAuth provider account (Google, Microsoft, GitHub, etc.)
- A running MCP server to proxy to

## Quick Start

### 1. Setup OAuth Credentials

OAuth credentials (Client ID and Client Secret) are used by the proxy to authenticate with external providers on behalf of your users. When users sign in, the proxy uses these credentials to verify their identity and obtain access tokens for external services.

#### Google OAuth

1. **Create Project**: Go to [Google Cloud Console](https://console.cloud.google.com/) and create a new project or select an existing one
2. **Enable APIs**: In the API & Services dashboard, click "Enable APIs and Services" and enable:
   - Google+ API (for basic profile info)
   - Any additional APIs your MCP server needs (Gmail, Drive, etc.)
3. **Configure OAuth Consent Screen**:
   - Go to "OAuth consent screen" in the sidebar
   - Go to "Audience"
   - Choose "External" for user type
   - Add users email address that you authorize to use the application
4. **Create OAuth Client**:
   - Go to "Credentials" in the sidebar
   - Click "Create Credentials" > "OAuth 2.0 Client IDs"
   - Choose "Web application" as application type
   - Give your client a name (e.g., "MCP OAuth Proxy")
   - Under "Authorized JavaScript origins", add: `http://localhost:8080`
   - Under "Authorized redirect URIs", add: `http://localhost:8080/callback`
   - Click "Create"
   - A dialog will show your **Client ID** and **Client Secret** - copy both immediately. You will need to use these in the next step.

#### Microsoft OAuth

1. **Azure Portal**: Go to [Azure Portal](https://portal.azure.com/).
2. **App Registration**:
   - Click "App registrations" > "New registration"
   - Name your application
   - Set redirect URI to `http://localhost:8080/callback`
   - Click "Register"
3. **Configure Permissions**:
   - Go to "API permissions"
   - Add permissions for Microsoft Graph (e.g., `User.Read`, `Mail.Read`)
4. **Create Secret**:
   - Go to "Certificates & secrets"
   - Click "New client secret"
   - Copy the **Client Secret** (save immediately - it won't be shown again)
   - Copy the **Application (client) ID** from the Overview page

#### GitHub OAuth

1. **GitHub Settings**: Go to GitHub Settings > Developer settings > [OAuth Apps](https://github.com/settings/applications/new)
2. **Create OAuth App**:
   - **Application name**: Your app name
   - **Homepage URL**: Your application's homepage
   - **Authorization callback URL**: `http://localhost:8080/callback`
   - **Application description**: Optional description
3. **Get Credentials**:
   - Click "Register application"
   - Copy your **Client ID**
   - Generate a **Client Secret** and copy it immediately

### 2. Setup Your MCP Server

The OAuth proxy requires a streamable HTTP MCP server to proxy requests to. Your MCP server must:

- **Support Streamable HTTP transport** - Accept HTTP requests (not just stdio)
- **Be accessible** - Running and reachable from the proxy container

**Example MCP Server Setup:**

```bash
# Example using Obot's Gmail MCP server
git clone https://github.com/obot-platform/tools
cd google/gmail
uv run python -m obot_gmail_mcp.server
```

This will start the server at http://localhost:9000/mcp/gmail.

:::note
If you are using gmail mcp server, make sure you setup google oauth credentials as described in the [Google OAuth](#google-oauth) section.
:::

### 3. Run OAuth Proxy

:::note
Different Auth provider have different authorize urls. Make sure you use the correct one for your auth provider. Here is a list of authorize urls for different auth providers:

- Google: https://accounts.google.com
- Microsoft: https://login.microsoftonline.com/common/oauth2/v2.0/authorize
- GitHub: https://github.com/login/oauth/authorize

Make sure you are supplying the correct authorize url in the environment variables `OAUTH_AUTHORIZE_URL`.

:::

#### Option A: Docker

```bash
docker run -d --name mcp-oauth-proxy -p 8080:8080 \
  -e OAUTH_CLIENT_ID="your-client-id" \
  -e OAUTH_CLIENT_SECRET="your-client-secret" \
  -e OAUTH_AUTHORIZE_URL="https://accounts.google.com" \
  -e SCOPES_SUPPORTED="openid,email,profile,https://www.googleapis.com/auth/gmail.readonly" \
  -e MCP_SERVER_URL="http://localhost:9000/mcp/gmail" \
  -e ENCRYPTION_KEY="your-encryption-key" \
  ghcr.io/obot-platform/mcp-oauth-proxy:latest
```

#### Option B: CLI Binary

1. **Download Latest Release**:

   - Go to [GitHub Releases](https://github.com/obot-platform/mcp-oauth-proxy/releases)
   - Download the appropriate binary for your platform (Linux, macOS, Windows)

2. **Run with Environment Variables**:

   ```bash
   # Set environment variables
   export OAUTH_CLIENT_ID="your-client-id"
   export OAUTH_CLIENT_SECRET="your-client-secret"
   export OAUTH_AUTHORIZE_URL="https://accounts.google.com"
   export SCOPES_SUPPORTED="openid,email,profile,https://www.googleapis.com/auth/gmail.readonly"
   export MCP_SERVER_URL="http://localhost:9000/mcp/gmail"
   export ENCRYPTION_KEY="your-encryption-key"

   # Run the binary
   ./mcp-oauth-proxy
   ```

## Environment Variables

| Variable              | Required | Description                                               |
| --------------------- | -------- | --------------------------------------------------------- |
| `OAUTH_CLIENT_ID`     | ✅       | OAuth client ID from provider                             |
| `OAUTH_CLIENT_SECRET` | ✅       | OAuth client secret                                       |
| `OAUTH_AUTHORIZE_URL` | ✅       | Provider's base URL (e.g., `https://accounts.google.com`) |
| `SCOPES_SUPPORTED`    | ✅       | Comma-separated OAuth scopes                              |
| `MCP_SERVER_URL`      | ✅       | Your MCP server endpoint                                  |
| `DATABASE_DSN`        | ❌       | Database connection string (defaults to SQLite)           |
| `ENCRYPTION_KEY`      | ✅       | Base64-encoded 32-byte AES key                            |

:::note
You can generate a random encryption key using the following command:

```bash
openssl rand -base64 32
```

Make sure you are storing the encryption key in a secure way.

:::

:::note
By default, you can use the proxy with sqlite database and it will store the db locally in a file called `oauth-proxy.db`.

In production, it is recommended to use postgres database. You can use the following command to create a postgres database:

```bash
docker run -d --name mcp-oauth-proxy-postgres -p 5432:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=oauth-proxy postgres
```

Then you can set the `DATABASE_DSN` environment variable to the connection string of your postgres database.
For example:

```bash
DATABASE_DSN="postgresql://postgres:postgres@localhost:5432/oauth-proxy"
```

:::

## Verification

You can now test your setup by using vscode MCP.

1. **Create Workspace Configuration**: Create a `.vscode/mcp.json` file in your workspace folder:

```json
{
  "servers": {
    "oauth-gmail": {
      "type": "http",
      "url": "http://localhost:8080/mcp/gmail"
    }
  }
}
```

4. **Authentication Flow**:

- When VSCode first connects, it will open a browser for OAuth authentication
- Sign in with your Google account
- Grant the requested permissions
- VSCode will receive the access token and can now communicate with the Gmail MCP server

5. **Test the Connection**:

- Open copilot panel in vscode
- prompt to get the list of emails
- you should see the list of emails

## Example

These are the example mcp servers that run with the proxy to integrate with external services.

- [Gmail MCP Server](https://github.com/obot-platform/tools/tree/main/google/gmail)
- [Google Drive MCP Server](https://github.com/obot-platform/tools/tree/main/google/drive)
- [Google Calendar MCP Server](https://github.com/obot-platform/tools/tree/main/google/calendar)
- [Google Sheets MCP Server](https://github.com/obot-platform/tools/tree/main/google/sheets)
- [Outlook MCP Server](https://github.com/obot-platform/tools/tree/main/microsoft365/outlook-mcp)
