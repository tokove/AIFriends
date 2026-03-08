<script setup>
import MicIcon from "@/components/character/icons/MicIcon.vue";
import SendIcon from "@/components/character/icons/SendIcon.vue";
import {ref, useTemplateRef} from "vue";
import streamApi from "@/js/http/streamApi.js";

const props = defineProps(['friendId'])
const emit = defineEmits(['pushBackMessage', 'addToLastMessage'])
const inputRef = useTemplateRef('input-ref')
const message = ref('')
let isProcessing = false

function focus() {
  inputRef.value.focus()
}

async function handleSend() {
  if (isProcessing) return
  isProcessing = true

  const content = message.value.trim()
  if (!content) return
  message.value = ""

  emit('pushBackMessage', {role: 'user', content: content, id: crypto.randomUUID()})
  emit('pushBackMessage', {role: 'ai', content: '', id: crypto.randomUUID()})

  try {
    await streamApi('/api/friend/message/chat/', {
      body: {
        friend_id: props.friendId,
        message: content,
      },
      onmessage(data, isDone) {
        if (isDone) {
          isProcessing = false
        } else if (data.content) {
          emit('addToLastMessage', data.content)
        }
      },
      onerror(err) {
        isProcessing = false
        console.log(err)
      }
    })
  } catch (err) {
    isProcessing = false
    console.log(err)
  }
}

defineExpose({
  focus,
})
</script>

<template>
  <form @submit.prevent="handleSend" class="absolute w-92 h-12 left-2 right-2 bottom-4 flex items-center">
    <input
        ref="input-ref"
        v-model="message"
        class="input bg-black/30 backdrop-blur-sm text-white text-base w-full h-full rounded-2xl pr-20"
        type="text"
        placeholder="心里的话，尽情说"
    >
    <div class="absolute right-10 w-8 h-8 flex justify-center items-center cursor-pointer">
      <MicIcon />
    </div>
    <div @click="handleSend" class="absolute right-2 w-8 h-8 flex justify-center items-center cursor-pointer">
      <SendIcon />
    </div>
  </form>
</template>

<style scoped>

</style>