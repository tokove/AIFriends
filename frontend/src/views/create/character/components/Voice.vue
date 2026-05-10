<script setup>
import {computed, ref, watch} from "vue";

const props = defineProps(['voices', 'curVoiceId'])
const getDefaultVoiceId = () => props.voices?.[0]?.id ?? ''
const myVoice = ref(props.curVoiceId || getDefaultVoiceId())
const dropdownTriggerRef = ref(null)

watch(() => props.curVoiceId, newValue => {
  myVoice.value = newValue || getDefaultVoiceId()
})

const currentVoiceName = computed(() => {
  return props.voices?.find(voice => voice.id === myVoice.value)?.name || props.voices?.[0]?.name || ''
})

function handleSelect(voiceId) {
  myVoice.value = voiceId
  document.activeElement?.blur?.()
  dropdownTriggerRef.value?.blur()
}

defineExpose({
  myVoice,
})
</script>

<template>
  <div class="flex justify-center">
    <fieldset class="fieldset w-98">
      <label class="label text-base">音色</label>
      <div class="dropdown w-full">
        <div
            ref="dropdownTriggerRef"
            tabindex="0"
            role="button"
            class="btn h-auto min-h-0 w-full justify-between border-base-300 bg-base-100 px-4 py-3 font-normal text-base-content shadow-none"
        >
          {{ currentVoiceName }}
        </div>
        <ul
            tabindex="-1"
            class="dropdown-content menu z-1 mt-2 w-full rounded-box border border-base-300 bg-base-100 p-2 shadow-sm"
        >
          <li
              v-for="voice in voices"
              :key="voice.id || 'default'"
          >
            <button
                type="button"
                class="flex w-full items-center justify-between rounded-xl px-3 py-2 text-left text-sm font-normal text-base-content transition-colors"
                :class="myVoice === voice.id ? 'bg-base-300' : 'hover:bg-base-200'"
                @click="handleSelect(voice.id)"
            >
              <span>{{ voice.name }}</span>
              <span
                  class="flex h-5 w-5 items-center justify-center text-sm font-bold transition-opacity"
                  :class="myVoice === voice.id ? 'opacity-100' : 'opacity-0'"
              >
                ✓
              </span>
            </button>
          </li>
        </ul>
      </div>
    </fieldset>
  </div>
</template>
