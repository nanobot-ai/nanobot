import type { ElicitRequestFormParams } from "@modelcontextprotocol/sdk/types.js";
import { createTool, toolResult } from "@nanobot-ai/nanomcp";
import * as z from "zod";

const optionSchema = z.object({
	label: z
		.string()
		.describe(
			"The display text for this option that the user will see and select. Should be concise (1-5 words) and clearly describe the choice.",
		),
	description: z
		.string()
		.describe(
			"Explanation of what this option means or what will happen if chosen. Useful for providing context about trade-offs or implications.",
		),
});

const questionSchema = z.object({
	question: z
		.string()
		.describe(
			'The complete question to ask the user. Should be clear, specific, and end with a question mark. Example: "Which library should we use for date formatting?" If multiSelect is true, phrase it accordingly, e.g. "Which features do you want to enable?"',
		),
	header: z
		.string()
		.describe(
			'Very short label displayed as a chip/tag (max 12 chars). Examples: "Auth method", "Library", "Approach".',
		),
	options: z
		.array(optionSchema)
		.min(2)
		.max(4)
		.describe(
			"The available choices for this question. Must have 2-4 options. Each option should be a distinct, mutually exclusive choice (unless multiSelect is enabled). There should be no 'Other' option, that will be provided automatically.",
		)
		.optional(),
	multiSelect: z
		.boolean()
		.describe(
			"Set to true to allow the user to select multiple options instead of just one. Use when choices are not mutually exclusive.",
		)
		.optional(),
});

const schema = z.object({
	questions: z
		.array(questionSchema)
		.min(1)
		.max(4)
		.describe("Questions to ask the user (1-4 questions)"),
});

export default createTool({
	title: "AskUserQuestion",
	description:
		'Use this tool when you need to ask the user questions during execution. This allows you to:\n1. Gather user preferences or requirements\n2. Clarify ambiguous instructions\n3. Get decisions on implementation choices as you work\n4. Offer choices to the user about what direction to take.\n\nUsage notes:\n- Users will always be able to select "Other" to provide custom text input\n- Use multiSelect: true to allow multiple answers to be selected for a question\n',
	messages: {
		invoking: "Asking user question",
		invoked: "User question answered",
	},
	inputSchema: schema,
	async handler(args, ctx) {
		const { questions } = args;

		console.log(ctx.elicit, !ctx.elicit);
		if (!ctx.elicit) {
			return toolResult.error("Client doesn't support elicitation");
		}

		try {
			// Build the elicitation message
			const message =
				questions.length === 1
					? questions[0].question
					: `Please answer the following questions:\n${questions.map((q, i) => `${i + 1}. ${q.question}`).join("\n")}`;

			// Build JSON Schema manually with enum and enumNames for MCP
			const properties: ElicitRequestFormParams["requestedSchema"]["properties"] =
				{};
			const required: string[] = [];

			for (let i = 0; i < questions.length; i++) {
				const q = questions[i];
				required.push(q.header);
				const options = q.options
					? [
							...q.options,
							{
								label: "Other",
								description: "Something else",
							},
						]
					: undefined;

				if (q.multiSelect && options) {
					// Multi-select: array of enums
					properties[q.header] = {
						title: q.header,
						description: q.question,
						type: "array",
						items: {
							anyOf: options.map((o) => {
								return {
									title: o.description,
									const: o.label,
								};
							}),
						},
					};
				} else if (options) {
					// Single select: enum
					properties[q.header] = {
						type: "string",
						title: q.header,
						description: q.question,
						oneOf: options.map((o) => {
							return {
								title: o.description,
								const: o.label,
							};
						}),
					};
				} else {
					properties[q.header] = {
						type: "string",
						title: q.header,
						description: q.question,
					};
				}
			}

			const jsonSchema: ElicitRequestFormParams["requestedSchema"] = {
				type: "object",
				properties,
				required,
			};

			// Send elicitation
			const result = await ctx.elicit(message, jsonSchema);

			if (!result.content) {
				return toolResult.error("No answer received from user");
			}

			// Map option IDs back to labels
			const answerText = Object.entries(result.content)
				.map(([key, value]) => {
					const questionIndex = questions.findIndex(
						(q, _i) => q.header === key,
					);

					if (questionIndex === -1) return `${key}: ${value}`;

					if (Array.isArray(value)) {
						return `${key}: ${value.join(", ")}`;
					} else {
						return `${key}: ${value}`;
					}
				})
				.join("\n");

			return toolResult.structured(
				`User answers:\n\n${answerText}`,
				result.content,
			);
		} catch (error) {
			const errorMessage =
				error instanceof Error ? error.message : String(error);
			return toolResult.error(`Error asking user question: ${errorMessage}`);
		}
	},
});
