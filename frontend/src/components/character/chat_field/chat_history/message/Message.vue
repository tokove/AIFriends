<script setup>
import {computed, onBeforeUnmount, ref} from "vue";
import api from "@/js/http/api.js";
import {useUserStore} from "@/stores/user.js";
import {claimPlayback, releasePlayback} from "@/js/audio/playbackCoordinator.js";

const props = defineProps(['message', 'character', 'friendId'])
const user = useUserStore()
const isPlaying = ref(false)
let currentAudio = null
let currentAudioUrl = ''
let playbackRequestId = 0

const durationLabel = computed(() => {
  const duration = Number(props.message?.durationMs || 0)
  if (!duration) return '语音'
  return `${Math.max(1, Math.round(duration / 1000))}''`
})

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
        <div class="chat-bubble whitespace-pre-wrap break-all">{{ message.content }}</div>
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
