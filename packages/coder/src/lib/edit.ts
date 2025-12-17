/**
 * Performs exact string replacement in file content
 */
export function performEdit(
	content: string,
	oldString: string,
	newString: string,
	replaceAll: boolean,
): { success: boolean; newContent?: string; error?: string } {
	if (oldString === newString) {
		return {
			success: false,
			error: "old_string and new_string must be different",
		};
	}

	if (!replaceAll) {
		// Check if old_string appears exactly once
		const firstIndex = content.indexOf(oldString);
		if (firstIndex === -1) {
			return {
				success: false,
				error: "old_string not found in file",
			};
		}

		const lastIndex = content.lastIndexOf(oldString);
		if (firstIndex !== lastIndex) {
			return {
				success: false,
				error:
					"old_string appears multiple times in file. Use replace_all: true or provide more context to make it unique",
			};
		}

		const newContent =
			content.substring(0, firstIndex) +
			newString +
			content.substring(firstIndex + oldString.length);

		return { success: true, newContent };
	}

	// Replace all occurrences
	if (!content.includes(oldString)) {
		return {
			success: false,
			error: "old_string not found in file",
		};
	}

	const newContent = content.split(oldString).join(newString);
	return { success: true, newContent };
}
