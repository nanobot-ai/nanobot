<script lang="ts">
	import { onMount } from 'svelte';
	import { writable } from 'svelte/store';

	let activeFile = writable('SOUL.md');
	let fileContent = writable('');
	let files = ['SOUL.md', 'USER.md', 'MEMORY.md'];

	async function fetchFileContent(fileName: string) {
		try {
			const response = await fetch(`/api/memory/${fileName}`);
			if (response.ok) {
				const data = await response.json();
				fileContent.set(data.content);
			} else {
				console.error('Failed to fetch file content');
				fileContent.set('Error loading file.');
			}
		} catch (error) {
			console.error('Error fetching file:', error);
			fileContent.set('Error loading file.');
		}
	}

	async function saveFileContent() {
		try {
			const response = await fetch(`/api/memory/${$activeFile}`, {
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ content: $fileContent })
			});
			if (!response.ok) {
				console.error('Failed to save file');
				alert('Failed to save file.');
			} else {
				alert('File saved successfully.');
			}
		} catch (error) {
			console.error('Error saving file:', error);
			alert('Error saving file.');
		}
	}

	function handleFileSelection(fileName: string) {
		activeFile.set(fileName);
		fetchFileContent(fileName);
	}

	onMount(() => {
		fetchFileContent($activeFile);
	});
</script>

<div class="memory-sidebar">
	<div class="file-tabs">
		{#each files as file}
			<button class:active={$activeFile === file} on:click={() => handleFileSelection(file)}>
				{file}
			</button>
		{/each}
	</div>
	<div class="editor-container">
		<textarea bind:value={$fileContent}></textarea>
	</div>
	<div class="actions">
		<button on:click={saveFileContent}>Save</button>
	</div>
</div>

<style>
	.memory-sidebar {
		display: flex;
		flex-direction: column;
		height: 100%;
		background-color: #f3f3f3;
	}
	.file-tabs {
		display: flex;
		border-bottom: 1px solid #ccc;
	}
	.file-tabs button {
		padding: 10px 15px;
		border: none;
		background: none;
		cursor: pointer;
		font-size: 14px;
	}
	.file-tabs button.active {
		background-color: #fff;
		border-bottom: 2px solid #007bff;
	}
	.editor-container {
		flex-grow: 1;
		padding: 10px;
	}
	textarea {
		width: 100%;
		height: 100%;
		border: 1px solid #ccc;
		border-radius: 4px;
		padding: 8px;
		font-family: monospace;
		font-size: 14px;
		resize: none;
	}
	.actions {
		padding: 10px;
		border-top: 1px solid #ccc;
		text-align: right;
	}
	.actions button {
		padding: 8px 16px;
		border: none;
		background-color: #007bff;
		color: white;
		border-radius: 4px;
		cursor: pointer;
	}
</style>
