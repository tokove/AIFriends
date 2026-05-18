<script setup>
import {computed, onBeforeUnmount, ref} from "vue";
import MarkdownIt from "markdown-it";
import hljs from "highlight.js/lib/core";
import bash from "highlight.js/lib/languages/bash";
import css from "highlight.js/lib/languages/css";
import go from "highlight.js/lib/languages/go";
import javascript from "highlight.js/lib/languages/javascript";
import json from "highlight.js/lib/languages/json";
import markdown from "highlight.js/lib/languages/markdown";
import python from "highlight.js/lib/languages/python";
import xml from "highlight.js/lib/languages/xml";
import "highlight.js/styles/github.css";
import api from "@/js/http/api.js";
import {useUserStore} from "@/stores/user.js";
import {claimPlayback, releasePlayback} from "@/js/audio/playbackCoordinator.js";

const props = defineProps(['message', 'character', 'friendId'])
const user = useUserStore()
const isPlaying = ref(false)
let currentAudio = null
let currentAudioUrl = ''
let playbackRequestId = 0
let codeBlockIdSeed = 0

hljs.registerLanguage('bash', bash)
hljs.registerLanguage('css', css)
hljs.registerLanguage('go', go)
hljs.registerLanguage('javascript', javascript)
hljs.registerLanguage('js', javascript)
hljs.registerLanguage('json', json)
hljs.registerLanguage('markdown', markdown)
hljs.registerLanguage('md', markdown)
hljs.registerLanguage('python', python)
hljs.registerLanguage('py', python)
hljs.registerLanguage('html', xml)
hljs.registerLanguage('xml', xml)

const markdownRenderer = new MarkdownIt({
  html: false,
  linkify: true,
  breaks: true,
  typographer: true,
  highlight(code, language) {
    const normalizedLanguage = language?.trim().toLowerCase()
    if (normalizedLanguage && hljs.getLanguage(normalizedLanguage)) {
      return hljs.highlight(code, {language: normalizedLanguage, ignoreIllegals: true}).value
    }
    return hljs.highlightAuto(code).value
  },
})

markdownRenderer.renderer.rules.fence = (tokens, index, options, env, self) => {
  const token = tokens[index]
  const language = token.info ? token.info.trim().split(/\s+/)[0] : ''
  const codeId = `code-block-${codeBlockIdSeed += 1}`
  const highlighted = token.content
    ? markdownRenderer.options.highlight?.(token.content, language, '', env) || ''
    : ''

  return `<div class="code-block" data-code-block-id="${codeId}">
    <button type="button" class="code-block-copy" data-code-block-copy="${codeId}" aria-label="复制代码块">复制</button>
    <pre><code class="hljs${language ? ` language-${language}` : ''}">${highlighted || self.escapeHtml(token.content)}</code></pre>
  </div>`
}

const defaultLinkRender = markdownRenderer.renderer.rules.link_open || ((tokens, index, options, env, self) => {
  return self.renderToken(tokens, index, options)
})

markdownRenderer.renderer.rules.link_open = (tokens, index, options, env, self) => {
  const token = tokens[index]
  token.attrSet('target', '_blank')
  token.attrSet('rel', 'noopener noreferrer')
  return defaultLinkRender(tokens, index, options, env, self)
}

const durationLabel = computed(() => {
  const duration = Number(props.message?.durationMs || 0)
  if (!duration) return '语音'
  return `${Math.max(1, Math.round(duration / 1000))}''`
})

const renderedContent = computed(() => {
  return markdownRenderer.render(props.message?.content || '')
})

function handleMarkdownClick(event) {
  const target = event.target
  if (!(target instanceof HTMLElement)) return

  const copyButton = target.closest('[data-code-block-copy]')
  if (!copyButton) return

  const codeBlock = copyButton.closest('[data-code-block-id]')
  const codeElement = codeBlock?.querySelector('code')
  const codeText = codeElement?.textContent || ''
  if (!codeText) return

  navigator.clipboard?.writeText(codeText).then(() => {
    const originalLabel = copyButton.textContent || '复制'
    copyButton.textContent = '已复制'
    copyButton.classList.add('is-copied')
    window.setTimeout(() => {
      copyButton.textContent = originalLabel
      copyButton.classList.remove('is-copied')
    }, 1200)
  }).catch(() => {})
}

