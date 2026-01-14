import { hooks } from "@nanobot-ai/nanomcp";
import {
	ensureConnected,
	type WorkspaceClient,
} from "@nanobot-ai/workspace-client";

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

	const parsedRaw = JSON.parse(mcpJson);

	const parsed = hooks.AgentConfigHookSchema.safeParse({
		agent: {},
		mcpServers: parsedRaw?.mcpServers,
	});
	if (parsed.success) {
		config.mcpServers = {
			...config.mcpServers,
			...parsed.data.mcpServers,
		};
		config.agent.mcpServers = [
			...(config.agent.mcpServers || []),
			...Object.keys(parsed.data?.mcpServers || {}),
		];
	} else {
		console.error(`Failed to parse MCP servers: ${parsed.error.message}`);
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

	const headers = {
		"X-Nanobot-Workspace-Id": workspaceId,
		...(config.sessionId && {
			"X-Nanobot-Session-Id": config.sessionId,
		}),
		...(config.accountId && {
			"X-Nanobot-Account-Id": config.accountId,
		}),
	};

	config.mcpServers = {
		...(config.mcpServers || {}),
		task: {
			url: "http://localhost:5173/mcp/tasks",
			headers,
		},
		skills: {
			url: "http://localhost:5173/mcp/skills",
			headers,
		},
		coder: {
			url: "http://localhost:5173/mcp/coder",
			headers,
		},
		workspaceResources: {
			url: "http://localhost:5173/mcp/workspace-resources",
			headers,
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
			"skills",
		];
	}

	await addMcpServers(client, config);

	return config;
}
