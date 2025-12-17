export type Encoding = "utf-8" | "base64";

export type SandboxMetadata = Record<string, unknown>;

export interface Sandbox {
	readonly id: string;

	getMeta(): Promise<SandboxMetadata>;

	resolvePath(path: string): string;

	readFile(
		path: string,
		opts?: {
			encoding?: Encoding;
			limit?: number;
			offset?: number;
		},
	): Promise<{
		content: string;
		encoding: Encoding;
	} | null>;

	writeFile(
		path: string,
		content: string,
		opts?: {
			encoding?: Encoding;
		},
	): Promise<void>;

	deleteFile(path: string): Promise<void>;

	readdir(
		path: string,
		opts?: {
			cursor?: string;
			recursive?: boolean;
			limit?: number;
		},
	): Promise<{
		entries: Array<{
			name: string;
			isFile: boolean;
			isDirectory: boolean;
			size: number;
		}>;
		cursor?: string;
	}>;

	execute(
		command: string,
		args: string[],
		opts?: {
			cwd?: string;
			env?: Record<string, string>;
			outputByteLimit?: number;
		},
	): Promise<string>;

	kill(id: string, signal?: string): Promise<void>;

	output(id: string): Promise<{
		output: string;
		truncated: boolean;
		exitCode: number;
		signal?: string;
	}>;

	wait(id: string): Promise<{
		exitCode: number;
		signal?: string;
	}>;

	release(id: string): Promise<void>;
}

export interface CreateSandboxOptions {
	baseUri?: string;
	id?: string;
	meta?: Record<string, unknown>;
}

export interface Manager {
	createSandbox(opts?: CreateSandboxOptions): Promise<Sandbox>;

	listSandboxes(): Promise<Sandbox[]>;

	getSandbox(id: string): Promise<Sandbox | null>;

	deleteSandbox(id: string): Promise<void>;
}
