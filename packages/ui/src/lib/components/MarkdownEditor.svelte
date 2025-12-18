<script lang="ts">
   import { Crepe } from '@milkdown/crepe';
   import '@milkdown/crepe/theme/common/style.css';
    import '@milkdown/crepe/theme/frame.css';

    interface Props {
        value: string;
    }

    let { value }: Props = $props();

    function editor(node: HTMLElement) {
        const crepe = new Crepe({
			root: node,
			defaultValue: value,
			features: {
				[Crepe.Feature.Toolbar]: false,
				[Crepe.Feature.Latex]: false
			},
		});

        crepe.create();

        function setBlockHandleVisibility(show: boolean) {
            setTimeout(() => {
                const blockHandle = node.querySelector('.milkdown-block-handle');
                console.log('blockHandle', blockHandle);
                if (blockHandle) {
                    blockHandle.setAttribute('data-show', String(show));
                }
            }, 200);
        }

        function onMouseLeave() {
            console.log('onMouseLeave');
            setBlockHandleVisibility(false);
        }

        node.addEventListener('mouseleave', onMouseLeave);

        return {
            destroy: () => {
                crepe.destroy();
            }
        }
    }
</script>

<div use:editor role="presentation"></div>

<style>
    :global(.milkdown) {
        --crepe-color-background: var(--color-base-100);
        --crepe-color-on-background: var(--color-base-content);
        --crepe-color-surface: var(--color-base-300);
        --crepe-color-surface-low: var(--color-base-300);
        --crepe-color-on-surface: var(--color-base-content);
        --crepe-color-on-surface-variant: color-mix(in oklch, var(--color-base-content) 50%, transparent);
        --crepe-color-outline: color-mix(in oklch, var(--color-base-content) 50%, transparent);
        --crepe-color-primary: var(--color-primary);
        --crepe-color-secondary: var(--color-secondary);
        --crepe-color-on-secondary: var(--color-secondary-content);
        --crepe-color-inverse: var(--color-neutral);
        --crepe-color-on-inverse: var(--color-neutral-content);
        --crepe-color-inline-code: var(--color-error);
        --crepe-color-error: var(--color-error);
        --crepe-color-hover: var(--color-base-200);
        --crepe-color-selected: var(--color-base-300);
        --crepe-color-inline-area: var(--color-base-300);
    }

    :global(.milkdown .ProseMirror) {
        padding-top: 0;
        padding-bottom: 0;
        padding-left: 5.5rem;
        padding-right: 5.5rem;
    }
</style>