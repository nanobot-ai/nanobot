import type { MilkdownPlugin } from '@milkdown/ctx';
import { $prose } from '@milkdown/kit/utils';
import type { Node } from '@milkdown/prose/model';
import { Plugin, PluginKey } from '@milkdown/prose/state';
import { Decoration, DecorationSet } from '@milkdown/prose/view';
import type { EditorView } from '@milkdown/prose/view';

const variablePillKey = new PluginKey('variablePill');

// Regex to match $VariableName (alphanumeric and underscores after $)
const VARIABLE_REGEX = /\$([a-zA-Z_][a-zA-Z0-9_]*)/g;

// Regex to match partial variable input ($ followed by optional letters)
const PARTIAL_VARIABLE_REGEX = /\$([a-zA-Z_][a-zA-Z0-9_]*)?$/;

// Separator characters that complete a variable (space, non-breaking space, dash, slash)
const SEPARATOR_CHARS = new Set([' ', '\u00A0', '-', '/']);

export interface VariablePillOptions {
	onVariableAddition?: (variable: string) => void;
	onVariableDeletion?: (variable: string) => void;
	getAvailableVariables?: () => string[];
}

// Autocomplete state
interface AutocompleteState {
	active: boolean;
	query: string;
	startPos: number;
	endPos: number;
	selectedIndex: number;
	suggestions: string[];
}

// Pill editor state (for editing existing pills)
interface PillEditorState {
	active: boolean;
	variableStart: number;
	variableEnd: number;
	currentVariable: string;
	selectedIndex: number;
	suggestions: string[];
}

// Create autocomplete dropdown element
function createAutocompleteDropdown(): HTMLElement {
	const dropdown = document.createElement('div');
	dropdown.className = 'variable-autocomplete-dropdown';
	dropdown.style.cssText = `
		position: absolute;
		z-index: 1000;
		background: var(--color-base-100, #fff);
		border: 1px solid var(--color-base-300, #e5e5e5);
		border-radius: 8px;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
		max-height: 200px;
		overflow-y: auto;
		min-width: 150px;
		display: none;
	`;
	document.body.appendChild(dropdown);
	return dropdown;
}

// Render autocomplete suggestions
function renderAutocompleteSuggestions(
	dropdown: HTMLElement,
	suggestions: string[],
	selectedIndex: number,
	onSelect: (variable: string) => void
): void {
	dropdown.innerHTML = '';

	if (suggestions.length === 0) {
		dropdown.style.display = 'none';
		return;
	}

	suggestions.forEach((variable, index) => {
		const item = document.createElement('div');
		item.className = 'variable-autocomplete-item';
		item.textContent = `$${variable}`;
		item.style.cssText = `
			padding: 8px 12px;
			cursor: pointer;
			font-family: var(--default-font-family);
			font-size: 14px;
			color: var(--color-base-content, #333);
			${index === selectedIndex ? 'background: var(--color-primary, #3b82f6); color: var(--color-primary-content, #fff);' : ''}
		`;

		item.addEventListener('mouseenter', () => {
			// Update visual selection on hover
			dropdown.querySelectorAll('.variable-autocomplete-item').forEach((el, i) => {
				(el as HTMLElement).style.background = i === index ? 'var(--color-primary, #3b82f6)' : '';
				(el as HTMLElement).style.color =
					i === index ? 'var(--color-primary-content, #fff)' : 'var(--color-base-content, #333)';
			});
		});

		item.addEventListener('mousedown', (e) => {
			e.preventDefault();
			e.stopPropagation();
			onSelect(variable);
		});

		dropdown.appendChild(item);
	});

	dropdown.style.display = 'block';
}

// Position the dropdown near the cursor
function positionDropdown(dropdown: HTMLElement, view: EditorView, pos: number): void {
	const coords = view.coordsAtPos(pos);

	dropdown.style.left = `${coords.left}px`;
	dropdown.style.top = `${coords.bottom + 4}px`;

	// Ensure dropdown doesn't go off-screen
	requestAnimationFrame(() => {
		const dropdownRect = dropdown.getBoundingClientRect();
		if (dropdownRect.right > window.innerWidth) {
			dropdown.style.left = `${window.innerWidth - dropdownRect.width - 8}px`;
		}
		if (dropdownRect.bottom > window.innerHeight) {
			dropdown.style.top = `${coords.top - dropdownRect.height - 4}px`;
		}
	});
}

