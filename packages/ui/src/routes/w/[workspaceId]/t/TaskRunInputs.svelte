<script lang="ts">
	import { X } from "@lucide/svelte";
	import type { Input, Task } from "./types";


    interface Props {
        onSubmit: (formData: (Input & { value: string })[]) => void;
        task?: Task | null;
        additionalInputs?: Input[];
    }

    let { onSubmit, task, additionalInputs }: Props = $props();
    
    let inputsModal = $state<HTMLDialogElement | null>(null);
    let runFormData = $state<(Input & { value: string })[]>([]);
    let name = $derived(task?.name || task?.steps[0].name || '');

    export function showModal() {
        const visibleInputMapping = new Map(additionalInputs?.map((input) => [input.id, input]) ?? []);
        runFormData = (task?.inputs ?? []).map((input) => ({
            ...input,
            ...(visibleInputMapping.get(input.id) ?? {}),
            value: input.default || visibleInputMapping.get(input.id)?.default || '',
        }));
        inputsModal?.showModal();
    }

    export function close() {
        inputsModal?.close();
    }
</script>

<dialog bind:this={inputsModal} class="modal">
    <div class="modal-box bg-base-200 dark:bg-base-100 p-0 border border-transparent dark:border-base-300">
      <form method="dialog">
          <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2">
              <X class="size-4" />
          </button>
        </form>
      <h4 class="text-lg font-semibold p-4 bg-base-100 dark:bg-base-200">Run {name || 'Task'}</h4>
      <div class="p-4 flex flex-col gap-2">
          {#each runFormData as input (input.id)}
              <label class="input w-full">
                  <span class="label h-full font-semibold text-primary bg-primary/15">{input.name}</span>
                  <input type="text" bind:value={input.value} placeholder={input.description} />
              </label>
          {/each}
      </div>
      <div class="modal-action mt-0 px-4 py-2 bg-base-100 dark:bg-base-200">
          <form method="dialog">
              <button class="btn btn-ghost" onclick={() => inputsModal?.close()}>Cancel</button>
              <button class="btn btn-primary" onclick={() => {
                onSubmit(runFormData);
                inputsModal?.close();
              }}>
                Run
              </button>
          </form>
      </div>
    </div>
    <form method="dialog" class="modal-backdrop">
      <button>close</button>
    </form>
  </dialog>
  