function playUserAudio() {
  if (!props.message?.audioUrl) return
  if (isPlaying.value) {
    stopAudio()
    return
  }
  claimPlayback(stopAudio)
  currentAudio = new Audio(props.message.audioUrl)
  currentAudio.onended = () => {
    stopAudio()
  }
  currentAudio.onpause = () => {
    if (currentAudio) {
      isPlaying.value = false
    }
  }
  isPlaying.value = true
  currentAudio.play().catch(() => {
    stopAudio()
  })
}

function stopAudio() {
  playbackRequestId += 1
  if (currentAudio) {
    currentAudio.pause()
    currentAudio.currentTime = 0
    currentAudio = null
  }
  if (currentAudioUrl) {
    URL.revokeObjectURL(currentAudioUrl)
    currentAudioUrl = ''
  }
  isPlaying.value = false
  releasePlayback(stopAudio)
}

async function playAIAudio() {
  if (!props.message?.messageId) return
  if (isPlaying.value) {
    stopAudio()
    return
  }

  const requestId = playbackRequestId + 1
  playbackRequestId = requestId
  isPlaying.value = true
  claimPlayback(stopAudio)
  try {
    const res = await api.post('/api/friend/message/tts', {
      friend_id: props.friendId,
      message_id: props.message.messageId,
    }, {
      responseType: 'blob'
    })
    if (requestId !== playbackRequestId) return

    const url = URL.createObjectURL(res.data)
    currentAudioUrl = url
    const audio = new Audio(url)
    currentAudio = audio
    audio.onended = () => {
      stopAudio()
    }
    audio.onerror = () => {
      stopAudio()
    }
    audio.onpause = () => {
      if (currentAudio === audio) {
        isPlaying.value = false
      }
    }
    await audio.play()
  } catch {
    if (requestId === playbackRequestId) {
      stopAudio()
    }
  }
}

onBeforeUnmount(() => {
  stopAudio()
})

defineExpose({
  stopAudio,
})
</script>

<template>
  <div v-if="message.content">
    <div v-if="message.role === 'ai'" class="chat chat-start">
      <div class="chat-image avatar">
        <div class="w-10 rounded-full">
          <img :src="character.photo"  alt=""/>
        </div>
      </div>
      <div class="group flex items-start gap-1">
        <div class="chat-bubble markdown-message" v-html="renderedContent" @click="handleMarkdownClick"></div>
        <button
            class="mt-1 flex h-5 min-w-8 shrink-0 items-center justify-center rounded-md bg-white/92 px-1.5 text-slate-600 opacity-0 shadow-sm ring-1 ring-black/5 transition-all group-hover:opacity-100 hover:bg-white"
            :class="{ 'opacity-100': isPlaying }"
            @click="playAIAudio"
        >
          <span v-if="isPlaying" class="wave-container" aria-hidden="true">
            <span class="bar"></span>
            <span class="bar"></span>
            <span class="bar"></span>
            <span class="bar"></span>
          </span>
          <svg v-else viewBox="0 0 24 24" class="h-3.5 w-3.5" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
            <path d="M11 5L6 9H2V15H6L11 19V5Z"
                  stroke="currentColor" stroke-width="2" stroke-linejoin="round"/>
            <path d="M15 9C16.2 10.2 16.2 13.8 15 15"
                  stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
            <path d="M17.5 6.5C20.5 9.5 20.5 14.5 17.5 17.5"
                  stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
          </svg>
        </button>
      </div>
    </div>
    <div v-else-if="message.type === 'voice'" class="message-stack">
      <div class="chat chat-end">
        <div class="chat-image avatar">
          <div class="w-10 rounded-full">
            <img :src="user.photo" alt="" />
          </div>
        </div>
        <button class="chat-bubble chat-bubble-success flex items-center gap-3 min-w-28 justify-between" @click="playUserAudio">
          <span>语音</span>
          <span>{{ durationLabel }}</span>
        </button>
      </div>
    </div>
    <div v-else class="chat chat-end">
      <div class="chat-image avatar">
        <div class="w-10 rounded-full">
          <img :src="user.photo" alt="" />
        </div>
      </div>
      <div class="chat-bubble chat-bubble-success whitespace-pre-wrap break-all">{{ message.content }}</div>
    </div>
  </div>
