import type { MilkdownPlugin } from '@milkdown/ctx';
import { $prose } from '@milkdown/kit/utils';
import type { Node } from '@milkdown/prose/model';
import { Plugin, PluginKey } from '@milkdown/prose/state';
import { Decoration, DecorationSet } from '@milkdown/prose/view';

const variablePillKey = new PluginKey('variablePill');

// Regex to match $VariableName (alphanumeric and underscores after $)
const VARIABLE_REGEX = /\$([a-zA-Z_][a-zA-Z0-9_]*)/g;

// Separator characters that complete a variable (space, non-breaking space, dash, slash)
const SEPARATOR_CHARS = new Set([' ', '\u00A0', '-', '/']);

export interface VariablePillOptions {
	onVariableAddition?: (variable: string) => void;
	onVariableDeletion?: (variable: string) => void;
}

interface VariableMatch {
	variable: string;
	start: number;
	end: number;
}

interface CompletedVariableInfo extends VariableMatch {
	isCompleted: boolean;
}

// Find all variables in the document with their completion status
function findAllVariablesWithStatus(
	doc: Node,
	completedVariables: Set<string>
): CompletedVariableInfo[] {
	const variables: CompletedVariableInfo[] = [];

	doc.descendants((node, pos) => {
		if (node.isText && node.text) {
			const text = node.text;
			let match;

			VARIABLE_REGEX.lastIndex = 0;
			while ((match = VARIABLE_REGEX.exec(text)) !== null) {
				const start = pos + match.index;
				const end = pos + match.index + match[0].length;
				const key = `${match[1]}:${start}`;

				variables.push({
					variable: match[1],
					start,
					end,
					isCompleted: completedVariables.has(key)
				});
			}
		}
	});

	return variables;
}

function findVariablesWithDecorations(
	doc: Parameters<typeof DecorationSet.create>[0],
	completedVariables: Set<string>
) {
	const decorations: Decoration[] = [];

	doc.descendants((node, pos) => {
		if (node.isText && node.text) {
			const text = node.text;
			let match;

			VARIABLE_REGEX.lastIndex = 0;
			while ((match = VARIABLE_REGEX.exec(text)) !== null) {
				const start = pos + match.index;
				const end = start + match[0].length;
				const key = `${match[1]}:${start}`;
				const isCompleted = completedVariables.has(key);

				// Add inline decoration for the variable pill
				decorations.push(
					Decoration.inline(start, end, {
						class: [
							'badge',
							'badge-soft',
							'font-semibold',
							isCompleted ? 'badge-primary variable-pill-completed pr-4! pl-2! py-1!' : 'px-2! py-1! text-base-content/50',
						].join(' '),
						'data-variable': match[1],
						'data-start': String(start),
						'data-end': String(end),
						'data-completed': isCompleted ? 'true' : 'false'
					})
				);
			}
		}
	});

	return DecorationSet.create(doc, decorations);
}

// Find all variables in the document (without requiring separators)
function findAllVariables(doc: Node): VariableMatch[] {
	const variables: VariableMatch[] = [];

	doc.descendants((node, pos) => {
		if (node.isText && node.text) {
			const text = node.text;
			let match;

			VARIABLE_REGEX.lastIndex = 0;
			while ((match = VARIABLE_REGEX.exec(text)) !== null) {
				variables.push({
					variable: match[1],
					start: pos + match.index,
					end: pos + match.index + match[0].length
				});
			}
		}
	});

	return variables;
}

// Check if there's a separator character at the given position in the document
function hasSeparatorAt(doc: Node, pos: number): boolean {
	try {
		const $pos = doc.resolve(pos);
		const nodeAfter = $pos.nodeAfter;
		const parent = $pos.parent;

		// Check if the node right after is text starting with a separator
		if (nodeAfter?.isText && nodeAfter.text) {
			const char = nodeAfter.text[0];
			return SEPARATOR_CHARS.has(char);
		}

		// Check within parent textblock at the offset position
		if (parent.isTextblock) {
			const offset = $pos.parentOffset;
			const textContent = parent.textContent;
			if (offset < textContent.length) {
				const char = textContent[offset];
				return SEPARATOR_CHARS.has(char);
			}
		}

		// No separator found - do NOT treat end-of-node as separator
		return false;
	} catch {
		return false;
	}
}

// Check if the variable is now followed by a new block (Enter was pressed)
function isFollowedByNewBlock(doc: Node, pos: number): boolean {
	try {
		const $pos = doc.resolve(pos);

		// Check if we're at the end of a textblock and the next node is a block
		if ($pos.parentOffset === $pos.parent.content.size) {
			// We're at the end of the parent node
			const indexInGrandparent = $pos.index($pos.depth - 1);
			const grandparent = $pos.node($pos.depth - 1);

			// Check if there's a sibling block after this one
			if (indexInGrandparent + 1 < grandparent.childCount) {
				return true; // There's a block after this one
			}
		}

		return false;
	} catch {
		return false;
	}
}

