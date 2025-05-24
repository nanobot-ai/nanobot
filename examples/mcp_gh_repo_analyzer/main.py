from fastmcp import FastMCP, Client
from fastmcp.client.transports import StdioTransport

mcp = FastMCP(
    name="MCP Server Repo Analyzer",
    version="0.1.0",
    description="An MCP Server that discovers MCP servers on GitHub."
    )

@mcp.tool()
async def test_mcp_server(command: str, args: list[str] = [],env: dict[str, str] = {}):
    """_summary_

    Args:
        command (str): _description_
        args (list[str], optional): _description_. Defaults to [].
        env (dict[str, str], optional): _description_. Pull the env vars from the env with the same name. Defaults to {}.

    Returns:
        _type_: _description_
    """
    client = Client(transport=StdioTransport(command=command, args=args, env=env))
    async with client:
        print(f"Client is connected: {client.is_connected()}")
        return await client.list_tools()

if __name__ == "__main__":
    mcp.run()