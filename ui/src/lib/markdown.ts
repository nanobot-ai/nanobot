import { Marked } from 'marked';
import { markedHighlight } from 'marked-highlight';
import hljs from 'highlight.js';

const marked = new Marked(
	markedHighlight({
		emptyLangClass: 'hljs',
		langPrefix: 'hljs language-',
		highlight(code, lang) {
			const language = hljs.getLanguage(lang) ? lang : 'plaintext';
			return hljs.highlight(code, { language }).value;
		}
	})
);

marked.setOptions({
	breaks: true,
	gfm: true
});

marked.use({
	renderer: {
		link(arg) {
			const base = marked.Renderer.prototype.link.call(this, arg);
			return base.replace('<a', '<a target="_blank" rel="noopener noreferrer"');
		}
	}
});

export function renderMarkdown(content: string): string {
	if (!content) return '';
	return marked.parse(content) as string;
}
