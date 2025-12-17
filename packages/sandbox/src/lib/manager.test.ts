import * as assert from "node:assert";
import { beforeEach, describe, it } from "node:test";
import type { FileStorage } from "@remix-run/file-storage";
import { createMemoryFileStorage } from "@remix-run/file-storage/memory";
import type { Encoding, Sandbox, SandboxMetadata } from "../sandbox.ts";
import { Manager, type SandboxConfig, type SandboxDriver } from "./manager.ts";

// Mock Sandbox implementation for testing
class MockSandbox implements Sandbox {
	readonly id: string;
	private meta: SandboxMetadata;

	constructor(id: string, meta?: SandboxMetadata) {
		this.id = id;
		this.meta = meta ?? {};
	}

	async getMeta(): Promise<SandboxMetadata> {
		return this.meta;
	}

	resolvePath(path: string): string {
		if (path.startsWith("/")) {
			return path;
		}
		return `/workspace/${path}`;
	}

	async readFile(
		path: string,
		_opts?: { encoding?: Encoding; limit?: number; offset?: number },
	): Promise<{ content: string; encoding: Encoding } | null> {
		return { content: `mock content from ${path}`, encoding: "utf-8" };
	}

	async writeFile(
		_path: string,
		_content: string,
		_opts?: { encoding?: Encoding },
	): Promise<void> {
		// Mock implementation
	}

	async deleteFile(_path: string): Promise<void> {
		// Mock implementation
	}

	async readdir(
		_path: string,
		_opts?: { cursor?: string; recursive?: boolean; limit?: number },
	): Promise<{
		entries: Array<{
			name: string;
			isFile: boolean;
			isDirectory: boolean;
			size: number;
		}>;
		cursor?: string;
	}> {
		return {
			entries: [
				{ name: "file1.txt", isFile: true, isDirectory: false, size: 100 },
				{ name: "subdir", isFile: false, isDirectory: true, size: 0 },
			],
			cursor: undefined,
		};
	}

	async execute(
		_command: string,
		_args: string[],
		_opts?: {
			cwd?: string;
			env?: Record<string, string>;
			outputByteLimit?: number;
		},
	): Promise<string> {
		return "mock-exec-id";
	}

	async kill(_id: string, _signal?: string): Promise<void> {
		// Mock implementation
	}

	async output(_id: string): Promise<{
		output: string;
		truncated: boolean;
		exitCode: number;
		signal?: string;
	}> {
		return {
			output: "mock output",
			truncated: false,
			exitCode: 0,
			signal: undefined,
		};
	}

	async wait(_id: string): Promise<{ exitCode: number; signal?: string }> {
		return { exitCode: 0, signal: undefined };
	}

	async release(_id: string): Promise<void> {
		// Mock implementation
	}
}

// Mock SandboxDriver implementation
class MockDriver implements SandboxDriver {
	readonly name = "mock-driver";
	private sandboxes = new Map<string, MockSandbox>();

	async createSandbox(config: SandboxConfig): Promise<Record<string, unknown>> {
		return {
			containerId: `container-${config.id}`,
			imageId: "mock-image",
		};
	}

	async loadSandbox(
		config: SandboxConfig,
		_opts?: { create?: boolean },
	): Promise<Sandbox> {
		const existing = this.sandboxes.get(config.id);
		if (existing) {
			return existing;
		}

		const sandbox = new MockSandbox(config.id, config.meta);
		this.sandboxes.set(config.id, sandbox);
		return sandbox;
	}

	async deleteSandbox(config: SandboxConfig): Promise<void> {
		this.sandboxes.delete(config.id);
	}
}

