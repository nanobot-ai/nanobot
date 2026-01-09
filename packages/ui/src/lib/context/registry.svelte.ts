import { getContext, setContext } from 'svelte';
import { env } from '$env/dynamic/public';
import type { Server } from '$lib/types';

const REGISTRY_CONTEXT_KEY = Symbol('registry');

type RegistryResponse = {
	metadata: {
		count: number;
		nextCursor?: string | null;
	};
	servers: {
		server: Server;
	}[];
};

export interface RegistryStore {
	servers: Server[];
	loading: boolean;
	error: string | null;
	getServerByName(name: string): Server | undefined;
	fetch: () => Promise<void>;
}

const CACHE_TTL_MS = 60 * 1000; // 1 minute

export function createRegistryStore(): RegistryStore {
	let servers = $state<Server[]>([]);
	let nextCursor = $state<string | null>(null);
	let loading = $state(false);
	let error = $state<string | null>(null);
	let lastFetchTime: number | null = null;

	const serversMap = $derived(
		servers.reduce<Record<string, Server>>((map, server) => {
			map[server.name] = server;
			return map;
		}, {})
	);

	async function fetch() {
		if (loading) return;
		if (lastFetchTime && Date.now() - lastFetchTime < CACHE_TTL_MS) {
			return;
		}

		loading = true;
		error = null;

		if (!env.PUBLIC_REGISTRY_ENDPOINT) {
			throw new Error('Registry endpoint is not set');
		}

		try {
			const response = await globalThis.fetch(`${env.PUBLIC_REGISTRY_ENDPOINT}?limit=100`);
			const data: RegistryResponse = await response.json();
			nextCursor = data.metadata.nextCursor ?? null;
			servers = data.servers.map((server) => server.server);

			// to fetch all but TODO: change to pagination
			while (nextCursor) {
				const response = await globalThis.fetch(
					`${env.PUBLIC_REGISTRY_ENDPOINT}?cursor=${nextCursor}`
				);
				const data: RegistryResponse = await response.json();
				nextCursor = data.metadata.nextCursor ?? null;
				servers = [...servers, ...data.servers.map((server) => server.server)];
			}

			lastFetchTime = Date.now();
		} catch (e) {
			error = e instanceof Error ? e.message : String(e);
			throw e;
		} finally {
			loading = false;
		}
	}

	function getServerByName(name: string): Server | undefined {
		return serversMap[name];
	}

	return {
		get servers() {
			return servers;
		},
		get loading() {
			return loading;
		},
		get error() {
			return error;
		},
		getServerByName,
		fetch
	};
}

export function setRegistryContext(store: RegistryStore) {
	setContext(REGISTRY_CONTEXT_KEY, store);
}

export function getRegistryContext(): RegistryStore {
	return getContext<RegistryStore>(REGISTRY_CONTEXT_KEY);
}
