<script setup>
import {computed, ref, watch} from "vue";

const props = defineProps(['voices', 'curVoiceId'])
const myVoice = ref(props.curVoiceId ?? '')
const isOpen = ref(false)

watch(() => props.curVoiceId, newValue => {
  myVoice.value = newValue ?? ''
})

const currentVoiceName = computed(() => {
  return props.voices?.find(voice => voice.id === myVoice.value)?.name || '默认音色（男音）'
})

function handleSelect(voiceId) {
  myVoice.value = voiceId
  isOpen.value = false
}

defineExpose({
  myVoice,
})
</script>

<template>
  <div class="flex justify-center">
    <fieldset class="fieldset">
      <label class="label text-base">音色</label>
      <div class="relative w-98">
        <button
            type="button"
            class="w-full flex items-center justify-between rounded-2xl border border-base-300 bg-base-100 px-4 py-3 text-sm text-left hover:bg-base-200 transition-colors"
            @click="isOpen = !isOpen"
        >
          <span>{{ currentVoiceName }}</span>
          <span class="text-base-content/60 transition-transform" :class="isOpen ? 'rotate-180' : ''">⌄</span>
        </button>

        <div
            v-if="isOpen"
            class="absolute left-0 right-0 top-[calc(100%+8px)] z-10 flex flex-col gap-2 rounded-2xl border border-base-300 bg-base-100 p-2 shadow-lg"
        >
          <button
              v-for="voice in voices"
              :key="voice.id || 'default'"
              type="button"
              class="flex items-center justify-between rounded-2xl px-4 py-3 text-sm transition-colors"
              :class="myVoice === voice.id ? 'bg-base-300' : 'hover:bg-base-200'"
              @click="handleSelect(voice.id)"
          >
            <span>{{ voice.name }}</span>
            <span
                class="w-5 h-5 rounded-full border-2 flex items-center justify-center transition-colors"
                :class="myVoice === voice.id ? 'border-neutral bg-neutral' : 'border-base-content/25 bg-transparent'"
            >
              <span
                  class="w-2.5 h-2.5 rounded-full bg-white transition-opacity"
                  :class="myVoice === voice.id ? 'opacity-100' : 'opacity-0'"
              ></span>
            </span>
          </button>
        </div>
      </div>
    </fieldset>
  </div>
</template>

<style scoped>

</style>