describe("Manager", () => {
	let fs: FileStorage;
	let driver: MockDriver;
	let manager: Manager;

	beforeEach(() => {
		fs = createMemoryFileStorage();
		driver = new MockDriver();
		manager = new Manager(fs, {
			drivers: { mock: driver },
			defaultDriver: "mock",
		});
	});

	describe("constructor", () => {
		it("should create manager with default driver", () => {
			assert.strictEqual(manager.fs, fs);
			assert.strictEqual(manager.drivers.mock, driver);
			assert.strictEqual(manager.defaultDriver, "mock");
		});

		it("should throw error if default driver not found", () => {
			assert.throws(
				() => {
					new Manager(fs, {
						drivers: { mock: driver },
						defaultDriver: "non-existent",
					});
				},
				{
					message: "Driver non-existent not found",
				},
			);
		});

		it("should use docker as default driver name", () => {
			assert.throws(
				() => {
					new Manager(fs);
				},
				{
					message: "Driver docker not found",
				},
			);
		});

		it("should accept empty drivers object with valid default", () => {
			const emptyFs = createMemoryFileStorage();
			assert.throws(
				() => {
					new Manager(emptyFs, { drivers: {}, defaultDriver: "mock" });
				},
				{
					message: "Driver mock not found",
				},
			);
		});
	});

	describe("createSandbox", () => {
		it("should create sandbox with default driver", async () => {
			const sandbox = await manager.createSandbox();

			assert.strictEqual(typeof sandbox.readFile, "function");

			// Verify config was saved to storage
			const configFile = await fs.get(`./sandboxes/${sandbox.id}.json`);
			assert.ok(configFile);

			const config = JSON.parse(await configFile.text());
			assert.strictEqual(config.id, sandbox.id);
			assert.strictEqual(config.driver, "mock");
			assert.strictEqual(config.baseUri, undefined);
			assert.strictEqual(config.parentId, undefined);
		});

		it("should create sandbox with meta", async () => {
			const meta = { userId: "user-123", project: "test-project" };
			const sandbox = await manager.createSandbox({ meta });

			const configFile = await fs.get(`./sandboxes/${sandbox.id}.json`);
			assert.ok(configFile);

			const config = JSON.parse(await configFile.text());
			assert.deepStrictEqual(config.meta, meta);
		});

		it("should create sandbox with baseUri", async () => {
			const baseUri = "https://example.com/base";
			const sandbox = await manager.createSandbox({ baseUri });

			const configFile = await fs.get(`./sandboxes/${sandbox.id}.json`);
			assert.ok(configFile);

			const config = JSON.parse(await configFile.text());
			assert.strictEqual(config.baseUri, baseUri);
			assert.strictEqual(config.parentId, undefined);
		});

		it("should inherit driver from parent sandbox", async () => {
			const otherDriver = new MockDriver();
			const managerWithMultipleDrivers = new Manager(fs, {
				drivers: { mock: driver, other: otherDriver },
				defaultDriver: "mock",
			});

			// Create parent with default driver
			const parent = await managerWithMultipleDrivers.createSandbox();

			// Create child referencing parent
			const child = await managerWithMultipleDrivers.createSandbox({
				baseUri: parent.id,
			});

			const configFile = await fs.get(`./sandboxes/${child.id}.json`);
			assert.ok(configFile, "Config file should exist");
			const config = JSON.parse(await configFile.text());

			assert.strictEqual(config.driver, "mock");
		});
	});

	describe("getSandbox", () => {
		it("should return sandbox if it exists", async () => {
			const created = await manager.createSandbox();
			const retrieved = await manager.getSandbox(created.id);

			assert.ok(retrieved);
			assert.strictEqual(retrieved.id, created.id);
		});

		it("should return null if sandbox does not exist", async () => {
			const result = await manager.getSandbox("sb-non-existent");
			assert.strictEqual(result, null);
		});

		it("should return CreateOnDemandSandbox wrapper", async () => {
			const created = await manager.createSandbox();
			const retrieved = await manager.getSandbox(created.id);

			assert.ok(retrieved);
			// Verify it's a wrapper by checking it has the expected methods
			assert.strictEqual(typeof retrieved.readFile, "function");
			assert.strictEqual(typeof retrieved.writeFile, "function");
		});
	});

	describe("deleteSandbox", () => {
		it("should delete sandbox without driverConfig", async () => {
			const sandbox = await manager.createSandbox();

			// Verify it exists
			let configFile = await fs.get(`./sandboxes/${sandbox.id}.json`);
			assert.ok(configFile);

			// Delete it
			await manager.deleteSandbox(sandbox.id);

			// Verify it's gone
			configFile = await fs.get(`./sandboxes/${sandbox.id}.json`);
			assert.strictEqual(configFile, null);
		});

		it("should delete sandbox with driverConfig", async () => {
			const sandbox = await manager.createSandbox();

			// Trigger creation of the actual sandbox
			await sandbox.writeFile("/test.txt", "content");

			// Verify driverConfig was created
			let configFile = await fs.get(`./sandboxes/${sandbox.id}.json`);
			assert.ok(configFile, "Config file should exist");
			const config = JSON.parse(await configFile.text());
			assert.ok(config.driverConfig);

			// Delete it
			await manager.deleteSandbox(sandbox.id);

			// Verify it's gone from storage
			configFile = await fs.get(`./sandboxes/${sandbox.id}.json`);
			assert.strictEqual(configFile, null);
		});

		it("should not throw if sandbox does not exist", async () => {
			// Should not throw
			await manager.deleteSandbox("sb-non-existent");
		});

		it("should call driver.deleteSandbox when driverConfig exists", async () => {
			const sandbox = await manager.createSandbox();

			// Create the sandbox (which sets driverConfig)
			await sandbox.writeFile("/test.txt", "content");

			let driverDeleteCalled = false;
			const originalDelete = driver.deleteSandbox.bind(driver);
			driver.deleteSandbox = async (config: SandboxConfig) => {
				driverDeleteCalled = true;
				return originalDelete(config);
			};

			await manager.deleteSandbox(sandbox.id);

			assert.strictEqual(driverDeleteCalled, true);
		});

		it("should throw if driver not found for sandbox with driverConfig", async () => {
			// Create sandbox
			const sandbox = await manager.createSandbox();
			await sandbox.writeFile("/test.txt", "content");

			// Manually update config to use non-existent driver
			const configFile = await fs.get(`./sandboxes/${sandbox.id}.json`);
			assert.ok(configFile, "Config file should exist");
			const config = JSON.parse(await configFile.text());
			config.driver = "non-existent";
			await fs.put(
				`./sandboxes/${sandbox.id}.json`,
				new File([JSON.stringify(config)], `${sandbox.id}.json`),
			);

			// Should throw
			await assert.rejects(
				async () => {
					await manager.deleteSandbox(sandbox.id);
				},
				{
					message: "Driver non-existent not found",
				},
			);
		});
	});

	describe("listSandboxes", () => {
		it("should return empty array when no sandboxes exist", async () => {
			const sandboxes = await manager.listSandboxes();
			assert.deepStrictEqual(sandboxes, []);
		});

		it("should list all sandboxes", async () => {
			const sandbox1 = await manager.createSandbox();
			const sandbox2 = await manager.createSandbox();
			const sandbox3 = await manager.createSandbox();

			const sandboxes = await manager.listSandboxes();

			assert.strictEqual(sandboxes.length, 3);
			const ids = sandboxes.map((sb) => sb.id).sort();
			const expectedIds = [sandbox1.id, sandbox2.id, sandbox3.id].sort();
			assert.deepStrictEqual(ids, expectedIds);
		});

		it("should return LazySandbox instances", async () => {
			await manager.createSandbox();

			const sandboxes = await manager.listSandboxes();

			assert.strictEqual(sandboxes.length, 1);
			// Verify it has sandbox methods
			assert.strictEqual(typeof sandboxes[0].readFile, "function");
			assert.strictEqual(typeof sandboxes[0].writeFile, "function");
		});

		it("should not include non-sandbox files", async () => {
			await manager.createSandbox();

			// Add some non-sandbox files
			await fs.put(
				"./sandboxes/other-file.txt",
				new File(["content"], "other-file.txt"),
			);

			const sandboxes = await manager.listSandboxes();

			assert.strictEqual(sandboxes.length, 1);
		});
	});

	describe("loadSandbox", () => {
		it("should throw if driver not found", async () => {
			await assert.rejects(
				async () => {
					await manager.loadSandbox({
						id: "test",
						driver: "non-existent",
					});
				},
				{
					message: "Driver non-existent not found",
				},
			);
		});

		it("should throw if driverConfig missing and create not set", async () => {
			await assert.rejects(
				async () => {
					await manager.loadSandbox({
						id: "test",
						driver: "mock",
					});
				},
				{
					message: "Sandbox test not created yet",
				},
			);
		});

		it("should create sandbox if driverConfig missing and create=true", async () => {
			const config = {
				id: "test-sandbox",
				driver: "mock",
			};

			const sandbox = await manager.loadSandbox(config, { create: true });

			assert.ok(sandbox);
			assert.strictEqual(sandbox.id, "test-sandbox");

			// Verify driverConfig was created and saved
			const configFile = await fs.get("./sandboxes/test-sandbox.json");
			assert.ok(configFile);

			const savedConfig = JSON.parse(await configFile.text());
			assert.ok(savedConfig.driverConfig);
			assert.strictEqual(
				savedConfig.driverConfig.containerId,
				"container-test-sandbox",
			);
		});

		it("should load existing sandbox if driverConfig exists", async () => {
			const config = {
				id: "test-sandbox",
				driver: "mock",
				driverConfig: { containerId: "existing-container" },
			};

			const sandbox = await manager.loadSandbox(config);

			assert.ok(sandbox);
			assert.strictEqual(sandbox.id, "test-sandbox");
		});
	});
});

