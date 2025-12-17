import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import { ensureConnected } from "@nanobot-ai/workspace-client";
import * as z from "zod";
import { getSkill, getSkills } from "../lib/skills.ts";

const schema = z.object({
	skill: z
		.string()
		.describe('The skill name (no arguments). E.g., "pdf" or "xlsx"'),
});

export default createTool({
	title: "Skill",
	enabled: async (ctx) => {
		const client = await ensureConnected(ctx.workspaceId);
		const skills = await getSkills(client);
		return skills.length > 0;
	},
	description: async (ctx) => {
		const client = await ensureConnected(ctx.workspaceId);
		const skills = await getSkills(client);
		const available = skills
			.map((s) => `${s.name}: ${s.description}`)
			.join("\n");
		return `Execute a skill within the main conversation

<skills_instructions>
When users ask you to perform tasks, check if any of the available skills below can help complete the task more effectively. Skills provide specialized capabilities and domain knowledge.

How to use skills:
- Invoke skills using this tool with the skill name only (no arguments)
- When you invoke a skill, you will see <command-message>The "{name}" skill is loading</command-message>
- The skill's prompt will expand and provide detailed instructions on how to complete the task
- Examples:
  - \`skill: "pdf"\` - invoke the pdf skill
  - \`skill: "xlsx"\` - invoke the xlsx skill
  - \`skill: "ms-office-suite:pdf"\` - invoke using fully qualified name

Important:
- Only use skills listed in <available_skills> below
- Do not invoke a skill that is already running
- Do not use this tool for built-in CLI commands (like /help, /clear, etc.)
</skills_instructions>

<available_skills>
${available}
</available_skills>
`;
	},
	messages: {
		invoking: "Loading skill",
		invoked: "Skill loaded",
	},
	inputSchema: schema,
	async handler({ skill: skillName }, ctx) {
		const client = await ensureConnected(ctx.workspaceId);
		const skill = await getSkill(client, skillName);
		if (!skill.name) {
			return toolResult.error(`skill not found: ${skillName}`);
		}
		return toolResult.text(
			`<command-message>The "${skill.name}" skill is loading</command-message>\n<command-name>${skill.name}</command-name>`,
			`The base directory for this skill is ${skill.baseDir}\n\n${skill.instructions}`,
		);
	},
});
