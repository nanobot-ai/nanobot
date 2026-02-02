import type { ChatMessageItemToolCall } from './types';

export function parseToolFilePath(item: ChatMessageItemToolCall) {
	if (!item.arguments) return null;
	const parsed = JSON.parse(item.arguments);
	const filePath = parsed.file_path;
	return filePath ? filePath.split('/').pop().split('.').shift() : null;
}
