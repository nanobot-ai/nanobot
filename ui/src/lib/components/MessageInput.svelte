<script lang="ts">
	import { Paperclip, X, Send } from '@lucide/svelte';

	interface Props {
		onSend?: (message: string) => void;
		onFileUpload?: (file: File, url: string) => void;
		placeholder?: string;
		disabled?: boolean;
		supportedMimeTypes?: string[];
		uploadUrl?: string;
	}

	let {
		onSend,
		onFileUpload,
		placeholder = 'Type a message...',
		disabled = false,
		supportedMimeTypes = [
			'image/*',
			'text/plain',
			'application/pdf',
			'application/json',
			'text/csv',
			'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
			'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
		],
		uploadUrl = '/api/upload'
	}: Props = $props();

	let message = $state('');
	let fileInput: HTMLInputElement;
	let isUploading = $state(false);
	let uploadedFiles = $state<Array<{ file: File; url: string; id: string }>>([]);

	function handleSubmit(e: Event) {
		e.preventDefault();
		if (message.trim() && onSend) {
			onSend(message.trim());
			message = '';
			// Clear uploaded files after sending
			uploadedFiles = [];
		}
	}

	function handleAttach() {
		fileInput?.click();
	}

	async function handleFileSelect(e: Event) {
		const target = e.target as HTMLInputElement;
		const file = target.files?.[0];

		if (!file) return;

		isUploading = true;

		try {
			const formData = new FormData();
			formData.append('file', file);

			const response = await fetch(uploadUrl, {
				method: 'POST',
				body: formData
			});

			if (response.ok) {
				const result = await response.json();
				const fileId = crypto.randomUUID();
				const fileUrl = result.url || uploadUrl;

				// Add to uploaded files list
				uploadedFiles.push({
					file,
					url: fileUrl,
					id: fileId
				});

				if (onFileUpload) {
					onFileUpload(file, fileUrl);
				}
				console.log('File uploaded successfully:', result);
			} else {
				console.error('Upload failed:', response.statusText);
			}
		} catch (error) {
			console.error('Upload error:', error);
		} finally {
			isUploading = false;
			target.value = '';
		}
	}

	function removeFile(fileId: string) {
		uploadedFiles = uploadedFiles.filter((f) => f.id !== fileId);
	}

	function getFileIcon(file: File) {
		if (file.type.startsWith('image/')) {
			return '🖼️';
		} else if (file.type === 'application/pdf') {
			return '📄';
		} else if (file.type.includes('text/') || file.type.includes('json')) {
			return '📝';
		} else if (file.type.includes('spreadsheet') || file.type.includes('csv')) {
			return '📊';
		} else if (file.type.includes('document')) {
			return '📋';
		}
		return '📎';
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			handleSubmit(e);
		}
	}
</script>

<form onsubmit={handleSubmit} class="flex items-end gap-2 p-4">
	<!-- Hidden file input -->
	<input
		bind:this={fileInput}
		type="file"
		accept={supportedMimeTypes.join(',')}
		onchange={handleFileSelect}
		class="hidden"
		aria-label="File upload"
	/>

	<button
		type="button"
		onclick={handleAttach}
		class="btn btn-ghost btn-sm"
		disabled={disabled || isUploading}
		aria-label="Attach file"
	>
		{#if isUploading}
			<span class="loading loading-xs loading-spinner"></span>
		{:else}
			<Paperclip class="h-5 w-5" />
		{/if}
	</button>

	<div class="flex-1">
		<!-- Uploaded files display -->
		{#if uploadedFiles.length > 0}
			<div class="mb-2 flex flex-wrap gap-1">
				{#each uploadedFiles as uploadedFile (uploadedFile.id)}
					<div class="flex items-center gap-1 rounded-full bg-base-200 px-2 py-1 text-xs">
						<span>{getFileIcon(uploadedFile.file)}</span>
						<span class="max-w-20 truncate">{uploadedFile.file.name}</span>
						<button
							type="button"
							onclick={() => removeFile(uploadedFile.id)}
							class="btn h-4 min-h-4 w-4 rounded-full p-0 btn-ghost btn-xs"
							aria-label="Remove file"
						>
							<X class="h-3 w-3" />
						</button>
					</div>
				{/each}
			</div>
		{/if}

		<textarea
			bind:value={message}
			onkeydown={handleKeydown}
			{placeholder}
			disabled={disabled || isUploading}
			class="textarea-bordered textarea max-h-32 min-h-[2.5rem] w-full resize-none"
			rows="1"
		></textarea>
	</div>

	<button
		type="submit"
		class="btn btn-sm btn-primary"
		disabled={disabled || isUploading || !message.trim()}
		aria-label="Send message"
	>
		<Send class="h-5 w-5" />
	</button>
</form>