</template>

<style scoped>
.message-stack {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.markdown-message {
  max-width: min(18rem, calc(100vw - 7rem));
  overflow-wrap: anywhere;
  word-break: normal;
}

.markdown-message :deep(p) {
  margin: 0.25rem 0;
}

.markdown-message :deep(p:first-child) {
  margin-top: 0;
}

.markdown-message :deep(p:last-child) {
  margin-bottom: 0;
}

.markdown-message :deep(ul),
.markdown-message :deep(ol) {
  margin: 0.35rem 0;
  padding-left: 1.25rem;
}

.markdown-message :deep(ul) {
  list-style: disc;
}

.markdown-message :deep(ol) {
  list-style: decimal;
}

.markdown-message :deep(li + li) {
  margin-top: 0.2rem;
}

.markdown-message :deep(a) {
  color: #2563eb;
  text-decoration: underline;
  text-underline-offset: 2px;
}

.markdown-message :deep(blockquote) {
  margin: 0.4rem 0;
  border-left: 3px solid rgba(71, 85, 105, 0.35);
  padding-left: 0.75rem;
  color: #475569;
}

.markdown-message :deep(code) {
  border-radius: 0.25rem;
  background: rgba(15, 23, 42, 0.08);
  padding: 0.08rem 0.25rem;
  font-size: 0.9em;
  word-break: break-word;
}

.markdown-message :deep(.code-block) {
  position: relative;
  margin: 0.5rem 0;
}

.markdown-message :deep(.code-block pre) {
  max-width: 100%;
  overflow-x: auto;
  border: 1px solid rgba(71, 85, 105, 0.18);
  border-radius: 0.45rem;
  background: #f8fafc;
  padding: 1rem 0.85rem 0.8rem;
  color: #0f172a;
}

.markdown-message :deep(.code-block code) {
  display: block;
  background: transparent;
  padding: 0;
  color: inherit;
  font-size: 0.82rem;
  line-height: 1.45;
  white-space: pre;
  word-break: normal;
}

.markdown-message :deep(.code-block-copy) {
  position: absolute;
  right: 0.4rem;
  top: 0.35rem;
  z-index: 1;
  border-radius: 0.3rem;
  border: 1px solid rgba(71, 85, 105, 0.18);
  background: rgba(255, 255, 255, 0.94);
  padding: 0.15rem 0.45rem;
  font-size: 0.72rem;
  line-height: 1.2;
  color: #334155;
}

.markdown-message :deep(.code-block-copy:hover) {
  background: #ffffff;
}

.markdown-message :deep(.code-block-copy.is-copied) {
  border-color: rgba(34, 197, 94, 0.35);
  color: #15803d;
}

.markdown-message :deep(table) {
  display: block;
  max-width: 100%;
  overflow-x: auto;
  border-collapse: collapse;
  margin: 0.5rem 0;
}

.markdown-message :deep(th),
.markdown-message :deep(td) {
  border: 1px solid rgba(71, 85, 105, 0.22);
  padding: 0.25rem 0.45rem;
  text-align: left;
}

.markdown-message :deep(hr) {
  margin: 0.6rem 0;
  border: 0;
  border-top: 1px solid rgba(71, 85, 105, 0.22);
}

.wave-container {
  display: flex;
  align-items: center;
  gap: 2px;
}

.bar {
  width: 2px;
  height: 7px;
  background: #007bff;
  border-radius: 999px;
  animation: jump 0.9s infinite ease-in-out;
}

.bar:nth-child(1) {
  animation-delay: 0s;
}

.bar:nth-child(2) {
  animation-delay: 0.15s;
}

.bar:nth-child(3) {
  animation-delay: 0.3s;
}

.bar:nth-child(4) {
  animation-delay: 0.45s;
}

@keyframes jump {
  0%, 100% {
    height: 5px;
  }
  50% {
    height: 13px;
  }
}
</style>