// Check if the position is at the end of the document content
function isAtEndOfDocument(doc: Node, pos: number): boolean {
	try {
		const $pos = doc.resolve(pos);

		// Check if we're at the end of the parent textblock
		if ($pos.parentOffset !== $pos.parent.content.size) {
			return false;
		}

		// Walk up the tree to check if we're at the end of the document
		for (let depth = $pos.depth - 1; depth >= 0; depth--) {
			const node = $pos.node(depth);
			const index = $pos.index(depth);
			// If there are more children after this one, we're not at the end
			if (index + 1 < node.childCount) {
				return false;
			}
		}

		return true;
	} catch {
		return false;
	}
}

// Find a completed variable at or around the given position
function findCompletedVariableAtPosition(
	doc: Node,
	pos: number,
	completedVariables: Set<string>,
	checkType: 'before' | 'after' | 'inside'
): CompletedVariableInfo | null {
	const variables = findAllVariablesWithStatus(doc, completedVariables);

	for (const v of variables) {
		if (!v.isCompleted) continue;

		switch (checkType) {
			case 'before':
				// Cursor is right before the variable (for Delete key)
				if (pos === v.start) return v;
				break;
			case 'after':
				// Cursor is right after the variable (for Backspace key)
				if (pos === v.end) return v;
				break;
			case 'inside':
				// Cursor is inside the variable
				if (pos > v.start && pos < v.end) return v;
				break;
		}
	}

	return null;
}

