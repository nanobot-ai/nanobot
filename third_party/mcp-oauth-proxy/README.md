# MCP OAuth Proxy

MCP OAuth Proxy is an open-source OAuth 2.1 proxy server that adds authentication and authorization to MCP (Model Context Protocol) servers.

## What is MCP OAuth Proxy?

MCP OAuth Proxy acts as a bridge between OAuth providers (Google, Microsoft, GitHub) and MCP servers, providing:

- **OAuth 2.1 Compliance** - Full OAuth 2.1 authorization server with PKCE support
- **MCP Integration** - Seamless proxy to MCP servers with user context injection
- **Multi-Provider Support** - Works with any OAuth 2.0 provider via auto-discovery
- **Database Flexibility** - PostgreSQL for production, SQLite for development

## Architecture

The proxy sits in front of your MCP server to handle OAuth 2.1 authentication and validate user identity from external providers.

Here's how it works:

1. **OAuth 2.1 Flow** - When a client needs access, the proxy redirects to an external auth provider (Google, Microsoft, GitHub) to verify user identity
2. **Token Issuance** - Once authentication is complete, the proxy issues an access token back to the client
3. **MCP Auth Compliance** - Follows the MCP authentication specification and works with any compatible MCP client
4. **Request Proxying** - Validates the access token and forwards authenticated requests to your MCP server
5. **User Context** - Sends necessary headers to the MCP server about user identity and access to external services based on the OAuth scopes you configured

## Headers Sent to MCP Server

When proxying requests to your MCP server, the OAuth proxy automatically injects the following headers with user information:

| Header                     | Description                               | Example                |
| -------------------------- | ----------------------------------------- | ---------------------- |
| `X-Forwarded-User`         | User ID from the OAuth provider           | `12345678901234567890` |
| `X-Forwarded-Email`        | User's email address                      | `user@example.com`     |
| `X-Forwarded-Name`         | User's display name                       | `John Doe`             |
| `X-Forwarded-Access-Token` | OAuth access token for external API calls | `ya29.a0ARrdaM...`     |

These headers allow your MCP server to:

- **Identify the user** making the request
- **Personalize responses** based on user information
- **Make authenticated API calls** to external services using the access token
- **Implement user-specific logic** and access controls

## Quick Start

### Prerequisites

- Docker installed on your system
- An OAuth provider account (Google, Microsoft, GitHub, etc.)
- A running MCP server to proxy to

### 1. Setup OAuth Credentials

OAuth credentials (Client ID and Client Secret) are used by the proxy to authenticate with external providers on behalf of your users.

#### Google OAuth

1. Go to [Google Cloud Console](https://console.cloud.google.com/) and create a new project
2. Enable Google+ API in "APIs & Services" > "Library"
3. Configure OAuth consent screen and add authorized users
4. Create OAuth Client:
   - Go to "Credentials" > "Create Credentials" > "OAuth 2.0 Client IDs"
   - Choose "Web application" type
   - Add `http://localhost:8080/callback` as redirect URI
   - Copy your **Client ID** and **Client Secret**

#### Microsoft OAuth

1. Go to [Azure Portal](https://portal.azure.com/) > "Azure Active Directory"
2. Register a new application with redirect URI `http://localhost:8080/callback`
3. Configure API permissions (Microsoft Graph: `User.Read`, `Mail.Read`)
4. Create client secret and copy **Application (client) ID** and **Client Secret**

#### GitHub OAuth

1. Go to GitHub Settings > Developer settings > [OAuth Apps](https://github.com/settings/applications/new)
2. Create OAuth App with callback URL `http://localhost:8080/callback`
3. Copy **Client ID** and generate **Client Secret**

### 2. Setup Your MCP Server

The OAuth proxy requires a streamable HTTP MCP server. Example using Obot's Gmail MCP server:

```bash
git clone https://github.com/obot-platform/tools
cd google/gmail
uv run python -m obot_gmail_mcp.server
```

This starts the server at `http://localhost:9000/mcp/gmail`.

### 3. Run OAuth Proxy

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

1. Download from [GitHub Releases](https://github.com/obot-platform/mcp-oauth-proxy/releases)
2. Run with environment variables:

```bash
export OAUTH_CLIENT_ID="your-client-id"
export OAUTH_CLIENT_SECRET="your-client-secret"
export OAUTH_AUTHORIZE_URL="https://accounts.google.com"
export SCOPES_SUPPORTED="openid,email,profile,https://www.googleapis.com/auth/gmail.readonly"
export MCP_SERVER_URL="http://localhost:9000/mcp/gmail"
export ENCRYPTION_KEY="your-encryption-key"

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

You should generate a random 32-byte AES key for the `ENCRYPTION_KEY` environment variable using the following command:

```bash
openssl rand -base64 32
```

**Different Auth Provider URLs:**

- Google: `https://accounts.google.com`
- Microsoft: `https://login.microsoftonline.com/common/oauth2/v2.0/authorize`
- GitHub: `https://github.com/login/oauth/authorize`

## VSCode Setup

Create a `.vscode/mcp.json` file in your workspace:

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

**Authentication Flow:**

- VSCode opens browser for OAuth authentication
- Sign in with your account and grant permissions
- VSCode receives access token and communicates with the Gmail MCP server
- Use Copilot panel to interact with your emails

## License

This project is licensed under the Apache License 2.0.
