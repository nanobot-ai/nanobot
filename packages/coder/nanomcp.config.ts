import { defineConfig } from "@nanobot-ai/nanomcp";

// Import all tools
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

export default defineConfig({
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
});
