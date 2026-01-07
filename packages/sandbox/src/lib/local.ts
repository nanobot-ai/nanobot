/**
 * Local filesystem sandbox implementation
 *
 * This module provides a SandboxDriver implementation that uses the local filesystem
 * directly without any containerization. It's designed to be used within Docker containers
 * that already have full filesystem access.
 *
 * @example
 * ```typescript
 * import { Manager } from "./manager.js";
 * import { LocalDriver } from "./local.js";
 * import { createMemoryFileStorage } from "@remix-run/file-storage/memory";
 *
 * // Create a manager with the Local driver
 * const driver = new LocalDriver({
 *   workdir: "/workspace",
 * });
 *
 * const manager = new Manager(createMemoryFileStorage(), {
 *   drivers: { local: driver },
 *   defaultDriver: "local",
 * });
 *
 * // Create and use a sandbox
 * const sandbox = await manager.createSandbox();
 * await sandbox.writeFile("hello.txt", "Hello, World!");
 * const content = await sandbox.readFile("hello.txt");
 * console.log(content); // { content: "Hello, World!", encoding: "utf-8" }
 *
 * // Execute commands
 * const procId = await sandbox.execute("echo", ["Hello from local!"]);
 * await sandbox.wait(procId);
 * const output = await sandbox.output(procId);
 * console.log(output.output); // "Hello from local!\n"
 *
 * // Cleanup
 * await manager.deleteSandbox(sandbox.id);
 * ```
 */

import { spawn } from "node:child_process";
import { existsSync } from "node:fs";
import { cp, mkdir, rm } from "node:fs/promises";
import { join, dirname } from "node:path";
import { fileURLToPath } from "node:url";
import type { Encoding, Sandbox, SandboxMetadata } from "../sandbox.js";
import type { SandboxConfig, SandboxDriver } from "./manager.js";

export interface LocalDriverConfig {
  workdir?: string;
  dataDir?: string;
  env?: Record<string, string>;
  helperScriptPath?: string;
}

interface LocalSandboxConfig {
  workdir: string;
  dataDir: string;
  helperScriptPath: string;
}

/**
 * Process state is persisted to disk in <dataDir>/.processes/<processId>/
 * Each process directory contains:
 * - meta.json: { command, args, startTime, exitCode, signal, outputByteLimit, truncated }
 * - stdout.txt: Process stdout output
 * - stderr.txt: Process stderr output
 * - exitcode.txt: Process exit code
 * - signal.txt: Signal that terminated process (if killed)
 * - pid.txt: Process PID for kill operations
 *
 * The sb_execute script manages all process state and output.
 */
interface ProcessMetadata {
  command: string;
  args: string[];
  startTime: number;
  exitCode: number | null;
  signal?: string | null;
  outputByteLimit?: number;
  truncated?: boolean;
}

export class LocalDriver implements SandboxDriver {
  readonly name = "local";
  private config: Required<LocalDriverConfig>;

  constructor(config?: LocalDriverConfig) {
    // Get the directory of this module
    const moduleDir = dirname(fileURLToPath(import.meta.url));

    this.config = {
      workdir: config?.workdir ?? process.cwd(),
      dataDir: config?.dataDir ?? join(process.cwd(), "local-sandbox"),
      env: config?.env ?? {},
      helperScriptPath: config?.helperScriptPath ?? moduleDir,
    };
  }

  async createSandbox(
    config: SandboxConfig,
    opts?: { recreate?: boolean },
  ): Promise<Record<string, unknown>> {
    // Create local directory for this sandbox
    const dataDir = join(this.config.dataDir, config.id);

    // If parentId is set, copy parent's data directory
    if (config.parentId) {
      const parentDataDir = join(this.config.dataDir, config.parentId);

      if (existsSync(parentDataDir)) {
        if (!opts?.recreate) {
          // Copy parent directory to new sandbox directory
          await cp(parentDataDir, dataDir, { recursive: true });
          console.log(`Copied parent data from ${parentDataDir} to ${dataDir}`);
        }
      } else {
        // Parent directory doesn't exist, create empty one
        await mkdir(dataDir, { recursive: true });
        console.warn(
          `Parent directory ${parentDataDir} not found, created empty directory`,
        );
      }
    } else {
      // No parent, create empty directory
      await mkdir(dataDir, { recursive: true });
    }

    return {
      workdir: this.config.workdir,
      dataDir,
      helperScriptPath: this.config.helperScriptPath,
    };
  }

