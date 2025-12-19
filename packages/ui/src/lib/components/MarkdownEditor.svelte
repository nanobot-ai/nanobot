<script lang="ts">
    import { Crepe } from '@milkdown/crepe';
    import '@milkdown/crepe/theme/common/style.css';
    import '@milkdown/crepe/theme/frame.css';
	import type { MilkdownPlugin } from '@milkdown/kit/ctx';

    interface Props {
        value: string;
        blockEditEnabled: boolean;
        plugins: MilkdownPlugin[];
    }

    let { value, blockEditEnabled, plugins = [] }: Props = $props();

    let editorNode: HTMLElement | null = null;
    let isCrepeReady = false;
    let crepe: Crepe | null = null;

    async function createEditor(node: HTMLElement, enableBlockEdit: boolean) {
        const instance = new Crepe({
            root: node,
            defaultValue: value,
            features: {
                [Crepe.Feature.Toolbar]: false,
                [Crepe.Feature.Latex]: false,
                [Crepe.Feature.BlockEdit]: enableBlockEdit
            },
        });

        // Apply any additional plugins
		for (const plugin of plugins) {
			instance.editor.use(plugin);
		}
        
        await instance.create();

        if (enableBlockEdit) {
            const proseMirror = node.querySelector('.ProseMirror');
            if (proseMirror) {
                proseMirror.classList.add('block-editor-enabled');
            }
        }

        return instance;
    }

    function destroyEditor() {
        if (crepe && isCrepeReady) {
            crepe.destroy();
            crepe = null;
        }
    }

    $effect(() => {
        if (editorNode) {
            // Access blockEditEnabled to create dependency
            const enableBlockEdit = blockEditEnabled;
            const node = editorNode;
            
            destroyEditor();
            createEditor(node, enableBlockEdit).then((instance) => {
                crepe = instance;
                isCrepeReady = true;
            });
        }

        return () => {
            destroyEditor();
            isCrepeReady = false;
        };
    });

    function editor(node: HTMLElement) {
        editorNode = node;

        function onMouseLeave() {
            const blockHandle = node.querySelector('.milkdown-block-handle');
            if (blockHandle) {
                blockHandle.setAttribute('data-show', 'false');
            }
        }

        node.addEventListener('mouseleave', onMouseLeave);

        return {
            destroy: () => {
                node.removeEventListener('mouseleave', onMouseLeave);
                editorNode = null;
            }
        };
    }
</script>

<div use:editor></div>

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
        padding-left: 0.25rem;
        padding-right: 0.25rem;
    }
    
    :global(.milkdown .ProseMirror.block-editor-enabled) {
        padding-left: 5.5rem;
        padding-right: 5.5rem;
    }
</style>