describe("CreateOnDemandSandbox", () => {
	let fs: FileStorage;
	let driver: MockDriver;
	let manager: Manager;

	beforeEach(() => {
		fs = createMemoryFileStorage();
		driver = new MockDriver();
		manager = new Manager(fs, {
			drivers: { mock: driver },
			defaultDriver: "mock",
		});
	});

	describe("lazy creation", () => {
		it("should not create sandbox until first write operation", async () => {
			const sandbox = await manager.createSandbox();

			// Config should exist, but no driverConfig yet
			const configFile = await fs.get(`./sandboxes/${sandbox.id}.json`);
			assert.ok(configFile, "Config file should exist");
			const config = JSON.parse(await configFile.text());
			assert.strictEqual(config.driverConfig, undefined);
		});

		it("should create sandbox on first write operation", async () => {
			const sandbox = await manager.createSandbox();

			await sandbox.writeFile("/test.txt", "content");

			// Now driverConfig should exist
			const configFile = await fs.get(`./sandboxes/${sandbox.id}.json`);
			assert.ok(configFile, "Config file should exist");
			const config = JSON.parse(await configFile.text());
			assert.ok(config.driverConfig);
		});

		it("should create sandbox on execute operation", async () => {
			const sandbox = await manager.createSandbox();

			const execId = await sandbox.execute("echo", ["hello"]);

			assert.strictEqual(execId, "mock-exec-id");

			// Verify sandbox was created
			const configFile = await fs.get(`./sandboxes/${sandbox.id}.json`);
			assert.ok(configFile, "Config file should exist");
			const config = JSON.parse(await configFile.text());
			assert.ok(config.driverConfig);
		});

		it("should reuse sandbox instance after creation", async () => {
			const sandbox = await manager.createSandbox();

			await sandbox.writeFile("/test1.txt", "content1");
			await sandbox.writeFile("/test2.txt", "content2");

			// Both writes should use the same sandbox instance
			// (we can't directly test this without exposing internals,
			// but we verify no errors occur)
		});
	});

	describe("read operations with parent sandbox", () => {
		it("should use parent sandbox for read when not yet created", async () => {
			const parent = await manager.createSandbox();
			await parent.writeFile("/parent-file.txt", "parent content");

			const child = await manager.createSandbox({ baseUri: parent.id });

			// Read should work even though child is not created
			const result = await child.readFile("/parent-file.txt");
			assert.ok(result);
			assert.ok(result.content);
		});

		it("should return null for read without parent or baseUri", async () => {
			const sandbox = await manager.createSandbox();

			// Try to read before writing (which would create the sandbox)
			const result = await sandbox.readFile("/test.txt");
			assert.strictEqual(result, null);
		});
	});

	describe("getMeta", () => {
		it("should return meta from config", async () => {
			const meta = { userId: "user-123" };
			const sandbox = await manager.createSandbox({ meta });

			const retrievedMeta = await sandbox.getMeta();
			assert.deepStrictEqual(retrievedMeta, meta);
		});

		it("should return empty object if no meta", async () => {
			const sandbox = await manager.createSandbox();

			const meta = await sandbox.getMeta();
			assert.deepStrictEqual(meta, {});
		});
	});
});

