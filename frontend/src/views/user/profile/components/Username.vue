<script setup>
import {ref, watch} from "vue";

const props = defineProps(['username'])
const myUsername = ref(props.username)

watch(() => props.username, newValue => {
  myUsername.value = newValue
})

defineExpose({
  myUsername,
})
</script>

<template>
 <div class="flex justify-center">
   <fieldset class="fieldset">
    <label class="label text-base ">用户名</label>
    <input v-model="myUsername" type="text" class="input w-98" placeholder="用户名" />
  </fieldset>
 </div>
</template>

<style scoped>

</style>