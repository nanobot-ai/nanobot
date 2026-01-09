<script lang="ts">
import { Crepe } from '@milkdown/crepe';
import '@milkdown/crepe/theme/common/style.css';
import '@milkdown/crepe/theme/frame.css';
import type { MilkdownPlugin } from '@milkdown/kit/ctx';
import { listener, listenerCtx } from '@milkdown/kit/plugin/listener';
import { replaceAll } from '@milkdown/kit/utils';
	import { untrack } from 'svelte';

    interface Props {
        value: string;
        blockEditEnabled?: boolean;
        plugins: MilkdownPlugin[];
        onChange?: (value: string) => void;
        readonly?: boolean;
    }

    let { value, blockEditEnabled, plugins = [], onChange, readonly }: Props = $props();

    let focused = $state(false);
    let prevValue = $state(untrack(() => value));
    let editorNode: HTMLElement | null = null;
    let isCrepeReady = false;
    let crepe: Crepe | null = null;

    async function createEditor(node: HTMLElement, enableBlockEdit: boolean, isReadonly: boolean) {
        const instance = new Crepe({
            root: node,
            defaultValue: value,
            features: {
                [Crepe.Feature.Toolbar]: false,
                [Crepe.Feature.Latex]: false,
                [Crepe.Feature.BlockEdit]: enableBlockEdit
            },
        });
        
        if (isReadonly) {
            instance.setReadonly(true);
        }

        let isFirstUpdate = true;
        instance.editor
            .config((ctx) => {
                ctx.get(listenerCtx).markdownUpdated((_, markdown, prevMarkdown) => {
                    if (isFirstUpdate) {
                        isFirstUpdate = false;
                        if (prevValue !== '') {
                            return;
                        }
                    }
                    
                    if (markdown === prevMarkdown) return;
                    if (!focused) return;
                    onChange?.(markdown);
                    prevValue = markdown;
                });
            })
            .use(listener);

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
            // Access blockEditEnabled and readonly to create dependencies
            const enableBlockEdit = blockEditEnabled;
            const isReadonly = readonly;
            const node = editorNode;
            
            destroyEditor();
            // Use untrack to prevent `value` (accessed in createEditor) from becoming a dependency
            untrack(() => {
                createEditor(node, enableBlockEdit ?? false, isReadonly ?? false).then((instance) => {
                    crepe = instance;
                    isCrepeReady = true;
                });
            });
        }

        return () => {
            destroyEditor();
            isCrepeReady = false;
        };
    });

    $effect(() => {
        if (value !== prevValue) {
            if (!focused && crepe && isCrepeReady) {
                setValue(value);
            }
            prevValue = value;
        }
    });

    function setValue(newValue: string) {
        if (crepe) {
            crepe.editor.action(replaceAll(newValue));
        }
    }

    function editor(node: HTMLElement) {
        editorNode = node;

        function onMouseLeave() {
            const blockHandle = node.querySelector('.milkdown-block-handle');
            if (blockHandle) {
                blockHandle.setAttribute('data-show', 'false');
            }
        }

        function onFocusIn() {
            focused = true;
        }

        function onFocusOut() {
            focused = false;
        }

        node.addEventListener('mouseleave', onMouseLeave);
        node.addEventListener('focusin', onFocusIn);
        node.addEventListener('focusout', onFocusOut);

        return {
            destroy: () => {
                node.removeEventListener('mouseleave', onMouseLeave);
                node.removeEventListener('focusin', onFocusIn);
                node.removeEventListener('focusout', onFocusOut);
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
        --crepe-color-surface: var(--color-base-200);
        --crepe-color-surface-low: var(--color-base-100);
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
        --crepe-color-hover: var(--color-base-300);
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

    :global(.milkdown .milkdown-code-block) {
        border-radius: var(--radius-box);
    }
</style>