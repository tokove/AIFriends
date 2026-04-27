<script setup>
defineProps(['character', 'enableTts'])
defineEmits(['toggleTts'])
</script>

<template>
  <div class="absolute left-3 right-22 top-3 h-12 w-fit bg-black/50 rounded-full flex justify-center items-center gap-3 px-2">
    <div class="avatar">
      <div class="w-10 rounded-full">
        <img :src="character.photo" alt="">
      </div>
    </div>
    <div class="text-base text-white line-clamp-1 break-all">
      {{ character.name }}
    </div>
    <button
        class="flex h-8 w-8 items-center justify-center rounded-full text-white/90 transition-colors hover:bg-white/10"
        @click="$emit('toggleTts')"
        aria-label="toggle realtime tts"
    >
      <svg v-if="enableTts" width="24" height="24" viewBox="0 0 24 24" fill="none"
           xmlns="http://www.w3.org/2000/svg">
        <path d="M11 5L6 9H2V15H6L11 19V5Z"
              stroke="currentColor" stroke-width="2" stroke-linejoin="round"/>
        <path d="M15 9C16.2 10.2 16.2 13.8 15 15"
              stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
        <path d="M17.5 6.5C20.5 9.5 20.5 14.5 17.5 17.5"
              stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
      </svg>
      <svg v-else width="24" height="24" viewBox="0 0 24 24" fill="none"
           xmlns="http://www.w3.org/2000/svg">
        <path d="M11 5L6 9H2V15H6L11 19V5Z"
              stroke="currentColor" stroke-width="2" stroke-linejoin="round"/>
        <path d="M16 9L21 14"
              stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
        <path d="M21 9L16 14"
              stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
      </svg>
    </button>
  </div>
</template>

<style scoped>

</style>
