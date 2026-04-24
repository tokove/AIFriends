<script setup>
import {nextTick, onBeforeUnmount, onMounted, ref, useTemplateRef} from "vue";
import api from "@/js/http/api.js";
import Character from "@/components/character/Character.vue";

const friends = ref([])
const isLoading = ref(false)
const hasFriends = ref(true)
const sentinelRef = useTemplateRef('sentinel-ref')

function checkSentinelVisible() {  // 判断哨兵是否能被看到
  if (!sentinelRef.value) return false

  const rect = sentinelRef.value.getBoundingClientRect()
  return rect.top < window.innerHeight && rect.bottom > 0
}

async function loadMore() {
  if (isLoading.value || !hasFriends.value) return
  isLoading.value = true

  let newFriends = []
  try {
    const lastFriend = friends.value.at(-1)
    const res = await api.get('/api/friend/get_list/', {
      params: {
        cursor_updated_at: lastFriend?.updated_at,
        cursor_id: lastFriend?.id
      }
    })
    const data = res.data
    if (data.result === 'success') {
      newFriends = data.friends
    }
  } catch (err) {
  } finally {
    isLoading.value = false
    if (newFriends.length === 0) {
      hasFriends.value = false
    } else {
      const existingIds = new Set(friends.value.map(friend => friend.id))
      const uniqueFriends = newFriends.filter(friend => !existingIds.has(friend.id))
      friends.value.push(...uniqueFriends)
      await nextTick()

      if (uniqueFriends.length > 0 && checkSentinelVisible()) {
        await loadMore()
      } else if (uniqueFriends.length === 0) {
        hasFriends.value = false
      }
    }
  }
}

let observer = null
onMounted(async() => {
  await loadMore()
  observer = new IntersectionObserver(
      entries => {
        entries.forEach(entry => {
          if (entry.isIntersecting) {
            loadMore()
          }
        })
      },
      {root: null, rootMargin: '2px', threshold: 0}
  )
  observer.observe(sentinelRef.value)
})

function removeFriend(friendId) {
  friends.value = friends.value.filter(f => f.id !== friendId)
}

onBeforeUnmount(() => {
  observer?.disconnect()
})
</script>

<template>
  <div class="flex flex-col items-center mb-12">
    <div class="grid grid-cols-[repeat(auto-fill,minmax(240px,1fr))] gap-9 mt-12 justify-items-center w-full px-9">
      <Character
          v-for="friend in friends"
          :key="friend.id"
          :character="friend.character"
          :canRemoveFriend="true"
          :friendId="friend.id"
          @remove="removeFriend"
      />
    </div>
    <div ref="sentinel-ref" class="h-2 mt-8"></div>
    <div v-if="isLoading" class="text-gray-500 mt-4">加载中...</div>
    <div v-else-if="!hasFriends" class="text-gray-500 mt-4">没有更多好友了</div>
  </div>
</template>

<style scoped>

</style>
