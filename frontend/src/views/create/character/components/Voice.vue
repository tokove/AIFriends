<script setup>
import {ref, watch} from "vue";

const props = defineProps(['voices', 'curVoiceId'])
const myVoice = ref(props.voice)

watch(() => props.name, newValue => {
  myVoice.value = newValue
})

defineExpose({
  myVoice,
})
</script>

<template>
  <div class="flex justify-center">
    <fieldset class="fieldset">
      <label class="label text-base">音色</label>
      <select v-model="myVoice" class="select">
        <option
            v-for="voice in voices"
            :id="voice.id"
            :value="voice.id"
        >{{ voice.name }}</option>
      </select>
    </fieldset>
  </div>
</template>

<style scoped>

</style>