export function createVariablePillPlugin(options: VariablePillOptions = {}): MilkdownPlugin {
	return $prose(() => {
		// Track variables from previous state
		let previousVariables: VariableMatch[] = [];
		const completedVariables = new Set<string>(); // "variable:start" keys that already triggered
		
		// Debounce timers for newly introduced variables that already have separators
		// This prevents spam when typing a variable in the middle of content
		const pendingVariableTimers = new Map<string, ReturnType<typeof setTimeout>>();
		const DEBOUNCE_MS = 500;

		return new Plugin({
			key: variablePillKey,
			state: {
				init(_, { doc }) {
					previousVariables = findAllVariables(doc);
					// Mark any variables that already have separators or are at end of document as completed
					for (const v of previousVariables) {
						if (hasSeparatorAt(doc, v.end) || isFollowedByNewBlock(doc, v.end) || isAtEndOfDocument(doc, v.end)) {
							completedVariables.add(`${v.variable}:${v.start}`);
						}
					}
					return findVariablesWithDecorations(doc, completedVariables);
				},
				apply(tr, decorations) {
					if (tr.docChanged) {
						// Map completed variable positions through the transaction mapping
						// This keeps positions valid when text is inserted/deleted before them
						const remappedCompleted = new Set<string>();
						for (const key of completedVariables) {
							const [varName, startStr] = key.split(':');
							const oldStart = parseInt(startStr, 10);
							const newStart = tr.mapping.map(oldStart);
							remappedCompleted.add(`${varName}:${newStart}`);
						}
						completedVariables.clear();
						for (const key of remappedCompleted) {
							completedVariables.add(key);
						}
						
						// Also remap pending timer positions
						const remappedTimers = new Map<string, ReturnType<typeof setTimeout>>();
						for (const [key, timer] of pendingVariableTimers) {
							const [varName, startStr] = key.split(':');
							const oldStart = parseInt(startStr, 10);
							const newStart = tr.mapping.map(oldStart);
							remappedTimers.set(`${varName}:${newStart}`, timer);
						}
						pendingVariableTimers.clear();
						for (const [key, timer] of remappedTimers) {
							pendingVariableTimers.set(key, timer);
						}
						
						const newVariables = findAllVariables(tr.doc);
						const previousKeys = new Set(previousVariables.map((v) => `${v.variable}:${tr.mapping.map(v.start)}`));

						if (options.onVariableAddition) {
							// For each variable that existed BEFORE this transaction,
							// check if it now has a separator (meaning user just typed one)
							for (const prevVar of previousVariables) {
								// Map the old position to the new document
								const mappedStart = tr.mapping.map(prevVar.start);
								const key = `${prevVar.variable}:${mappedStart}`;

								// Skip if already completed
								if (completedVariables.has(key)) {
									continue;
								}

								// Find the same variable in the new document (using mapped position)
								const newVar = newVariables.find(
									(v) => v.variable === prevVar.variable && v.start === mappedStart
								);

								// If the variable still exists, check if there's now a separator after it
								if (newVar) {
									const hasSep = hasSeparatorAt(tr.doc, newVar.end);
									const hasBlock = isFollowedByNewBlock(tr.doc, newVar.end);
									const atEnd = isAtEndOfDocument(tr.doc, newVar.end);

									if (hasSep || hasBlock || atEnd) {
										options.onVariableAddition(prevVar.variable);
										completedVariables.add(key);
									}
								}
							}
						}

						// Check newly introduced variables (e.g. from paste or typing in middle of content)
						// that already have separators - debounce these to avoid spam when typing
						for (const newVar of newVariables) {
							const key = `${newVar.variable}:${newVar.start}`;

							// Skip if this variable existed before or is already completed
							if (previousKeys.has(key) || completedVariables.has(key)) {
								// Clear any pending timer since the variable is stable now
								if (pendingVariableTimers.has(key)) {
									clearTimeout(pendingVariableTimers.get(key));
									pendingVariableTimers.delete(key);
								}
								continue;
							}

							// If this new variable already has a separator or is at end, debounce before marking as completed
							const hasSep = hasSeparatorAt(tr.doc, newVar.end);
							const hasBlock = isFollowedByNewBlock(tr.doc, newVar.end);
							const atEnd = isAtEndOfDocument(tr.doc, newVar.end);

							if (hasSep || hasBlock || atEnd) {
								// Clear any existing timer for this key
								if (pendingVariableTimers.has(key)) {
									clearTimeout(pendingVariableTimers.get(key));
								}
								
								// Set a debounced timer - only trigger if variable persists
								const timer = setTimeout(() => {
									pendingVariableTimers.delete(key);
									// Double-check the variable still exists and isn't already completed
									if (!completedVariables.has(key)) {
										if (options.onVariableAddition) {
											options.onVariableAddition(newVar.variable);
										}
										completedVariables.add(key);
									}
								}, DEBOUNCE_MS);
								
								pendingVariableTimers.set(key, timer);
							}
						}

						// Update previous variables for next transaction
						previousVariables = newVariables;

						// Clean up completed set - remove entries for variables that no longer exist
						// and call onVariableDeletion if no instances of the variable remain
						const currentKeys = new Set(newVariables.map((v) => `${v.variable}:${v.start}`));
						const deletedCompletedVars: string[] = [];

						for (const key of completedVariables) {
							if (!currentKeys.has(key)) {
								const varName = key.split(':')[0];
								deletedCompletedVars.push(varName);
								completedVariables.delete(key);
							}
						}
						
						// Clean up pending timers for variables that no longer exist
						for (const key of pendingVariableTimers.keys()) {
							if (!currentKeys.has(key)) {
								clearTimeout(pendingVariableTimers.get(key));
								pendingVariableTimers.delete(key);
							}
						}

						// Call onVariableDeletion for completed variables that no longer have any instances
						if (options.onVariableDeletion) {
							for (const varName of deletedCompletedVars) {
								// Check if there are any remaining instances of this variable in the document
								const hasRemainingInstances = newVariables.some((v) => v.variable === varName);
								if (!hasRemainingInstances) {
									options.onVariableDeletion(varName);
								}
							}
						}

						return findVariablesWithDecorations(tr.doc, completedVariables);
					}
					return decorations.map(tr.mapping, tr.doc);
				}
			},
			props: {
				decorations(state) {
					return this.getState(state);
				},

				handleKeyDown(view, event) {
					const { state } = view;
					const { selection } = state;

					// Only handle when there's no text selection (cursor is collapsed)
					if (!selection.empty) return false;

					const pos = selection.from;

					if (event.key === 'Backspace') {
						// Check if cursor is right after a completed variable
						let targetVar = findCompletedVariableAtPosition(
							state.doc,
							pos,
							completedVariables,
							'after'
						);

						// Also check if cursor is inside a completed variable
						if (!targetVar) {
							targetVar = findCompletedVariableAtPosition(
								state.doc,
								pos,
								completedVariables,
								'inside'
							);
						}

						if (targetVar) {
							// Delete the entire variable
							const tr = state.tr.delete(targetVar.start, targetVar.end);
							view.dispatch(tr);
							return true; // Prevent default backspace behavior
						}
					}

					if (event.key === 'Delete') {
						// Check if cursor is right before a completed variable
						let targetVar = findCompletedVariableAtPosition(
							state.doc,
							pos,
							completedVariables,
							'before'
						);

						// Also check if cursor is inside a completed variable
						if (!targetVar) {
							targetVar = findCompletedVariableAtPosition(
								state.doc,
								pos,
								completedVariables,
								'inside'
							);
						}

						if (targetVar) {
							// Delete the entire variable
							const tr = state.tr.delete(targetVar.start, targetVar.end);
							view.dispatch(tr);
							return true; // Prevent default delete behavior
						}
					}

					return false;
				},

				handleDOMEvents: {
					mousedown(view, event) {
						const target = event.target as HTMLElement;

						// Handle click on delete button (::after pseudo-element area)
						if (target.classList.contains('variable-pill-completed')) {
							const rect = target.getBoundingClientRect();
							const clickX = event.clientX;
							// Check if click is in the rightmost 20px (where the × is)
							if (clickX > rect.right - 20) {
								const start = parseInt(target.getAttribute('data-start') || '0', 10);
								const end = parseInt(target.getAttribute('data-end') || '0', 10);

								if (start !== end) {
									const tr = view.state.tr.delete(start, end);
									view.dispatch(tr);
									event.preventDefault();
									event.stopPropagation();
									return true;
								}
							}
						}

						return false;
					}
				}
			}
		});
	});
}

// Default export for backwards compatibility (no callback)
export const variablePillPlugin: MilkdownPlugin = createVariablePillPlugin();