// Position dropdown near an element (for pill editor)
function positionDropdownNearElement(dropdown: HTMLElement, element: HTMLElement): void {
	const rect = element.getBoundingClientRect();

	dropdown.style.left = `${rect.left}px`;
	dropdown.style.top = `${rect.bottom + 4}px`;

	// Ensure dropdown doesn't go off-screen
	requestAnimationFrame(() => {
		const dropdownRect = dropdown.getBoundingClientRect();
		if (dropdownRect.right > window.innerWidth) {
			dropdown.style.left = `${window.innerWidth - dropdownRect.width - 8}px`;
		}
		if (dropdownRect.bottom > window.innerHeight) {
			dropdown.style.top = `${rect.top - dropdownRect.height - 4}px`;
		}
	});
}

// Render pill editor suggestions (similar to autocomplete but highlights current variable)
function renderPillEditorSuggestions(
	dropdown: HTMLElement,
	suggestions: string[],
	selectedIndex: number,
	currentVariable: string,
	onSelect: (variable: string) => void
): void {
	dropdown.innerHTML = '';

	if (suggestions.length === 0) {
		dropdown.style.display = 'none';
		return;
	}

	suggestions.forEach((variable, index) => {
		const item = document.createElement('div');
		item.className = 'variable-autocomplete-item';
		
		const isCurrent = variable === currentVariable;
		const isSelected = index === selectedIndex;
		
		item.innerHTML = `<span>$${variable}</span>${isCurrent ? '<span style="margin-left: 8px; opacity: 0.6; font-size: 12px;">(current)</span>' : ''}`;
		item.style.cssText = `
			padding: 8px 12px;
			cursor: pointer;
			font-family: var(--default-font-family);
			font-size: 14px;
			display: flex;
			align-items: center;
			justify-content: space-between;
			color: var(--color-base-content, #333);
			${isSelected ? 'background: var(--color-primary, #3b82f6); color: var(--color-primary-content, #fff);' : ''}
			${isCurrent && !isSelected ? 'opacity: 0.7;' : ''}
		`;

		item.addEventListener('mouseenter', () => {
			dropdown.querySelectorAll('.variable-autocomplete-item').forEach((el, i) => {
				(el as HTMLElement).style.background = i === index ? 'var(--color-primary, #3b82f6)' : '';
				(el as HTMLElement).style.color =
					i === index ? 'var(--color-primary-content, #fff)' : 'var(--color-base-content, #333)';
			});
		});

		item.addEventListener('mousedown', (e) => {
			e.preventDefault();
			e.stopPropagation();
			if (!isCurrent) {
				onSelect(variable);
			}
		});

		dropdown.appendChild(item);
	});

	dropdown.style.display = 'block';
}

// Get text before cursor in the current text node
function getTextBeforeCursor(doc: Node, pos: number): { text: string; startPos: number } | null {
	try {
		const $pos = doc.resolve(pos);
		const parent = $pos.parent;

		if (!parent.isTextblock) return null;

		const textContent = parent.textContent;
		const offset = $pos.parentOffset;
		const textBefore = textContent.slice(0, offset);

		// Find the start position of the parent textblock content
		const startPos = pos - offset;

		return { text: textBefore, startPos };
	} catch {
		return null;
	}
}

