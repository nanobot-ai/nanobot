import {McpServer} from "@modelcontextprotocol/sdk/server/mcp.js";
import {z} from "zod";
import {sampleText, text} from "../utils";
import {StdioServerTransport} from "@modelcontextprotocol/sdk/server/stdio.js";

const server = new McpServer({
    name: "Research Bot",
    description: "A bot that helps you find research papers.",
    version: "1.0.0",
});

type Queries = {
    searches: {
        reason: string;
        query: string;
    }[];
}

server.tool("research", {query: z.string()}, async ({query}, ctx) => {
    const queries = await sampleText(ctx, "planner", `Query: ${query}`);

    const parsedQueries = JSON.parse(queries) as Queries;

    const searches = await Promise.all(parsedQueries.searches.map(async search => {
        return sampleText(ctx, "searchAgent", `Search term: ${search.query}\nReason for searching: ${search.reason}`)
    }))

    const paper = await sampleText(ctx, "writer",
        `Original query: ${query}\nSummarized search results: ${JSON.stringify(searches)}\n\n`)

    return text(paper);
})

const transport = new StdioServerTransport()
await server.connect(transport)
