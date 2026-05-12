<script setup>
import {ref, watch} from "vue";

const props = defineProps(['voices', 'curVoiceId'])
const getVoiceId = voice => voice?.voice_id ?? voice?.id ?? ''
const getDefaultVoiceId = () => getVoiceId(props.voices?.[0])
const myVoice = ref(props.curVoiceId || getDefaultVoiceId())

watch(() => props.curVoiceId, newValue => {
  myVoice.value = newValue || getDefaultVoiceId()
})

watch(() => props.voices, () => {
  if (!myVoice.value) {
    myVoice.value = getDefaultVoiceId()
  }
})

defineExpose({
  myVoice,
})
</script>

<template>
  <div class="flex justify-center">
    <fieldset class="fieldset">
      <label class="label text-base">音色</label>
      <select
          v-model="myVoice"
          class="select appearance-none w-98"
      >
        <option disabled value="">请选择音色</option>
        <option
            v-for="voice in voices"
            :key="getVoiceId(voice) || 'default'"
            :value="getVoiceId(voice)"
        >
          {{ voice.name }}
        </option>
      </select>
    </fieldset>
  </div>
</template>
