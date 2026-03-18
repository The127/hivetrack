<!--
  MarkdownContent — renders a markdown string as sanitized HTML.

  Uses marked for parsing, DOMPurify for XSS sanitization,
  and highlight.js for syntax-highlighted code blocks.

  @prop {String} content — markdown source string
-->
<script setup>
import { computed } from 'vue'
import DOMPurify from 'dompurify'
import { md } from '@/lib/markdown'

const props = defineProps({
  content: { type: String, default: '' },
})

const html = computed(() => DOMPurify.sanitize(md.parse(props.content ?? '')))
</script>

<template>
  <div class="prose prose-sm prose-slate max-w-none" v-html="html" />
</template>
