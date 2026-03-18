import { Marked } from 'marked'
import hljs from 'highlight.js'

const renderer = {
  code({ text, lang }) {
    const language = lang && hljs.getLanguage(lang) ? lang : 'plaintext'
    const highlighted = hljs.highlight(text, { language }).value
    return `<pre><code class="hljs language-${language}">${highlighted}</code></pre>`
  },
}

export const md = new Marked({ renderer })
