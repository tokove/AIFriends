<script setup>
import {ref, watch} from "vue";

const props = defineProps(['profile'])
const myProfile = ref(props.profile)

watch(() => props.profile, newValue => {
  myProfile.value = newValue
})

defineExpose({
  myProfile,
})
</script>

<template>
  <div class="flex justify-center">
     <fieldset class="fieldset">
      <label class="label text-base">简介</label>
      <textarea v-model="myProfile" rows="6" class="textarea w-98 resize-none"></textarea>
    </fieldset>
   </div>
</template>

<style scoped>

</style>