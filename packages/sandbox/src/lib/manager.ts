import type { FileStorage } from "@remix-run/file-storage";
import type { StandardSchemaV1 } from "@standard-schema/spec";
import { z } from "zod/v4-mini";
import type {
	Encoding,
	Sandbox,
	Manager as SandboxManager,
	SandboxMetadata,
} from "../sandbox.ts";

export type SandboxConfig = {
	id: string;
	meta?: SandboxMetadata;
	parentId?: string;
	baseUri?: string;
	driver: string;
	driverConfig?: Record<string, unknown>;
};

const sandboxConfigSchema: StandardSchemaV1<SandboxConfig> = z.object({
	id: z.string(),
	parentId: z.optional(z.string()),
	baseUri: z.optional(z.string()),
	driver: z.string(),
	driverConfig: z.optional(z.record(z.string(), z.unknown())),
	meta: z.optional(z.record(z.string(), z.unknown())),
});

export interface SandboxDriver {
	readonly name: string;

	createSandbox(config: SandboxConfig): Promise<Record<string, unknown>>;

	loadSandbox(
		config: SandboxConfig,
		opts?: { create?: boolean },
	): Promise<Sandbox>;

	deleteSandbox(config: SandboxConfig): Promise<void>;
}

export class Manager implements SandboxManager {
	readonly fs: FileStorage;
	readonly drivers: Record<string, SandboxDriver>;
	readonly defaultDriver: string;

	constructor(
		fs: FileStorage,
		opts?: { drivers?: Record<string, SandboxDriver>; defaultDriver?: string },
	) {
		this.fs = fs;
		this.drivers = opts?.drivers ?? {};
		this.defaultDriver = opts?.defaultDriver ?? "docker";
		if (!this.drivers[this.defaultDriver]) {
			throw new Error(`Driver ${this.defaultDriver} not found`);
		}
	}

	async loadSandbox(
		cfg: SandboxConfig,
		opts?: { create?: boolean },
	): Promise<Sandbox> {
		const driver = this.drivers[cfg.driver];
		if (!driver) {
			throw new Error(`Driver ${cfg.driver} not found`);
		}
		if (!cfg.driverConfig && !opts?.create) {
			throw new Error(`Sandbox ${cfg.id} not created yet`);
		}
		if (!cfg.driverConfig && opts?.create) {
			cfg.driverConfig = await driver.createSandbox(cfg);
			await this.fs.set(
				`./sandboxes/${cfg.id}.json`,
				new File([JSON.stringify(cfg)], `${cfg.id}.json`),
			);
		}
		return driver.loadSandbox(cfg, opts);
	}

	async createSandbox(opts?: {
		id?: string;
		parentId?: string;
		baseUri?: string;
		meta?: SandboxMetadata;
	}): Promise<Sandbox> {
		let driver = this.defaultDriver;
		let { parentId, baseUri } = opts ?? {};
		if (parentId) {
			const parentConfig = await this.#getSandboxConfig(parentId);
			if (parentConfig.driver) {
				driver = parentConfig.driver;
			}
			baseUri = undefined;
		}
		const id = opts?.id || `sb-${crypto.randomUUID()}`;
		const config: SandboxConfig = {
			id,
			driver,
			baseUri,
			parentId,
			meta: opts?.meta,
		};
		await this.fs.put(
			`./sandboxes/${id}.json`,
			new File([JSON.stringify(config)], `${id}.json`),
		);
		return new CreateOnDemandSandbox(config, this);
	}

	async deleteSandbox(id: string): Promise<void> {
		const config = await this.#getSandboxConfig(id, { allowNotFound: true });
		if (!config) {
			return;
		}
		if (config.driverConfig) {
			const driver = this.drivers[config.driver];
			if (!driver) {
				throw new Error(`Driver ${config.driver} not found`);
			}
			await driver.deleteSandbox(config);
		}
		return this.fs.remove(`./sandboxes/${id}.json`);
	}

