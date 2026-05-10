<script setup>
import SendIcon from "@/components/character/icons/SendIcon.vue";
import MicIcon from "@/components/character/icons/MicIcon.vue";
import {ref, useTemplateRef} from "vue";
import streamApi from "@/js/http/streamApi.js";

const props = defineProps(['friendId', 'enableTts'])
const emit = defineEmits(['pushBackMessage', 'addToLastMessage', 'bindLastAIMessageId', 'toggleVoice'])
const inputRef = useTemplateRef('input-ref')
const message = ref('')
let processId = 0
let audioContext = null
let audioWorkletNode = null
let nextAudioTime = 0
const pcmSampleRate = 24000

function focus() {
  inputRef.value.focus()
}

function stopAudio() {
  if (audioContext) {
    audioContext.close().catch(() => {})
    audioContext = null
  }
  audioWorkletNode = null
  nextAudioTime = 0
}

function decodeBase64PCM(base64) {
  const binary = atob(base64)
  const sampleCount = Math.floor(binary.length / 2)
  const samples = new Float32Array(sampleCount)
  for (let i = 0; i < sampleCount; i++) {
    const lo = binary.charCodeAt(i * 2)
    const hi = binary.charCodeAt(i * 2 + 1)
    let value = (hi << 8) | lo
    if (value >= 0x8000) value -= 0x10000
    samples[i] = value / 0x8000
  }
  return samples
}

function ensureAudioContext() {
  if (audioContext) {
    return
  }
  audioContext = new AudioContext({ sampleRate: pcmSampleRate })
  nextAudioTime = audioContext.currentTime
  audioContext.audioWorklet.addModule('/pcm-player-worklet.js')
      .then(() => {
        if (!audioContext) return
        audioWorkletNode = new AudioWorkletNode(audioContext, 'pcm-player')
        audioWorkletNode.connect(audioContext.destination)
      })
      .catch(() => {
        audioWorkletNode = null
      })
}

function enqueuePCMChunk(base64) {
  if (!base64) return

  ensureAudioContext()
  const samples = decodeBase64PCM(base64)
  if (!samples.length) return

  if (audioWorkletNode) {
    audioWorkletNode.port.postMessage({ type: 'pcm', samples: samples.buffer }, [samples.buffer])
    return
  }

  const buffer = audioContext.createBuffer(1, samples.length, pcmSampleRate)
  buffer.copyToChannel(samples, 0)

  const source = audioContext.createBufferSource()
  source.buffer = buffer
  source.connect(audioContext.destination)

  const startTime = Math.max(nextAudioTime, audioContext.currentTime + 0.02)
  source.start(startTime)
  nextAudioTime = startTime + buffer.duration
}

async function sendMessage(content, messageMeta = null) {
  if (!content) return

  const curId = ++ processId

  stopAudio()
  if (props.enableTts) {
    ensureAudioContext()
  }

  message.value = ""

  emit('pushBackMessage', {
    role: 'user',
    type: messageMeta?.type || 'text',
    content: content,
    asrText: messageMeta?.asrText || '',
    audioUrl: messageMeta?.audioUrl || '',
    durationMs: messageMeta?.durationMs || 0,
    id: crypto.randomUUID()
  })
  emit('pushBackMessage', {role: 'ai', type: 'text', content: '', id: crypto.randomUUID(), messageId: 0})

  try {
    const body = {
      friend_id: props.friendId,
      message: content,
      user_message_type: messageMeta?.type || 'text',
      user_audio: messageMeta?.audioUrl || '',
      user_asr_text: messageMeta?.asrText || '',
      user_audio_duration_ms: messageMeta?.durationMs || 0,
      enable_tts: !!props.enableTts,
    }

    await streamApi('/api/friend/message/chat/', {
      body,
      onmessage(data, isDone) {
        if (curId !== processId) return

        if (isDone) {
          return
        }

        if (data.content) {
          emit('addToLastMessage', data.content)
        }

        if (data.message_id) {
          emit('bindLastAIMessageId', data.message_id)
        }

        if (props.enableTts && data.audio) {
          enqueuePCMChunk(data.audio)
        }
      },
      onerror(err) {
      }
    })
  } catch (err) {
  } finally {
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
