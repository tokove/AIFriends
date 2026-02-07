<script setup>
import {ref, watch} from "vue";

const props = defineProps(['name'])
const myName = ref(props.name)

watch(() => props.name, newValue => {
  myName.value = newValue
})

defineExpose({
  myName,
})
</script>

<template>
  <div class="flex justify-center">
    <fieldset class="fieldset">
      <label class="label text-base">名字</label>
      <input v-model="myName" type="text" class="input w-98" placeholder="名字">
    </fieldset>
  </div>
</template>

<style scoped>

</style>