	async #getSandboxConfig<T extends GetSandboxOptions>(
		id: string,
		opts?: T,
	): Promise<SandboxConfigResult<T>> {
		const file = await this.fs.get(`./sandboxes/${id}.json`);
		if (!file) {
			if (opts?.allowNotFound) {
				return null as SandboxConfigResult<T>;
			}
			throw new Error(`Sandbox ${id} not found`);
		}

		const cfg = await validate(
			sandboxConfigSchema,
			JSON.parse(await file.text()),
		);
		if (cfg.parentId && !cfg.baseUri) {
			const parentConfig = await this.#getSandboxConfig(cfg.parentId);
			if (parentConfig.baseUri) {
				cfg.baseUri = parentConfig.baseUri;
			}
		}

		return cfg;
	}

	async getSandbox(id: string): Promise<Sandbox | null> {
		const config = await this.#getSandboxConfig(id, { allowNotFound: true });
		if (!config) {
			return null;
		}
		return new CreateOnDemandSandbox(config, this);
	}

	async listSandboxes(): Promise<Sandbox[]> {
		const list = await this.fs.list({
			prefix: "./sandboxes/",
		});

		const sandboxes: Sandbox[] = [];

		for (const entry of list.files) {
			if (entry.key.endsWith(".json")) {
				// Extract id from "./sandboxes/sb-xxx.json" -> "sb-xxx"
				const parts = entry.key.split("/");
				const filename = parts[parts.length - 1];
				if (filename) {
					const id = filename.slice(0, -5); // Remove ".json"
					sandboxes.push(new LazySandbox(this, id));
				}
			}
		}

		return sandboxes;
	}
}

export async function validate<T extends StandardSchemaV1>(
	schema: T,
	input: StandardSchemaV1.InferInput<T>,
): Promise<StandardSchemaV1.InferOutput<T>> {
	let result = schema["~standard"].validate(input);
	if (result instanceof Promise) result = await result;
	if (result.issues) {
		throw new Error(JSON.stringify(result.issues, null, 2));
	}
	return result.value;
}

class CreateOnDemandSandbox implements Sandbox {
	readonly config: SandboxConfig;
	readonly manager: Manager;
	sandbox: Sandbox | undefined = undefined;

	constructor(config: SandboxConfig, manager: Manager) {
		this.config = config;
		this.manager = manager;
	}

	get id() {
		return this.config.id;
	}

	async getMeta() {
		return this.config.meta ?? {};
	}

	resolvePath(path: string): string {
		// For CreateOnDemandSandbox, we need to delegate to the underlying sandbox
		// But if it's not created yet, we can't know the workdir, so we assume absolute paths
		if (path.startsWith("/")) {
			return path;
		}
		// Default to /workspace if sandbox not created yet
		return `/workspace/${path}`;
	}

	async #create(read?: boolean): Promise<Sandbox> {
		if (this.sandbox) {
			return this.sandbox;
		}
		if (!read) {
			this.sandbox = await this.manager.loadSandbox(this.config, {
				create: true,
			});
			return this.sandbox;
		}
		if (this.config.driverConfig) {
			this.sandbox = await this.manager.loadSandbox(this.config);
			return this.sandbox;
		}

		// not created yet, use parent if possible
		if (this.config.parentId) {
			const parentSandbox = await this.manager.getSandbox(this.config.parentId);
			if (parentSandbox) {
				return parentSandbox;
			}
		}

		if (!this.config.baseUri) {
			return new EmptyOnDemandSandbox(this.config, this.manager);
		}

		this.sandbox = await this.manager.loadSandbox(this.config, {
			create: true,
		});
		return this.sandbox;
	}

	readFile(
		path: string,
		opts?: { encoding?: Encoding; limit?: number; offset?: number },
	): Promise<{
		content: string;
		encoding: Encoding;
	} | null> {
		return this.#create(true).then((sandbox) => sandbox.readFile(path, opts));
	}

	writeFile(
		path: string,
		content: string,
		opts?: { encoding?: Encoding },
	): Promise<void> {
		return this.#create().then((sandbox) =>
			sandbox.writeFile(path, content, opts),
		);
	}

	deleteFile(path: string): Promise<void> {
		return this.#create().then((sandbox) => sandbox.deleteFile(path));
	}

	readdir(
		path: string,
		opts?: { cursor?: string; recursive?: boolean; limit?: number },
	): Promise<{
		entries: Array<{
			name: string;
			isFile: boolean;
			isDirectory: boolean;
			size: number;
		}>;
		cursor?: string;
	}> {
		return this.#create(true).then((sandbox) => sandbox.readdir(path, opts));
	}

	execute(
		command: string,
		args: string[],
		opts?: {
			cwd?: string;
			env?: Record<string, string>;
			outputByteLimit?: number;
		},
	): Promise<string> {
		return this.#create().then((sandbox) =>
			sandbox.execute(command, args, opts),
		);
	}

	kill(id: string, signal?: string): Promise<void> {
		return this.#create().then((sandbox) => sandbox.kill(id, signal));
	}

	output(id: string): Promise<{
		output: string;
		truncated: boolean;
		exitCode: number;
		signal?: string;
	}> {
		return this.#create().then((sandbox) => sandbox.output(id));
	}

	wait(id: string): Promise<{ exitCode: number; signal?: string }> {
		return this.#create().then((sandbox) => sandbox.wait(id));
	}

	release(id: string): Promise<void> {
		return this.#create().then((sandbox) => sandbox.release(id));
	}
}

