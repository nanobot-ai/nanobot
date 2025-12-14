<script lang="ts">
	interface Props {
		elicitation: Elicitation;
		open?: boolean;
		onresult?: (result: ElicitationResult) => void;
	}

	import type { Elicitation, ElicitationResult, PrimitiveSchemaDefinition } from '$lib/types';
	import { Copy } from '@lucide/svelte';

	let { elicitation, open = false, onresult }: Props = $props();

	let formData = $state<{ [key: string]: string | number | boolean }>({});
	let showCopiedTooltip = $state(false);

	// Initialize form data with defaults
	$effect(() => {
		const newFormData: { [key: string]: string | number | boolean } = {};

		for (const [key, schema] of Object.entries(elicitation.requestedSchema.properties)) {
			if (schema.type === 'boolean' && schema.default !== undefined) {
				newFormData[key] = schema.default;
			} else if (
				schema.type === 'string' ||
				schema.type === 'number' ||
				schema.type === 'integer'
			) {
				newFormData[key] = schema.type === 'string' ? '' : 0;
			} else if ('enum' in schema && schema.enum) {
				newFormData[key] = (schema.enum as string[])[0] || '';
			}
		}

		formData = newFormData;
	});

	function handleAccept() {
		onresult?.({
			action: 'accept',
			content: { ...formData }
		});
	}

	function handleDecline() {
		onresult?.({
			action: 'decline'
		});
	}

	function handleCancel() {
		onresult?.({
			action: 'cancel'
		});
	}

	function isRequired(key: string): boolean {
		return elicitation.requestedSchema.required?.includes(key) ?? false;
	}

	function getFieldTitle(key: string, schema: PrimitiveSchemaDefinition): string {
		return schema.title || key;
	}

	function validateForm(): boolean {
		if (!elicitation.requestedSchema.required) return true;

		for (const requiredField of elicitation.requestedSchema.required) {
			const value = formData[requiredField];
			if (value === undefined || value === '' || value === null) {
				return false;
			}
		}
		return true;
	}

	function isOAuthElicitation(): boolean {
		return Boolean(elicitation._meta?.['ai.nanobot.meta/oauth-url']);
	}

	function getOAuthUrl(): string {
		return elicitation._meta?.['ai.nanobot.meta/oauth-url'] as string;
	}

	function getServerName(): string {
		return (elicitation._meta?.['ai.nanobot.meta/server-name'] as string) || 'MCP Server';
	}

	function openOAuthLink() {
		const url = getOAuthUrl();
		window.open(url, '_blank');
		handleAccept();
	}

	async function copyToClipboard() {
		const url = getOAuthUrl();
		await navigator.clipboard.writeText(url);
		showCopiedTooltip = true;
		setTimeout(() => {
			showCopiedTooltip = false;
		}, 2000);
	}
</script>

