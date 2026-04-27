<script setup>
import SendIcon from "@/components/character/icons/SendIcon.vue";
import MicIcon from "@/components/character/icons/MicIcon.vue";
import {ref, useTemplateRef} from "vue";
import streamApi from "@/js/http/streamApi.js";

const props = defineProps(['friendId'])
const emit = defineEmits(['pushBackMessage', 'addToLastMessage', 'toggleVoice'])
const inputRef = useTemplateRef('input-ref')
const message = ref('')
let processId = 0
const STREAM_FLUSH_INTERVAL = 32
let streamBuffer = ''
let flushTimer = null
let currentAudio = null
let currentAudioUrl = ''
let mediaSource = null
let sourceBuffer = null
let sourceBufferQueue = []
let audioStreamFinished = false

function clearFlushTimer() {
  if (flushTimer) {
    clearTimeout(flushTimer)
    flushTimer = null
  }
}

function flushBufferedStream() {
  if (!streamBuffer) return
  emit('addToLastMessage', streamBuffer)
  streamBuffer = ''
}

function scheduleStreamFlush() {
  if (flushTimer) return
  flushTimer = setTimeout(() => {
    flushBufferedStream()
    flushTimer = null
  }, STREAM_FLUSH_INTERVAL)
}

function focus() {
  inputRef.value.focus()
}

function stopAudio() {
  sourceBufferQueue = []
  audioStreamFinished = false
  sourceBuffer = null
  mediaSource = null
  if (currentAudio) {
    currentAudio.pause()
    currentAudio = null
  }
  if (currentAudioUrl) {
    URL.revokeObjectURL(currentAudioUrl)
    currentAudioUrl = ''
  }
}

function decodeBase64Audio(base64) {
  const binary = atob(base64)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i)
  }
  return bytes
}

function flushSourceBuffer() {
  if (!sourceBuffer || sourceBuffer.updating) {
    return
  }

  if (sourceBufferQueue.length > 0) {
    sourceBuffer.appendBuffer(sourceBufferQueue.shift())
    return
  }

  if (audioStreamFinished && mediaSource && mediaSource.readyState === 'open') {
    try {
      mediaSource.endOfStream()
    } catch (err) {
    }
  }
}

function ensureAudioStream() {
  if (mediaSource || typeof MediaSource === 'undefined') {
    return
  }

  mediaSource = new MediaSource()
  currentAudioUrl = URL.createObjectURL(mediaSource)
  currentAudio = new Audio(currentAudioUrl)

  mediaSource.addEventListener('sourceopen', () => {
    if (!mediaSource || mediaSource.readyState !== 'open' || sourceBuffer) {
      return
    }

    sourceBuffer = mediaSource.addSourceBuffer('audio/mpeg')
    sourceBuffer.mode = 'sequence'
    sourceBuffer.addEventListener('updateend', flushSourceBuffer)
    flushSourceBuffer()
  }, { once: true })

  currentAudio.play().catch(() => {
  })
}

function enqueueAudioChunk(base64) {
  if (!base64) return

  console.debug('[tts] received audio chunk')
  ensureAudioStream()
  sourceBufferQueue.push(decodeBase64Audio(base64))
  flushSourceBuffer()
}

function finishAudioStream() {
  audioStreamFinished = true
  flushSourceBuffer()
}

async function sendMessage(content) {
  if (!content) return

  const curId = ++ processId

  stopAudio()
  ensureAudioStream()
  clearFlushTimer()
  streamBuffer = ''

  message.value = ""

  emit('pushBackMessage', {role: 'user', content: content, id: crypto.randomUUID()})
  emit('pushBackMessage', {role: 'ai', content: '', id: crypto.randomUUID()})

  try {
    await streamApi('/api/friend/message/chat/', {
      body: {
        friend_id: props.friendId,
        message: content,
      },
      onmessage(data, isDone) {
        if (curId !== processId) return

        if (isDone) {
          flushBufferedStream()
          finishAudioStream()
          return
        }

        if (data.content) {
          streamBuffer += data.content
          scheduleStreamFlush()
        }

        if (data.audio) {
          enqueueAudioChunk(data.audio)
        }
      },
      onerror(err) {
      }
    })
  } catch (err) {
  } finally {
    if (curId === processId) {
      clearFlushTimer()
      flushBufferedStream()
    }
  }
}

async function handleSend() {
  const content = message.value.trim()
  message.value = ""
  await sendMessage(content)
}

function close() {
  ++ processId
  stopAudio()
  clearFlushTimer()
  streamBuffer = ''
}

defineExpose({
  focus,
  close,
  sendMessage,
})
</script>

<template>
  <form @submit.prevent="handleSend" class="absolute w-92 h-12 left-2 right-2 bottom-4 flex items-center">
    <input
        ref="input-ref"
        v-model="message"
        class="input bg-black/30 backdrop-blur-sm text-white text-base w-full h-full rounded-2xl pr-22"
        type="text"
        placeholder="心里的话，尽情说"
    >
    <div @click="emit('toggleVoice')" class="absolute right-12 w-8 h-8 flex justify-center items-center cursor-pointer">
      <MicIcon />
    </div>
    <div @click="handleSend" class="absolute right-2 w-8 h-8 flex justify-center items-center cursor-pointer">
      <SendIcon />
    </div>
  </form>
</template>

<style scoped>

</style>
