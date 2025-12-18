import { Server } from "@nanobot-ai/nanomcp";
import Bash from "./src/tools/bash.js";
import BashOutput from "./src/tools/bashoutput.js";
import Edit from "./src/tools/edit.js";
import Glob from "./src/tools/glob.js";
import Grep from "./src/tools/grep.js";
import KillShell from "./src/tools/killshell.js";
import Read from "./src/tools/read.js";
import TodoWrite from "./src/tools/todowrite.js";
import WebFetch from "./src/tools/webfetch.js";
import Write from "./src/tools/write.js";

const server = new Server(
	{
		name: "coder",
		version: "0.0.1",
	},
	{
		tools: {
			Bash,
			BashOutput,
			Edit,
			Glob,
			Grep,
			KillShell,
			Read,
			TodoWrite,
			WebFetch,
			Write,
		},
	},
);

export default server;

if (import.meta.main) {
	await server.serve(9013);
}
