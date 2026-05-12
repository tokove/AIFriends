<script setup>
import {computed, nextTick, ref, useTemplateRef} from "vue";
import InputField from "@/components/character/chat_field/input_field/InputField.vue";
import MicroPhone from "@/components/character/chat_field/input_field/MicroPhone.vue";
import CharacterPhotoField from "@/components/character/chat_field/character_photo_field/CharacterPhotoField.vue";
import ChatHistory from "@/components/character/chat_field/chat_history/ChatHistory.vue";
import {stopPlayback} from "@/js/audio/playbackCoordinator.js";

const props = defineProps(['friend'])
const modalRef = useTemplateRef('modal-ref')
const inputRef = useTemplateRef('input-ref')
const chatHistoryRef = useTemplateRef('chat-history-ref')
const history = ref([])
const isVoiceMode = ref(false)
const enableTts = ref(true)
let lastScrollTask = null

async function showModal() {
  modalRef.value.showModal()

  await nextTick()
  inputRef.value.focus()
}

function handleClose() {
  stopPlayback()
  inputRef.value.close()
  if (lastScrollTask) {
    clearTimeout(lastScrollTask)
    lastScrollTask = null
  }
  chatHistoryRef.value?.stopAudio()
  isVoiceMode.value = false
}

function handlePushBackMessage(msg) {
  history.value.push(msg)
  chatHistoryRef.value.scrollToBottom()
}

function handleAddToLastMessage(delta) {
  if (!delta) return
  history.value.at(-1).content += delta
  if (lastScrollTask) return
  lastScrollTask = setTimeout(() => {
    chatHistoryRef.value.scrollToBottom()
    lastScrollTask = null
  }, 0)
}

function handleBindLastAIMessageId(messageId) {
  const lastMessage = history.value.at(-1)
  if (!lastMessage) return
  lastMessage.messageId = messageId
  lastMessage.id = `ai-${messageId}`
}

function handlePushFrontMessage(msg) {
  history.value.unshift(msg)
}

function handleToggleVoice() {
  stopPlayback()
  inputRef.value?.close()
  chatHistoryRef.value?.stopAudio()
  isVoiceMode.value = !isVoiceMode.value
}

function toggleTts() {
  enableTts.value = !enableTts.value
  stopPlayback()
  inputRef.value?.close()
}

async function handleVoiceSend(payload) {
  await inputRef.value.sendMessage(payload?.text?.trim() || '', {
    type: 'voice',
    audioUrl: payload?.audioUrl || '',
    asrText: payload?.text?.trim() || '',
    durationMs: payload?.durationMs || 0,
  })
}

const modalStyle = computed(() => {
  if (props.friend) {
    return {
      backgroundImage: `url(${props.friend.character.background_image})`,
      backgroundSize: 'cover',
      backgroundPosition: 'center',
      backgroundRepeat: 'no-repeat',
    }
  } else {
    return {}
  }
})

defineExpose({
  showModal,
})
</script>

<template>
  <dialog ref="modal-ref" class="modal" @close="handleClose">
    <div class="modal-box w-96 h-160" :style="modalStyle">
      <button @click="modalRef.close()" class="btn btn-sm btn-circle btn-ghost bg-transparent absolute right-3 top-3">✕</button>
      <ChatHistory
          ref="chat-history-ref"
          v-if="friend"
          :history="history"
          :friendId="friend.id"
          :character="friend.character"
          @pushFrontMessage="handlePushFrontMessage"
      />
      <InputField
          v-if="friend"
          ref="input-ref"
          v-show="!isVoiceMode"
          :friendId="friend.id"
          :enable-tts="enableTts"
          @pushBackMessage="handlePushBackMessage"
          @addToLastMessage="handleAddToLastMessage"
          @bindLastAIMessageId="handleBindLastAIMessageId"
          @toggleVoice="handleToggleVoice"
      />
      <MicroPhone
          v-if="friend && isVoiceMode"
          @close="handleToggleVoice"
          @send="handleVoiceSend"
          @stop="inputRef.close()"
      />
      <CharacterPhotoField
          v-if="friend"
          :character="friend.character"
          :enable-tts="enableTts"
          @toggleTts="toggleTts"
      />
    </div>
  </dialog>
</template>

<style scoped>

</style>
