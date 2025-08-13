import { marked } from 'marked';
import hljs from 'highlight.js/lib/core';
import javascript from 'highlight.js/lib/languages/javascript';
import typescript from 'highlight.js/lib/languages/typescript';
import python from 'highlight.js/lib/languages/python';
import json from 'highlight.js/lib/languages/json';
import bash from 'highlight.js/lib/languages/bash';
import css from 'highlight.js/lib/languages/css';
import xml from 'highlight.js/lib/languages/xml'; // for HTML

// Register languages
hljs.registerLanguage('javascript', javascript);
hljs.registerLanguage('typescript', typescript);
hljs.registerLanguage('python', python);
hljs.registerLanguage('json', json);
hljs.registerLanguage('bash', bash);
hljs.registerLanguage('css', css);
hljs.registerLanguage('html', xml);
hljs.registerLanguage('xml', xml);

// Configure marked with syntax highlighting
marked.setOptions({
	highlight: function (code, lang) {
		if (lang && hljs.getLanguage(lang)) {
			try {
				const result = hljs.highlight(code, { language: lang });
				return result.value;
			} catch (err) {
				console.warn('Syntax highlighting failed for', lang, ':', err);
			}
		} else {
			console.warn('Language not found:', lang, 'Available:', hljs.listLanguages());
		}

		const autoResult = hljs.highlightAuto(code);
		return autoResult.value;
	},
	breaks: true,
	gfm: true,
	langPrefix: 'hljs language-' // Ensure proper CSS classes
});

export function renderMarkdown(content: string): string {
	const result = marked(content) as string;
	return result;
}
