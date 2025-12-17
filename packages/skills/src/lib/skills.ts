import * as path from "node:path";
import type { WorkspaceClient } from "@nanobot-ai/workspace-client";

export type Skill = {
	name: string;
	description: string;
	instructions: string;
	baseDir: string;
};

const emptySkill: Skill = {
	name: "",
	description: "",
	instructions: "",
	baseDir: "",
};

function parseYAMLFrontMatter(text: string): {
	frontMatter: Record<string, string>;
	instructions: string;
} {
	const match = text.match(/^---\n([\s\S]*?)\n---\n([\s\S]*)$/);
	if (!match) return { frontMatter: {}, instructions: text.trim() };
	const frontMatter = match[1].split(/\n/).reduce(
		(obj, line) => {
			const match = line.match(/^\s*([^:]+):\s*(.*)$/);
			if (match) obj[match[1]] = match[2];
			return obj;
		},
		{} as Record<string, string>,
	);
	return { frontMatter: frontMatter, instructions: match[2].trim() };
}

const skillsRoot = ".nanobot/skills";

export async function getSkill(client: WorkspaceClient, skillName: string) {
	const skill = await getSkillByDirectoryName(client, skillName);
	if (skill.name) {
		return skill;
	}

	// Check for skill where name doesn't match directory name
	const skills = await getSkills(client);
	return skills.find((s) => s.name === skillName) || emptySkill;
}

async function getSkillByDirectoryName(
	client: WorkspaceClient,
	skillName: string,
) {
	const skillDir = await client.resolvePath(path.join(skillsRoot, skillName));
	const content = await client.readTextFile(
		path.join(skillsRoot, skillName, "SKILL.md"),
		{ ignoreNotFound: true },
	);
	if (!content) {
		// Check for skill where name doesn't match directory name
		return emptySkill;
	}

	const { frontMatter, instructions: parsedContent } =
		parseYAMLFrontMatter(content);
	return {
		name: frontMatter.name || skillName,
		description: frontMatter.description,
		instructions: parsedContent,
		baseDir: skillDir,
	};
}

export async function getSkills(client: WorkspaceClient) {
	const dirEntries = await client.listDir(skillsRoot, {
		ignoreNotFound: true,
	});

	const skills = Promise.all(
		dirEntries.entries.map(async (entry): Promise<Skill> => {
			if (!entry.isDirectory) {
				return emptySkill;
			}

			return getSkillByDirectoryName(client, entry.name);
		}),
	);

	return (await skills).filter((x) => x.name);
}