{#if open}
	<dialog class="modal-open modal">
		<div class="modal-box w-full max-w-2xl">
			<form method="dialog">
				<button
					class="btn absolute top-2 right-2 btn-circle btn-ghost btn-sm"
					onclick={handleCancel}>✕</button
				>
			</form>

			{#if isOAuthElicitation()}
				<!-- OAuth Authentication Dialog -->
				<h3 class="mb-4 text-lg font-bold">Authentication Required</h3>

				<div class="mb-6">
					<p class="mb-4 text-base-content/80">
						The <strong>{getServerName()}</strong> server requires authentication to continue.
					</p>
					<p class="mb-4 text-base-content/80">Please click the link below to authenticate:</p>

					<div class="group relative mb-4 rounded-lg bg-base-200 p-4">
						<p class="pr-8 font-mono text-sm break-all text-base-content/90">{getOAuthUrl()}</p>
						<button
							type="button"
							class="btn absolute top-2 right-2 opacity-60 btn-ghost transition-opacity btn-xs hover:opacity-100"
							onclick={copyToClipboard}
							title="Copy to clipboard"
						>
							<Copy class="h-4 w-4" />
						</button>
						{#if showCopiedTooltip}
							<div
								class="absolute -top-8 right-2 rounded bg-success px-2 py-1 text-xs text-success-content shadow-lg transition-opacity duration-500 {showCopiedTooltip
									? 'opacity-100'
									: 'opacity-0'}"
							>
								Copied!
							</div>
						{/if}
					</div>
				</div>

				<div class="modal-action">
					<button type="button" class="btn btn-error" onclick={handleDecline}> Decline </button>
					<button type="button" class="btn btn-success" onclick={openOAuthLink}>
						Authenticate
					</button>
				</div>
			{:else}
				<!-- Generic Elicitation Form -->
				<h3 class="mb-4 text-lg font-bold">Information Request</h3>

				<div class="mb-6">
					<p class="whitespace-pre-wrap text-base-content/80">{elicitation.message}</p>
				</div>

				<form
					class="space-y-4"
					onsubmit={(e) => {
						e.preventDefault();
						handleAccept();
					}}
				>
					{#each Object.entries(elicitation.requestedSchema.properties) as [key, schema] (key)}
						<div class="form-control">
							<label class="label" for={key}>
								<span class="label-text font-medium">
									{getFieldTitle(key, schema)}
									{#if isRequired(key)}
										<span class="text-error">*</span>
									{/if}
								</span>
							</label>

							{#if schema.description}
								<div class="label">
									<span class="label-text-alt text-base-content/60">{schema.description}</span>
								</div>
							{/if}

							{#if schema.type === 'string' && 'enum' in schema}
								<!-- Enum/Select field -->
								<select
									id={key}
									bind:value={formData[key]}
									class="select-bordered select w-full"
									required={isRequired(key)}
								>
									{#each schema.enum as option, i (option)}
										<option value={option}>
											{schema.enumNames?.[i] || option}
										</option>
									{/each}
								</select>
							{:else if schema.type === 'boolean'}
								<!-- Boolean/Checkbox field -->
								<div class="form-control">
									<label class="label cursor-pointer justify-start gap-3">
										<input
											id={key}
											type="checkbox"
											checked={Boolean(formData[key])}
											onchange={(e) => (formData[key] = e.currentTarget.checked)}
											class="checkbox"
										/>
										<span class="label-text">Enable</span>
									</label>
								</div>
							{:else if schema.type === 'number' || schema.type === 'integer'}
								<!-- Number field -->
								<input
									id={key}
									type="number"
									bind:value={formData[key]}
									class="input-bordered input w-full"
									required={isRequired(key)}
									min={schema.minimum}
									max={schema.maximum}
									step={schema.type === 'integer' ? '1' : 'any'}
								/>
							{:else if schema.type === 'string'}
								<!-- String field -->
								{#if schema.format === 'email'}
									<input
										id={key}
										type="email"
										bind:value={formData[key]}
										class="input-bordered input w-full"
										required={isRequired(key)}
										minlength={schema.minLength}
										maxlength={schema.maxLength}
									/>
								{:else if schema.format === 'uri'}
									<input
										id={key}
										type="url"
										bind:value={formData[key]}
										class="input-bordered input w-full"
										required={isRequired(key)}
										minlength={schema.minLength}
										maxlength={schema.maxLength}
									/>
								{:else if schema.format === 'date'}
									<input
										id={key}
										type="date"
										bind:value={formData[key]}
										class="input-bordered input w-full"
										required={isRequired(key)}
									/>
								{:else if schema.format === 'date-time'}
									<input
										id={key}
										type="datetime-local"
										bind:value={formData[key]}
										class="input-bordered input w-full"
										required={isRequired(key)}
									/>
								{:else}
									<input
										id={key}
										type="text"
										bind:value={formData[key]}
										class="input-bordered input w-full"
										required={isRequired(key)}
										minlength={schema.minLength}
										maxlength={schema.maxLength}
									/>
								{/if}
							{/if}
						</div>
					{/each}
				</form>

				<div class="modal-action">
					<button type="button" class="btn btn-error" onclick={handleDecline}> Decline </button>
					<button
						type="button"
						class="btn btn-primary"
						disabled={!validateForm()}
						onclick={handleAccept}
					>
						Accept
					</button>
				</div>
			{/if}
		</div>
	</dialog>
{/if}
