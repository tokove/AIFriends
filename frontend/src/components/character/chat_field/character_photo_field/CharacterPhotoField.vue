<script setup>
defineProps(['character'])
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
  </div>
</template>

<style scoped>

</style>