// Check for partial variable input and return autocomplete state
function checkForAutocomplete(
	doc: Node,
	pos: number,
	getAvailableVariables: () => string[]
): AutocompleteState | null {
	const textInfo = getTextBeforeCursor(doc, pos);
	if (!textInfo) return null;

	const { text, startPos } = textInfo;
	const match = text.match(PARTIAL_VARIABLE_REGEX);

	if (!match) return null;

	const query = match[1] || '';
	const matchStart = startPos + match.index!;
	const available = getAvailableVariables();

	// Filter suggestions (case insensitive)
	const suggestions = available.filter((v) => v.toLowerCase().startsWith(query.toLowerCase()));

	// Don't show autocomplete if no suggestions or exact match already typed
	if (suggestions.length === 0) return null;
	if (suggestions.length === 1 && suggestions[0].toLowerCase() === query.toLowerCase()) return null;

	return {
		active: true,
		query,
		startPos: matchStart,
		endPos: pos,
		selectedIndex: 0,
		suggestions
	};
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
							isCompleted
								? 'badge-primary variable-pill-completed pr-4! pl-2! py-1!'
								: 'px-2! py-1! text-base-content/50'
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

		// Autocomplete state
		let autocompleteState: AutocompleteState | null = null;
		let autocompleteDropdown: HTMLElement | null = null;
		let currentView: EditorView | null = null;

		// Pill editor state
		let pillEditorState: PillEditorState | null = null;
		let pillEditorDropdown: HTMLElement | null = null;
		let pillEditorElement: HTMLElement | null = null;

		// Function to insert selected variable
		function insertVariable(variable: string) {
			if (!currentView || !autocompleteState) return;

			const { state } = currentView;
			const startPos = autocompleteState.startPos;

			const tr = state.tr.replaceWith(
				startPos,
				autocompleteState.endPos,
				state.schema.text(`$${variable} `)
			);

			// Mark the variable as completed immediately so it gets the blue styling
			completedVariables.add(`${variable}:${startPos}`);

			// Refocus the editor before dispatch to ensure onChange callback fires
			// (clicking on dropdown causes focusout, but we need focus for the listener)
			currentView.focus();
			currentView.dispatch(tr);
			hideAutocomplete();
		}

		// Function to hide autocomplete
		function hideAutocomplete() {
			autocompleteState = null;
			if (autocompleteDropdown) {
				autocompleteDropdown.style.display = 'none';
			}
		}

		// Function to replace variable in pill editor
		function replaceVariable(newVariable: string) {
			if (!currentView || !pillEditorState) return;

			const { state } = currentView;
			const startPos = pillEditorState.variableStart;

			const tr = state.tr.replaceWith(
				startPos,
				pillEditorState.variableEnd,
				state.schema.text(`$${newVariable} `)
			);

			// Mark the new variable as completed immediately so it gets the blue styling
			// The new variable starts at the same position as the old one
			completedVariables.add(`${newVariable}:${startPos}`);

			// Refocus the editor before dispatch to ensure onChange callback fires
			// (clicking on dropdown causes focusout, but we need focus for the listener)
			currentView.focus();
			currentView.dispatch(tr);
			hidePillEditor();
		}

		// Function to hide pill editor
		function hidePillEditor() {
			pillEditorState = null;
			pillEditorElement = null;
			if (pillEditorDropdown) {
				pillEditorDropdown.style.display = 'none';
			}
		}

		// Function to show pill editor
		function showPillEditor(
			element: HTMLElement,
			start: number,
			end: number,
			currentVariable: string
		) {
			if (!options.getAvailableVariables) return;

			const available = options.getAvailableVariables();
			if (available.length === 0) return;

			// Create dropdown if it doesn't exist
			if (!pillEditorDropdown) {
				pillEditorDropdown = createAutocompleteDropdown();
			}

			// Find the index of the current variable in the list
			const currentIndex = available.indexOf(currentVariable);

			pillEditorState = {
				active: true,
				variableStart: start,
				variableEnd: end,
				currentVariable,
				selectedIndex: currentIndex >= 0 ? currentIndex : 0,
				suggestions: available
			};

			pillEditorElement = element;
			positionDropdownNearElement(pillEditorDropdown, element);
			renderPillEditorSuggestions(
				pillEditorDropdown,
				available,
				pillEditorState.selectedIndex,
				currentVariable,
				replaceVariable
			);
		}

		// Function to update autocomplete
		function updateAutocomplete(view: EditorView) {
			if (!options.getAvailableVariables) {
				hideAutocomplete();
				return;
			}

			currentView = view;
			const { state } = view;
			const { selection } = state;

			if (!selection.empty) {
				hideAutocomplete();
				return;
			}

			const newState = checkForAutocomplete(
				state.doc,
				selection.from,
				options.getAvailableVariables
			);

			if (!newState) {
				hideAutocomplete();
				return;
			}

			// Create dropdown if it doesn't exist
			if (!autocompleteDropdown) {
				autocompleteDropdown = createAutocompleteDropdown();
			}

			autocompleteState = newState;
			positionDropdown(autocompleteDropdown, view, newState.startPos);
			renderAutocompleteSuggestions(
				autocompleteDropdown,
				newState.suggestions,
				newState.selectedIndex,
				insertVariable
			);
		}

		return new Plugin({
			key: variablePillKey,
			state: {
				init(_, { doc }) {
					previousVariables = findAllVariables(doc);
					// Mark any variables that already have separators or are at end of document as completed
					for (const v of previousVariables) {
						if (
							hasSeparatorAt(doc, v.end) ||
							isFollowedByNewBlock(doc, v.end) ||
							isAtEndOfDocument(doc, v.end)
						) {
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
						const previousKeys = new Set(
							previousVariables.map((v) => `${v.variable}:${tr.mapping.map(v.start)}`)
						);

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
										// If user explicitly typed a separator, always add the variable
										// Only skip if at end of document and variable could be autocompleted
										if (!hasSep && !hasBlock && atEnd && options.getAvailableVariables) {
											const available = options.getAvailableVariables();
											const couldBeAutocompleted = available.some(
												(v) =>
													v.toLowerCase() !== prevVar.variable.toLowerCase() &&
													v.toLowerCase().startsWith(prevVar.variable.toLowerCase())
											);
											if (couldBeAutocompleted) {
												continue; // Don't add this variable yet, user might be autocompleting
											}
										}
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
								// If user explicitly typed a separator, always add the variable
								// Only skip if at end of document and variable could be autocompleted
								if (!hasSep && !hasBlock && atEnd && options.getAvailableVariables) {
									const available = options.getAvailableVariables();
									const couldBeAutocompleted = available.some(
										(v) =>
											v.toLowerCase() !== newVar.variable.toLowerCase() &&
											v.toLowerCase().startsWith(newVar.variable.toLowerCase())
									);
									if (couldBeAutocompleted) {
										continue; // Don't add this variable yet, user might be autocompleting
									}
								}

								// Clear any existing timer for this key
								if (pendingVariableTimers.has(key)) {
									clearTimeout(pendingVariableTimers.get(key));
								}

								// Set a debounced timer - only trigger if variable persists
								const timer = setTimeout(() => {
									pendingVariableTimers.delete(key);
									// Double-check the variable still exists and isn't already completed
									if (!completedVariables.has(key)) {
										// Re-check if variable could be autocompleted
										if (options.getAvailableVariables) {
											const available = options.getAvailableVariables();
											const couldBeAutocompleted = available.some(
												(v) =>
													v.toLowerCase() !== newVar.variable.toLowerCase() &&
													v.toLowerCase().startsWith(newVar.variable.toLowerCase())
											);
											if (couldBeAutocompleted) {
												return; // Don't add - user might be autocompleting
											}
										}
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

					// Handle autocomplete keyboard navigation
					if (autocompleteState && autocompleteState.active) {
						if (event.key === 'ArrowDown') {
							event.preventDefault();
							autocompleteState.selectedIndex = Math.min(
								autocompleteState.selectedIndex + 1,
								autocompleteState.suggestions.length - 1
							);
							if (autocompleteDropdown) {
								renderAutocompleteSuggestions(
									autocompleteDropdown,
									autocompleteState.suggestions,
									autocompleteState.selectedIndex,
									insertVariable
								);
							}
							return true;
						}

						if (event.key === 'ArrowUp') {
							event.preventDefault();
							autocompleteState.selectedIndex = Math.max(autocompleteState.selectedIndex - 1, 0);
							if (autocompleteDropdown) {
								renderAutocompleteSuggestions(
									autocompleteDropdown,
									autocompleteState.suggestions,
									autocompleteState.selectedIndex,
									insertVariable
								);
							}
							return true;
						}

						if (event.key === 'Enter' || event.key === 'Tab') {
							event.preventDefault();
							const selectedVar = autocompleteState.suggestions[autocompleteState.selectedIndex];
							if (selectedVar) {
								insertVariable(selectedVar);
							}
							return true;
						}

						if (event.key === 'Escape') {
							event.preventDefault();
							hideAutocomplete();
							return true;
						}
					}

					// Handle pill editor keyboard navigation
					if (pillEditorState && pillEditorState.active) {
						if (event.key === 'ArrowDown') {
							event.preventDefault();
							pillEditorState.selectedIndex = Math.min(
								pillEditorState.selectedIndex + 1,
								pillEditorState.suggestions.length - 1
							);
							if (pillEditorDropdown) {
								renderPillEditorSuggestions(
									pillEditorDropdown,
									pillEditorState.suggestions,
									pillEditorState.selectedIndex,
									pillEditorState.currentVariable,
									replaceVariable
								);
							}
							return true;
						}

						if (event.key === 'ArrowUp') {
							event.preventDefault();
							pillEditorState.selectedIndex = Math.max(pillEditorState.selectedIndex - 1, 0);
							if (pillEditorDropdown) {
								renderPillEditorSuggestions(
									pillEditorDropdown,
									pillEditorState.suggestions,
									pillEditorState.selectedIndex,
									pillEditorState.currentVariable,
									replaceVariable
								);
							}
							return true;
						}

						if (event.key === 'Enter' || event.key === 'Tab') {
							event.preventDefault();
							const selectedVar = pillEditorState.suggestions[pillEditorState.selectedIndex];
							if (selectedVar && selectedVar !== pillEditorState.currentVariable) {
								replaceVariable(selectedVar);
							} else {
								hidePillEditor();
							}
							return true;
						}

						if (event.key === 'Escape') {
							event.preventDefault();
							hidePillEditor();
							return true;
						}
					}

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

						// Hide autocomplete if clicking outside
						if (autocompleteDropdown && !autocompleteDropdown.contains(target)) {
							hideAutocomplete();
						}

						// Hide pill editor if clicking outside
						if (
							pillEditorDropdown &&
							!pillEditorDropdown.contains(target) &&
							target !== pillEditorElement &&
							!pillEditorElement?.contains(target)
						) {
							hidePillEditor();
						}

						// Handle click on completed variable pill
						if (target.classList.contains('variable-pill-completed')) {
							const rect = target.getBoundingClientRect();
							const clickX = event.clientX;
							const start = parseInt(target.getAttribute('data-start') || '0', 10);
							const end = parseInt(target.getAttribute('data-end') || '0', 10);
							const variable = target.getAttribute('data-variable') || '';

							// Check if click is in the rightmost 20px (where the × is)
							if (clickX > rect.right - 20) {
								// Delete the variable
								if (start !== end) {
									const tr = view.state.tr.delete(start, end);
									view.dispatch(tr);
									event.preventDefault();
									event.stopPropagation();
									return true;
								}
							} else {
								// Click on pill body - show variable selector
								currentView = view;
								
								// Toggle pill editor: hide if clicking same pill, show if different
								if (
									pillEditorState &&
									pillEditorState.variableStart === start &&
									pillEditorState.variableEnd === end
								) {
									hidePillEditor();
								} else {
									showPillEditor(target, start, end, variable);
								}
								event.preventDefault();
								event.stopPropagation();
								return true;
							}
						}

						return false;
					},

					blur() {
						// Delay hiding to allow click events on dropdown items
						setTimeout(() => {
							hideAutocomplete();
							hidePillEditor();
						}, 150);
						return false;
					}
				}
			},

			view(editorView) {
				currentView = editorView;
				return {
					update(view) {
						currentView = view;
						updateAutocomplete(view);
					},
					destroy() {
						if (autocompleteDropdown) {
							autocompleteDropdown.remove();
							autocompleteDropdown = null;
						}
						if (pillEditorDropdown) {
							pillEditorDropdown.remove();
							pillEditorDropdown = null;
						}
						pillEditorElement = null;
						currentView = null;
					}
				};
			}
		});
	});
}

// Default export for backwards compatibility (no callback)
export const variablePillPlugin: MilkdownPlugin = createVariablePillPlugin();
