<script setup>
import SendIcon from "@/components/character/icons/SendIcon.vue";
import {ref, useTemplateRef} from "vue";
import streamApi from "@/js/http/streamApi.js";

const props = defineProps(['friendId'])
const emit = defineEmits(['pushBackMessage', 'addToLastMessage'])
const inputRef = useTemplateRef('input-ref')
const message = ref('')
let processId = 0
const STREAM_FLUSH_INTERVAL = 32
let streamBuffer = ''
let flushTimer = null

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

async function handleSend() {
  const content = message.value.trim()

  if (!content) return

  const curId = ++ processId

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
          return
        }

        if (data.content) {
          streamBuffer += data.content
          scheduleStreamFlush()
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

function close() {
  ++ processId
  clearFlushTimer()
  streamBuffer = ''
}

defineExpose({
  focus,
  close,
})
</script>

<template>
  <form @submit.prevent="handleSend" class="absolute w-92 h-12 left-2 right-2 bottom-4 flex items-center">
    <input
        ref="input-ref"
        v-model="message"
        class="input bg-black/30 backdrop-blur-sm text-white text-base w-full h-full rounded-2xl pr-12"
        type="text"
        placeholder="心里的话，尽情说"
    >
    <div @click="handleSend" class="absolute right-2 w-8 h-8 flex justify-center items-center cursor-pointer">
      <SendIcon />
    </div>
  </form>
</template>

<style scoped>

</style>