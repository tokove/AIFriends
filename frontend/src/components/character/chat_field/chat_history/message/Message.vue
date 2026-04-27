<script setup>
import {computed, ref} from "vue";
import api from "@/js/http/api.js";
import {useUserStore} from "@/stores/user.js";

const props = defineProps(['message', 'character', 'friendId'])
const user = useUserStore()
const isPlaying = ref(false)
let currentAudio = null

const durationLabel = computed(() => {
  const duration = Number(props.message?.durationMs || 0)
  if (!duration) return '语音'
  return `${Math.max(1, Math.round(duration / 1000))}''`
})

function playUserAudio() {
  if (!props.message?.audioUrl) return
  if (currentAudio) {
    currentAudio.pause()
  }
  currentAudio = new Audio(props.message.audioUrl)
  currentAudio.play().catch(() => {})
}

async function playAIAudio() {
  if (!props.message?.content) return
  if (isPlaying.value) {
    currentAudio?.pause()
    isPlaying.value = false
    return
  }

  isPlaying.value = true
  try {
    const res = await api.post('/api/friend/message/tts', {
      friend_id: props.friendId,
      text: props.message.content,
    }, {
      responseType: 'blob'
    })
    const url = URL.createObjectURL(res.data)
    if (currentAudio) {
      currentAudio.pause()
    }
    currentAudio = new Audio(url)
    currentAudio.onended = () => {
      URL.revokeObjectURL(url)
      isPlaying.value = false
    }
    currentAudio.onerror = () => {
      URL.revokeObjectURL(url)
      isPlaying.value = false
    }
    currentAudio.onpause = () => {
      isPlaying.value = false
    }
    await currentAudio.play()
  } catch {
    isPlaying.value = false
  }
}
</script>

<template>
  <div v-if="message.content">
    <div v-if="message.role === 'ai'" class="chat chat-start">
      <div class="chat-image avatar">
        <div class="w-10 rounded-full">
          <img :src="character.photo"  alt=""/>
        </div>
      </div>
      <div class="group relative max-w-[80%]">
        <div class="chat-bubble whitespace-pre-wrap break-all">{{ message.content }}</div>
        <button
            class="absolute -right-3 top-0 flex h-7 w-7 items-center justify-center rounded-full bg-white/92 text-slate-600 shadow-sm ring-1 ring-black/5 opacity-0 transition-all group-hover:opacity-100 hover:bg-white"
            :class="{ 'opacity-100': isPlaying }"
            @click="playAIAudio"
        >
          <svg viewBox="0 0 24 24" class="h-4 w-4" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
            <path d="M11 5L6 9H2V15H6L11 19V5Z"
                  stroke="currentColor" stroke-width="2" stroke-linejoin="round"/>
            <path d="M15 9C16.2 10.2 16.2 13.8 15 15"
                  stroke="currentColor" stroke-width="2" stroke-linecap="round" :class="{ 'wave-pulse': isPlaying }"/>
            <path d="M17.5 6.5C20.5 9.5 20.5 14.5 17.5 17.5"
                  stroke="currentColor" stroke-width="2" stroke-linecap="round" :class="{ 'wave-pulse delay-150': isPlaying }"/>
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

.wave-pulse {
  animation: wavePulse 0.9s ease-in-out infinite;
}

.delay-150 {
  animation-delay: 0.15s;
}

@keyframes wavePulse {
  0%, 100% { opacity: 0.25; }
  50% { opacity: 1; }
}
</style>
