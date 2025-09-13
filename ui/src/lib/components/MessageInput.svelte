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
	let textareaRef: HTMLTextAreaElement;
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

	function autoResize() {
		if (!textareaRef) return;

		// Reset height to auto to get the correct scrollHeight
		textareaRef.style.height = 'auto';

		// Set the height based on scrollHeight, but respect min and max constraints
		const newHeight = Math.min(Math.max(textareaRef.scrollHeight, 40), 128); // min 40px (2.5rem), max 128px (8rem)
		textareaRef.style.height = `${newHeight}px`;
	}

	// Auto-resize when message changes
	$effect(() => {
		if (textareaRef) {
			autoResize();
		}
	});
</script>

<div class="p-4">
	<!-- Uploaded files display -->
	{#if uploadedFiles.length > 0}
		<div class="mb-3 flex flex-wrap gap-2">
			{#each uploadedFiles as uploadedFile (uploadedFile.id)}
				<div class="flex items-center gap-2 rounded-xl bg-base-200 px-3 py-2 text-sm">
					<span>{getFileIcon(uploadedFile.file)}</span>
					<span class="max-w-32 truncate">{uploadedFile.file.name}</span>
					<button
						type="button"
						onclick={() => removeFile(uploadedFile.id)}
						class="btn h-5 w-5 rounded-full p-0 btn-ghost btn-xs"
						aria-label="Remove file"
					>
						<X class="h-3 w-3" />
					</button>
				</div>
			{/each}
		</div>
	{/if}

	<!-- Hidden file input -->
	<input
		bind:this={fileInput}
		type="file"
		accept={supportedMimeTypes.join(',')}
		onchange={handleFileSelect}
		class="hidden"
		aria-label="File upload"
	/>

	<form onsubmit={handleSubmit}>
		<div
			class="space-y-3 rounded-2xl border-2 border-base-200 bg-base-100 p-3 transition-colors focus-within:border-primary"
		>
			<!-- Top row: Full-width input -->
			<textarea
				bind:value={message}
				onkeydown={handleKeydown}
				oninput={autoResize}
				{placeholder}
				disabled={disabled || isUploading}
				class="max-h-32 min-h-[2.5rem] w-full resize-none bg-transparent p-1 text-sm leading-6 outline-none placeholder:text-base-content/50"
				rows="1"
				bind:this={textareaRef}
			></textarea>

			<!-- Bottom row: Model select on left, buttons on right -->
			<div class="flex items-center justify-end">
				<!-- Model selector -->
				<select
					class="select hidden w-48 select-ghost select-sm"
					disabled={disabled || isUploading}
				>
					<option value="gpt-4">GPT-4</option>
					<option value="gpt-3.5-turbo">GPT-3.5 Turbo</option>
					<option value="claude-3-opus">Claude 3 Opus</option>
					<option value="claude-3-sonnet">Claude 3 Sonnet</option>
					<option value="gemini-pro">Gemini Pro</option>
				</select>

				<!-- Action buttons -->
				<div class="flex gap-2">
					<!-- Attach button -->
					<button
						type="button"
						onclick={handleAttach}
						class="btn h-9 w-9 rounded-full p-0 btn-ghost btn-sm"
						disabled={disabled || isUploading}
						aria-label="Attach file"
					>
						{#if isUploading}
							<span class="loading loading-xs loading-spinner"></span>
						{:else}
							<Paperclip class="h-4 w-4" />
						{/if}
					</button>

					<!-- Submit button -->
					<button
						type="submit"
						class="btn h-9 w-9 rounded-full p-0 btn-sm btn-primary"
						disabled={disabled || isUploading || !message.trim()}
						aria-label="Send message"
					>
						<Send class="h-4 w-4" />
					</button>
				</div>
			</div>
		</div>
	</form>
</div>