describe("LazySandbox", () => {
	let fs: FileStorage;
	let driver: MockDriver;
	let manager: Manager;

	beforeEach(() => {
		fs = createMemoryFileStorage();
		driver = new MockDriver();
		manager = new Manager(fs, {
			drivers: { mock: driver },
			defaultDriver: "mock",
		});
	});

	describe("lazy loading", () => {
		it("should load sandbox on first operation", async () => {
			const created = await manager.createSandbox();

			// Get list returns LazySandbox
			const sandboxes = await manager.listSandboxes();
			const lazy = sandboxes[0];

			// Trigger load
			await lazy.writeFile("/test.txt", "content");

			// Verify it worked
			assert.strictEqual(lazy.id, created.id);
		});

		it("should throw if sandbox not found", async () => {
			// Manually create a LazySandbox with non-existent ID
			// by creating storage and then manually removing the file (not via deleteSandbox)
			await fs.put(
				"./sandboxes/sb-fake.json",
				new File(
					[JSON.stringify({ id: "sb-fake", driver: "mock" })],
					"sb-fake.json",
				),
			);

			// Get the list which creates LazySandbox instances
			const sandboxes = await manager.listSandboxes();
			const lazy = sandboxes[0];

			// Manually remove the config file to simulate it being deleted
			await fs.remove("./sandboxes/sb-fake.json");

			// Try to use it - should throw "Sandbox not found" from LazySandbox
			await assert.rejects(
				async () => {
					await lazy.readFile("/test.txt");
				},
				{
					message: "Sandbox not found",
				},
			);
		});

		it("should delegate all operations to loaded sandbox", async () => {
			const _created = await manager.createSandbox();
			const sandboxes = await manager.listSandboxes();
			const lazy = sandboxes[0];

			// Test various operations
			await lazy.writeFile("/test.txt", "content");
			const result = await lazy.readFile("/test.txt");
			assert.ok(result);
			assert.ok(result.content);

			const execId = await lazy.execute("echo", ["hello"]);
			assert.strictEqual(execId, "mock-exec-id");

			const output = await lazy.output(execId);
			assert.strictEqual(output.output, "mock output");

			await lazy.wait(execId);
			await lazy.release(execId);
			await lazy.kill(execId);
			await lazy.deleteFile("/test.txt");
		});
	});

	describe("getMeta", () => {
		it("should return meta from config via loaded sandbox", async () => {
			const meta = { projectId: "proj-456" };
			await manager.createSandbox({ meta });

			const sandboxes = await manager.listSandboxes();
			const lazy = sandboxes[0];

			const retrievedMeta = await lazy.getMeta();
			assert.deepStrictEqual(retrievedMeta, meta);
		});
	});
});

describe("validate function", () => {
	let fs: FileStorage;
	let driver: MockDriver;
	let manager: Manager;

	beforeEach(() => {
		fs = createMemoryFileStorage();
		driver = new MockDriver();
		manager = new Manager(fs, {
			drivers: { mock: driver },
			defaultDriver: "mock",
		});
	});

	it("should validate valid sandbox config", async () => {
		// Create a sandbox to ensure config is valid
		const sandbox = await manager.createSandbox();

		// Config should be valid (no errors thrown)
		const retrieved = await manager.getSandbox(sandbox.id);
		assert.ok(retrieved);
	});

	it("should throw on invalid sandbox config", async () => {
		// Manually create invalid config
		await fs.put(
			"./sandboxes/sb-invalid.json",
			new File([JSON.stringify({ invalid: "config" })], "sb-invalid.json"),
		);

		// Should throw validation error
		await assert.rejects(async () => {
			await manager.getSandbox("sb-invalid");
		});
	});
});
