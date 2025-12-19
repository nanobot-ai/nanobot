
// this file is generated — do not edit it


declare module "svelte/elements" {
	export interface HTMLAttributes<T> {
		'data-sveltekit-keepfocus'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-noscroll'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-preload-code'?:
			| true
			| ''
			| 'eager'
			| 'viewport'
			| 'hover'
			| 'tap'
			| 'off'
			| undefined
			| null;
		'data-sveltekit-preload-data'?: true | '' | 'hover' | 'tap' | 'off' | undefined | null;
		'data-sveltekit-reload'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-replacestate'?: true | '' | 'off' | undefined | null;
	}
}

export {};


declare module "$app/types" {
	export interface AppTypes {
		RouteId(): "/" | "/c" | "/c/[id]" | "/w" | "/w/[workspaceId]" | "/w/[workspaceId]/t" | "/w/[workspaceId]/t/[taskId]";
		RouteParams(): {
			"/c/[id]": { id: string };
			"/w/[workspaceId]": { workspaceId: string };
			"/w/[workspaceId]/t": { workspaceId: string };
			"/w/[workspaceId]/t/[taskId]": { workspaceId: string; taskId: string }
		};
		LayoutParams(): {
			"/": { id?: string; workspaceId?: string; taskId?: string };
			"/c": { id?: string };
			"/c/[id]": { id: string };
			"/w": { workspaceId?: string; taskId?: string };
			"/w/[workspaceId]": { workspaceId: string; taskId?: string };
			"/w/[workspaceId]/t": { workspaceId: string; taskId?: string };
			"/w/[workspaceId]/t/[taskId]": { workspaceId: string; taskId: string }
		};
		Pathname(): "/" | "/c" | "/c/" | `/c/${string}` & {} | `/c/${string}/` & {} | "/w" | "/w/" | `/w/${string}` & {} | `/w/${string}/` & {} | `/w/${string}/t` & {} | `/w/${string}/t/` & {} | `/w/${string}/t/${string}` & {} | `/w/${string}/t/${string}/` & {};
		ResolvedPathname(): `${"" | `/${string}`}${ReturnType<AppTypes['Pathname']>}`;
		Asset(): "/robots.txt" | string & {};
	}
}