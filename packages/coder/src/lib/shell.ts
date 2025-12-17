import type { WorkspaceClient } from "@nanobot-ai/workspace-client";

/**
 * Manages background shell sessions
 */
const backgroundShells = new Map<
	string,
	{ terminalId: string; command: string }
>();

export function registerBackgroundShell(
	shellId: string,
	terminalId: string,
	command: string,
): void {
	backgroundShells.set(shellId, { terminalId, command });
}

export function getBackgroundShell(
	shellId: string,
): { terminalId: string; command: string } | undefined {
	return backgroundShells.get(shellId);
}

export function removeBackgroundShell(shellId: string): void {
	backgroundShells.delete(shellId);
}

/**
 * Execute a bash command and wait for it to complete
 */
export async function executeBashCommand(
	client: WorkspaceClient,
	command: string,
	options?: {
		timeout?: number;
		cwd?: string;
		env?: Record<string, string>;
	},
): Promise<{ output: string; exitCode: number; timedOut: boolean }> {
	// Create terminal with the command wrapped in sh -c
	const { terminalId } = await client.terminalCreate("bash", {
		args: ["-c", command],
		cwd: options?.cwd,
		env: options?.env,
	});

	try {
		// Wait for completion with timeout
		const timeout = options?.timeout || 120000; // Default 2 minutes
		const timeoutPromise = new Promise<null>((resolve) =>
			setTimeout(() => resolve(null), timeout),
		);

		const waitPromise = client.terminalWait(terminalId);
		const result = await Promise.race([waitPromise, timeoutPromise]);

		if (result === null) {
			// Timeout occurred
			await client.terminalKill(terminalId);
			const { output } = await client.terminalOutput(terminalId);
			await client.terminalRelease(terminalId);
			return {
				output: `${output}\n\nCommand timed out after ${timeout}ms`,
				exitCode: 124,
				timedOut: true,
			};
		}

		// Get output
		const { output } = await client.terminalOutput(terminalId);
		await client.terminalRelease(terminalId);

		return {
			output,
			exitCode: result.exitCode,
			timedOut: false,
		};
	} catch (error) {
		// Clean up on error
		try {
			await client.terminalRelease(terminalId);
		} catch {
			// Ignore cleanup errors
		}
		throw error;
	}
}

/**
 * Generate a unique shell ID for background shells
 */
export function generateShellId(): string {
	return `shell-${Date.now()}-${Math.random().toString(36).slice(2, 11)}`;
}