  async deleteSandbox(config: SandboxConfig): Promise<void> {
    const driverConfig = config.driverConfig as LocalSandboxConfig | undefined;

    // Clean up local data directory
    if (driverConfig?.dataDir) {
      try {
        await rm(driverConfig.dataDir, { recursive: true, force: true });
      } catch (_error) {
        // Directory might not exist or already be removed, ignore
      }
    }
  }

  async loadSandbox(
    config: SandboxConfig,
    _opts?: { create?: boolean },
  ): Promise<Sandbox> {
    const driverConfig = config.driverConfig as LocalSandboxConfig | undefined;
    if (!driverConfig) {
      throw new Error(`Sandbox ${config.id} not created yet`);
    }

    return new LocalSandbox(config.id, driverConfig, config.meta ?? {});
  }
}

class LocalSandbox implements Sandbox {
  readonly id: string;
  private readonly workdir: string;
  private readonly dataDir: string;
  private readonly meta: SandboxMetadata;
  private readonly helperScriptPath: string;

  constructor(id: string, config: LocalSandboxConfig, meta: SandboxMetadata) {
    this.id = id;
    this.workdir = config.workdir;
    this.dataDir = config.dataDir;
    this.meta = meta;
    this.helperScriptPath = config.helperScriptPath;
  }

  async getMeta(): Promise<SandboxMetadata> {
    return this.meta;
  }

  resolvePath(path: string): string {
    return this.#_resolvePath(path);
  }

