import { Server } from "@nanobot-ai/nanomcp";
import DeleteFile from "./src/tools/fs-deletefile.ts";
import ListDir from "./src/tools/fs-listdir.ts";
import ReadTextFile from "./src/tools/fs-readtextfile.ts";
import ResolvePath from "./src/tools/fs-resolvepath.ts";
import WriteTextFile from "./src/tools/fs-writetextfile.ts";
import SessionCreate from "./src/tools/session-create.ts";
import SessionDelete from "./src/tools/session-delete.ts";
import TerminalCreate from "./src/tools/terminal-create.ts";
import TerminalKill from "./src/tools/terminal-kill.ts";
import TerminalOutput from "./src/tools/terminal-output.ts";
import TerminalRelease from "./src/tools/terminal-release.ts";
import TerminalWait from "./src/tools/terminal-wait.ts";

const server = new Server(
	{
		name: "Nanobot Workspace MCP Server",
		version: "0.0.1",
	},
	{
		tools: {
			listDir: ListDir,
			readTextFile: ReadTextFile,
			resolvePath: ResolvePath,
			deleteFile: DeleteFile,
			writeTextFile: WriteTextFile,
			terminalCreate: TerminalCreate,
			terminalKill: TerminalKill,
			terminalOutput: TerminalOutput,
			terminalRelease: TerminalRelease,
			terminalWait: TerminalWait,
			sessionCreate: SessionCreate,
			sessionDelete: SessionDelete,
		},
	},
);

export default server;

if (import.meta.main) {
	await server.serve(9011);
}
