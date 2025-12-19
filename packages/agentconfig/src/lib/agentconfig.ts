import { hooks } from "@nanobot-ai/nanomcp";
import {
	ensureConnected,
	type WorkspaceClient,
} from "@nanobot-ai/workspace-client";
import { parse as jsoncParse } from "jsonc-parser";

async function addInstructions(client: WorkspaceClient, agent: hooks.Agent) {
	const instructions = await client.readTextFile(".nanobot/agent.md", {
		ignoreNotFound: true,
	});
	if (instructions) {
		if (typeof agent.instructions === "string") {
			agent.instructions = `${agent.instructions}\n\n${instructions}`.trim();
		} else {
			agent.instructions = instructions.trim();
		}
	}
}

async function addMcpServers(
	client: WorkspaceClient,
	config: hooks.AgentConfigHook,
) {
	const mcpJson = await client.readTextFile(".nanobot/mcp.json", {
		ignoreNotFound: true,
	});
	if (!mcpJson) {
		return;
	}

	const parsed = hooks.AgentConfigHookSchema.safeParse({
		mcpServers: jsoncParse(mcpJson).mcpServers,
	});
	if (parsed.success) {
		config.mcpServers = parsed.data.mcpServers;
	} else {
		console.error(`Failed to parse MCP servers: ${parsed.error.message}`);
	}

	if (config.agent && !config.agent.resources) {
		config.agent.resources = [];
	}
}

export async function amendAgent(client: WorkspaceClient, agent: hooks.Agent) {
	await addInstructions(client, agent);
	return agent;
}

function getWorkspaceId(
	workspaceId: string,
	config: hooks.AgentConfigHook,
): string {
	return config._meta?.workspace?.id || config.sessionId || workspaceId;
}

export async function amendConfig(
	workspaceId: string,
	config: hooks.AgentConfigHook,
) {
	workspaceId = getWorkspaceId(workspaceId, config);
	console.log(
		`Original Agent Config: workspace=[${workspaceId}] ${JSON.stringify(config, null, 2)}`,
	);
	const client = await ensureConnected(workspaceId);

	if (config.agent) {
		await amendAgent(client, config.agent);
	}

	await addMcpServers(client, config);

	config.mcpServers = {
		...(config.mcpServers || {}),
		task: {
			url: "http://localhost:5173/mcp/tasks",
			headers: {
				"X-Nanobot-Workspace-Id": workspaceId,
			},
		},
		coder: {
			url: "http://localhost:5173/mcp/coder",
			headers: {
				"X-Nanobot-Workspace-Id": workspaceId,
			},
		},
		workspaceResources: {
			url: "http://localhost:5173/mcp/workspace-resources",
			headers: {
				"X-Nanobot-Workspace-Id": workspaceId,
			},
		},
	};

	if (config.agent) {
		config.agent.resources = [
			...(config.agent.resources ?? []),
			"workspaceResources",
		];
		config.agent.mcpServers = [
			...(config.agent.mcpServers ?? []),
			"task",
			"coder",
		];
	}

	return config;
}
