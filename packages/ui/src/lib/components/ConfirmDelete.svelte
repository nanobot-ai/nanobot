<script lang="ts">
    type Props = {
        title: string;
        message: string;
        onConfirm?: () => void;
        onCancel?: () => void;
    }
    let { title, message, onConfirm, onCancel }: Props = $props();
    let confirmDeleteModal = $state<HTMLDialogElement | null>(null);

    export function showModal() {
        confirmDeleteModal?.showModal();
    }
    export function close() {
        confirmDeleteModal?.close();
    }
</script>

<dialog bind:this={confirmDeleteModal} class="modal">
    <div class="modal-box dark:bg-base-200">
        <h3 class="text-lg font-bold">{title}</h3>
        {#if message}
            <p class="py-2">{message}</p>
        {:else}
            <p class="py-2">This will be permanently deleted and cannot be recovered.</p>
        {/if}
        <p>Are you sure you wish to continue?</p>
        <div class="modal-action">
            <button class="btn btn-ghost" onclick={() => {
                onCancel?.();
                confirmDeleteModal?.close();
            }}>
                Cancel
            </button>
            <button class="btn btn-error" onclick={() => {
                onConfirm?.();
                confirmDeleteModal?.close();
            }}>
                Confirm
            </button>
        </div>
    </div>
    <form method="dialog" class="modal-backdrop">
        <button>close</button>
    </form>
</dialog>