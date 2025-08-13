<script lang="ts">
	import { slide } from 'svelte/transition';
	import { X, CheckCircle, AlertCircle, AlertTriangle, Info } from '@lucide/svelte';
	import { notifications } from '$lib/stores/notifications.svelte';
	import type { Notification } from '$lib/types';

	function getNotificationClasses(type: Notification['type']): string {
		const baseClasses = 'alert shadow-lg border';

		switch (type) {
			case 'success':
				return `${baseClasses} alert-success`;
			case 'error':
				return `${baseClasses} alert-error`;
			case 'warning':
				return `${baseClasses} alert-warning`;
			case 'info':
				return `${baseClasses} alert-info`;
			default:
				return `${baseClasses}`;
		}
	}

	function handleClose(id: string) {
		notifications.remove(id);
	}
</script>

<!-- Notification container -->
<div class="fixed right-4 bottom-4 z-50 flex w-80 flex-col gap-3">
	{#each notifications.notifications as notification (notification.id)}
		<div
			class={getNotificationClasses(notification.type)}
			in:slide={{ duration: 300 }}
			out:slide={{ duration: 200 }}
		>
			<div class="flex items-start gap-3">
				<!-- Notification icon -->
				<div class="flex-shrink-0">
					{#if notification.type === 'success'}
						<CheckCircle class="h-5 w-5" />
					{:else if notification.type === 'error'}
						<AlertCircle class="h-5 w-5" />
					{:else if notification.type === 'warning'}
						<AlertTriangle class="h-5 w-5" />
					{:else}
						<Info class="h-5 w-5" />
					{/if}
				</div>

				<!-- Notification content -->
				<div class="min-w-0 flex-1">
					<div class="text-sm font-medium">{notification.title}</div>
					{#if notification.message}
						<div class="mt-1 text-xs opacity-80">{notification.message}</div>
					{/if}
				</div>

				<!-- Close button -->
				<button
					class="btn btn-square flex-shrink-0 btn-ghost btn-xs"
					onclick={() => handleClose(notification.id)}
					aria-label="Close notification"
				>
					<X class="h-4 w-4" />
				</button>
			</div>

			<!-- Progress bar for auto-dismiss notifications -->
			{#if notification.autoClose && notification.duration && notification.duration > 0}
				<div class="mt-2 h-1 overflow-hidden rounded bg-black/10">
					<div
						class="h-full animate-pulse bg-current opacity-60"
						style="animation: shrink {notification.duration}ms linear forwards;"
					></div>
				</div>
			{/if}
		</div>
	{/each}
</div>

<style>
	@keyframes shrink {
		from {
			width: 100%;
		}
		to {
			width: 0%;
		}
	}
</style>