  #_resolvePath(path: string): string {
    if (path.startsWith("/")) {
      return path;
    }
    return `${this.workdir}/${path}`;
  }

  private async getProcessMetadata(
    processId: string,
  ): Promise<ProcessMetadata | null> {
    const processDir = `${this.dataDir}/.processes/${processId}`;
    try {
      const metaJson = await this.execHelper([
        join(this.helperScriptPath, "sb_read"),
        `${processDir}/meta.json`,
        "utf-8",
      ]);
      return JSON.parse(metaJson) as ProcessMetadata;
    } catch (_error) {
      return null;
    }
  }

  private async getProcessExitCode(processId: string): Promise<number | null> {
    const processDir = `${this.dataDir}/.processes/${processId}`;
    try {
      const exitCodeStr = await this.execHelper([
        join(this.helperScriptPath, "sb_read"),
        `${processDir}/exitcode.txt`,
        "utf-8",
      ]);
      return Number.parseInt(exitCodeStr.trim(), 10);
    } catch (_error) {
      return null;
    }
  }

  async readFile(
    path: string,
    opts?: { encoding?: Encoding; limit?: number; offset?: number },
  ): Promise<{
    content: string;
    encoding: Encoding;
  } | null> {
    const encoding = opts?.encoding ?? "utf-8";
    const absolutePath = this.#_resolvePath(path);

    try {
      const args = [
        join(this.helperScriptPath, "sb_read"),
        absolutePath,
        encoding,
      ];

      if (opts?.offset) {
        args.push(String(opts.offset));
      } else {
        args.push("1");
      }

      if (opts?.limit) {
        args.push(String(opts.limit));
      }

      const content = await this.execHelper(args);

      return {
        content: encoding === "base64" ? content.trim() : content,
        encoding,
      };
    } catch (error) {
      // Check if the error is a "file not found" error
      const errorMessage =
        error instanceof Error ? error.message : String(error);
      if (
        errorMessage.includes("No such file") ||
        errorMessage.includes("not found")
      ) {
        return null;
      }
      // For other errors, still throw
      throw new Error(`Failed to read file ${path}: ${errorMessage}`);
    }
  }

  async writeFile(
    path: string,
    content: string,
    opts?: { encoding?: Encoding },
  ): Promise<void> {
    const encoding = opts?.encoding ?? "utf-8";
    const absolutePath = this.#_resolvePath(path);

    try {
      // sb_write will create directories as needed
      await this.execHelper(
        [join(this.helperScriptPath, "sb_write"), absolutePath, encoding],
        content,
      );
    } catch (error) {
      throw new Error(
        `Failed to write file ${path}: ${error instanceof Error ? error.message : String(error)}`,
      );
    }
  }

  async deleteFile(path: string): Promise<void> {
    const absolutePath = this.#_resolvePath(path);

    try {
      await this.execHelper(["rm", "-f", absolutePath]);
    } catch (error) {
      throw new Error(
        `Failed to delete file ${path}: ${error instanceof Error ? error.message : String(error)}`,
      );
    }
  }

  async readdir(
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
      skipped?: boolean;
    }>;
    cursor?: string;
  }> {
    const absolutePath = this.#_resolvePath(path);

    try {
      // Calculate offset from cursor
      const offset = opts?.cursor ? opts.cursor : "0";
      const limit = opts?.limit ? String(opts.limit) : "";
      const recursive = opts?.recursive ? "1" : "0";

      // Call sb_readdir script
      const args = [
        join(this.helperScriptPath, "sb_readdir"),
        absolutePath,
        offset,
        limit,
        recursive,
      ];
      const output = await this.execHelper(args);

      // Parse the output
      const lines = output
        .trim()
        .split("\n")
        .filter((line) => line.length > 0);

      // Check if we got more entries than the limit (indicates more pages)
      let hasMore = false;
      if (opts?.limit && lines.length > opts.limit) {
        hasMore = true;
        lines.pop(); // Remove the extra entry
      }

      const entries = lines.map((line) => {
        const [type, sizeStr, fullPath, skipped] = line.split("|");
        const size = Number.parseInt(sizeStr, 10) || 0;

        // Get relative path from the directory being listed
        let name = fullPath;
        if (fullPath.startsWith(`${absolutePath}/`)) {
          name = fullPath.substring(absolutePath.length + 1);
        } else if (fullPath === absolutePath) {
          name = ".";
        }

        return {
          name,
          isFile: type === "f",
          isDirectory: type === "d",
          size,
          skipped: skipped === "1" ? true : undefined,
        };
      });

      // Calculate next cursor if there are more entries
      const nextCursor = hasMore
        ? String(Number.parseInt(offset, 10) + entries.length)
        : undefined;

      return {
        entries,
        cursor: nextCursor,
      };
    } catch (error) {
      // Check if the error is a "not found" or "not a directory" error
      const errorMessage =
        error instanceof Error ? error.message : String(error);
      if (
        errorMessage.includes("not found") ||
        errorMessage.includes("not a directory")
      ) {
        return { entries: [], cursor: undefined };
      }
      throw new Error(`Failed to read directory ${path}: ${errorMessage}`);
    }
  }

  async execute(
    command: string,
    args: string[],
    opts?: {
      cwd?: string;
      env?: Record<string, string>;
      outputByteLimit?: number;
    },
  ): Promise<string> {
    // Generate unique process ID
    const processId = `proc-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;

    // Default output byte limit to 0 (unlimited)
    const outputByteLimit = opts?.outputByteLimit ?? 0;

    // Build sb_execute command with process ID and output byte limit
    const execArgs = [
      join(this.helperScriptPath, "sb_execute"),
      processId,
      String(outputByteLimit),
      command,
      ...args,
    ];

    // Set working directory and environment
    const execOpts: {
      cwd?: string;
      env?: Record<string, string>;
      detached?: boolean;
    } = {
      detached: true,
    };

    if (opts?.cwd) {
      execOpts.cwd = this.#_resolvePath(opts.cwd);
    }

    if (opts?.env) {
      execOpts.env = opts.env;
    }

    // Execute in detached mode
    await this.execHelper(execArgs, undefined, execOpts);

    return processId;
  }

  async kill(id: string, signal?: string): Promise<void> {
    const metadata = await this.getProcessMetadata(id);
    if (!metadata) {
      throw new Error(`Process ${id} not found`);
    }

    // Default to SIGTERM if no signal specified
    const killSignal = signal ?? "SIGTERM";

    // Read PID from pid.txt file
    const processDir = `${this.dataDir}/.processes/${id}`;
    let pid: string | null = null;
    try {
      pid = (
        await this.execHelper([
          join(this.helperScriptPath, "sb_read"),
          `${processDir}/pid.txt`,
          "utf-8",
        ])
      ).trim();
    } catch (_error) {
      // PID file might not exist
    }

    // Try to kill the process by PID if available, otherwise by command name
    try {
      if (pid) {
        await this.execHelper(["kill", `-${killSignal}`, pid]);
      } else {
        await this.execHelper([
          "pkill",
          `-${killSignal}`,
          "-f",
          metadata.command,
        ]);
      }

      // Update metadata with signal information
      await this.execHelper([
        "sh",
        "-c",
        `echo "${killSignal}" > "${processDir}/signal.txt"`,
      ]);

      // Update the metadata JSON with signal
      const updatedMeta = { ...metadata, signal: killSignal };
      await this.execHelper(
        [
          join(this.helperScriptPath, "sb_write"),
          `${processDir}/meta.json`,
          "utf-8",
        ],
        JSON.stringify(updatedMeta, null, 2),
      );
    } catch (_error) {
      // Process might already be done or not found
    }
  }

  async output(id: string): Promise<{
    output: string;
    truncated: boolean;
    exitCode: number;
    signal?: string;
  }> {
    const metadata = await this.getProcessMetadata(id);
    if (!metadata) {
      throw new Error(`Process ${id} not found`);
    }

    const processDir = `${this.dataDir}/.processes/${id}`;

    // Read stdout and stderr
    let stdout = "";
    let stderr = "";

    try {
      stdout = await this.execHelper([
        join(this.helperScriptPath, "sb_read"),
        `${processDir}/stdout.txt`,
        "utf-8",
      ]);
    } catch (_error) {
      // File might not exist yet
    }

    try {
      stderr = await this.execHelper([
        join(this.helperScriptPath, "sb_read"),
        `${processDir}/stderr.txt`,
        "utf-8",
      ]);
    } catch (_error) {
      // File might not exist yet
    }

    // Get exit code from file or metadata
    let exitCode = await this.getProcessExitCode(id);
    if (exitCode === null) {
      exitCode = metadata.exitCode ?? -1;
    }

    // Get signal from metadata (if process was killed)
    const signal = metadata.signal ?? undefined;

    // Get truncated flag from metadata
    const truncated = metadata.truncated ?? false;

    return {
      output: stdout + stderr,
      truncated,
      exitCode,
      signal,
    };
  }

  async wait(id: string): Promise<{ exitCode: number; signal?: string }> {
    const metadata = await this.getProcessMetadata(id);
    if (!metadata) {
      throw new Error(`Process ${id} not found`);
    }

    // Poll for exit code file to exist (sb_execute writes this when done)
    const maxWait = 3600000; // 1 hour
    const pollInterval = 100; // 100ms
    const startTime = Date.now();

    while (Date.now() - startTime < maxWait) {
      const exitCode = await this.getProcessExitCode(id);
      if (exitCode !== null) {
        // Refresh metadata to get signal information
        const updatedMetadata = await this.getProcessMetadata(id);
        const signal = updatedMetadata?.signal ?? undefined;
        return { exitCode, signal };
      }
      await new Promise((resolve) => setTimeout(resolve, pollInterval));
    }

    throw new Error(`Process ${id} timed out after ${maxWait}ms`);
  }

  async release(id: string): Promise<void> {
    const metadata = await this.getProcessMetadata(id);
    if (!metadata) {
      throw new Error(`Process ${id} not found`);
    }

    // Remove the process directory
    const processDir = `${this.dataDir}/.processes/${id}`;
    try {
      await this.execHelper(["rm", "-rf", processDir]);
    } catch (_error) {
      // Directory might not exist or already removed
    }
  }

  private async execHelper(
    args: string[],
    stdin?: string,
    opts?: { cwd?: string; env?: Record<string, string>; detached?: boolean },
  ): Promise<string> {
    return new Promise((resolve, reject) => {
      const execOpts: {
        cwd?: string;
        env?: Record<string, string | undefined>;
        detached?: boolean;
      } = {};

      // Set working directory if provided
      if (opts?.cwd) {
        execOpts.cwd = opts.cwd;
      }

      // Merge environment variables
      if (opts?.env) {
        execOpts.env = { ...process.env, ...opts.env };
      }

      // Set detached mode if requested
      if (opts?.detached) {
        execOpts.detached = true;
      }

      const proc = spawn(args[0], args.slice(1), execOpts);

      let stdout = "";
      let stderr = "";

      if (proc.stdout) {
        proc.stdout.on("data", (data) => {
          stdout += data.toString();
        });
      }

      if (proc.stderr) {
        proc.stderr.on("data", (data) => {
          stderr += data.toString();
        });
      }

      proc.on("close", (code) => {
        if (opts?.detached) {
          // For detached processes, don't check exit code
          resolve(stdout);
        } else if (code !== 0) {
          reject(
            new Error(
              `Command failed with exit code ${code}: ${stderr || stdout}`,
            ),
          );
        } else {
          resolve(stdout);
        }
      });

      proc.on("error", (error) => {
        reject(error);
      });

      // Write stdin if provided
      if (stdin !== undefined && proc.stdin) {
        proc.stdin.write(stdin);
        proc.stdin.end();
      }

      // For detached processes, unref so they don't keep the parent alive
      if (opts?.detached) {
        proc.unref();
      }
    });
  }
}
