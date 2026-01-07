import { ChatService } from '$lib/chat.svelte';

// Shared chat instance that persists across page navigations
// Using an object wrapper so we can export a reactive reference
const chatState = $state<{ current: ChatService | null }>({ current: null });

// Export the reactive state object - access via sharedChat.current
export const sharedChat = chatState;

export function setSharedChat(chat: ChatService): void {
	chatState.current = chat;
}

export function clearSharedChat(): void {
	chatState.current = null;
}
