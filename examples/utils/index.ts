import {
    type ServerRequest,
    type ServerNotification,
    type CreateMessageRequest, CreateMessageResultSchema, type CallToolResult
} from "@modelcontextprotocol/sdk/types.js";
import type {RequestHandlerExtra} from "@modelcontextprotocol/sdk/shared/protocol.js";

export function text(result: string): CallToolResult {
    return {
        content: [
            {
                type: "text",
                text: result,
            },
        ],
    };
}

export async function sampleText(ctx: RequestHandlerExtra<ServerRequest, ServerNotification>, model: string, prompt: string): Promise<string> {
    const request: CreateMessageRequest = {
        method: "sampling/createMessage",
        params: {
            messages: [
                {
                    role: "user",
                    content: {
                        type: "text",
                        text: prompt,
                    },
                },
            ],
            maxTokens: 1_000_000,
            modelPreferences: {
                hints: [{
                    name: model,
                }]
            }
        }
    }
    const response = await ctx.sendRequest(request, CreateMessageResultSchema)
    if (response.content?.type === "text") {
        return response.content.text
    }
    throw new Error("No response from model");
}