class EmptyOnDemandSandbox extends CreateOnDemandSandbox {
	async readFile(
		path: string,
		opts?: { encoding?: Encoding; limit?: number; offset?: number },
	): Promise<{
		content: string;
		encoding: Encoding;
	} | null> {
		if (this.sandbox) {
			return super.readFile(path, opts);
		}
		return null;
	}

	async readdir(
		path: string,
		opts?: { cursor?: string; recursive?: boolean; limit?: number },
	): Promise<{
		entries: Array<{
			name: string;
			isFile: boolean;
			isDirectory: boolean;
			size: number;
		}>;
		cursor?: string;
	}> {
		if (this.sandbox) {
			return super.readdir(path, opts);
		}
		return { entries: [] };
	}
}

class LazySandbox implements Sandbox {
	readonly manager: Manager;
	readonly id: string;

	constructor(manager: Manager, id: string) {
		this.manager = manager;
		this.id = id;
	}

	async getMeta(): Promise<SandboxMetadata> {
		const sb = await this.#getSandbox();
		return sb.getMeta();
	}

	resolvePath(path: string): string {
		// For lazy sandbox, we can't wait for async getSandbox
		// so we return the default resolution like CreateOnDemandSandbox
		if (path.startsWith("/")) {
			return path;
		}
		return `/workspace/${path}`;
	}

	async #getSandbox(): Promise<Sandbox> {
		const sandbox = await this.manager.getSandbox(this.id);
		if (!sandbox) {
			throw new Error("Sandbox not found");
		}
		return sandbox;
	}

	deleteFile(path: string): Promise<void> {
		return this.#getSandbox().then((sandbox) => sandbox.deleteFile(path));
	}

	readdir(
		path: string,
		opts?: { cursor?: string; recursive?: boolean; limit?: number },
	): Promise<{
		entries: Array<{
			name: string;
			isFile: boolean;
			isDirectory: boolean;
			size: number;
		}>;
		cursor?: string;
	}> {
		return this.#getSandbox().then((sandbox) => sandbox.readdir(path, opts));
	}

	execute(
		command: string,
		args: string[],
		opts?: {
			cwd?: string;
			env?: Record<string, string>;
			outputByteLimit?: number;
		},
	): Promise<string> {
		return this.#getSandbox().then((sandbox) =>
			sandbox.execute(command, args, opts),
		);
	}

	kill(id: string, signal?: string): Promise<void> {
		return this.#getSandbox().then((sandbox) => sandbox.kill(id, signal));
	}

	output(id: string): Promise<{
		output: string;
		truncated: boolean;
		exitCode: number;
		signal?: string;
	}> {
		return this.#getSandbox().then((sandbox) => sandbox.output(id));
	}

	readFile(
		path: string,
		opts?: { encoding?: Encoding; limit?: number; offset?: number },
	): Promise<{
		content: string;
		encoding: Encoding;
	} | null> {
		return this.#getSandbox().then((sandbox) => sandbox.readFile(path, opts));
	}

	release(id: string): Promise<void> {
		return this.#getSandbox().then((sandbox) => sandbox.release(id));
	}

	wait(id: string): Promise<{ exitCode: number; signal?: string }> {
		return this.#getSandbox().then((sandbox) => sandbox.wait(id));
	}

	writeFile(
		path: string,
		content: string,
		opts?: { encoding?: Encoding },
	): Promise<void> {
		return this.#getSandbox().then((sandbox) =>
			sandbox.writeFile(path, content, opts),
		);
	}
}

interface GetSandboxOptions {
	allowNotFound?: boolean;
}

type SandboxConfigResult<T> = T extends { allowNotFound: true }
	? SandboxConfig | null
	: SandboxConfig;
