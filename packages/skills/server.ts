import { Server } from "@nanobot-ai/nanomcp";
import Skill from "./src/tools/skill.js";

const server = new Server(
	{
		name: "Nanobot Skills",
		version: "0.0.0",
	},
	{
		tools: {
			Skill,
		},
	},
);

export default server;

if (import.meta.main) {
	await server.serve(9012);
}
