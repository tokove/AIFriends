<script setup>
import NavBar from "@/components/navbar/NavBar.vue";
import {onMounted} from "vue";
import {useUserStore} from "@/stores/user.js";
import api from "@/js/http/api.js";
import {useRoute, useRouter} from "vue-router";

const user = useUserStore()
const route = useRoute()
const router = useRouter()

async function restoreAccessToken() {
  if (user.accessToken) {
    return true
  }

  const res = await api.post('/api/user/account/refresh_token', {})
  const data = res.data
  if (data?.result !== 'success' || !data?.access) {
    return false
  }

  user.setAccessToken(data.access)
  return true
}

async function loadUserInfo() {
  const res = await api.get('/api/user/account/get_user_info/')
  const data = res.data
  if (data?.result === 'success') {
    user.setUserInfo(data)
    return true
  }
  return false
}

onMounted(async () => {
  try {
    const restored = await restoreAccessToken()
    if (restored) {
      await loadUserInfo()
    }
  } catch(err) {
    user.logout()
  } finally {
    user.setHasPulledUserInfo(true)

    if (route.meta.needLogin && !user.isLogin()) {
      await router.replace({
        name: 'user-account-login-index'
      })
    }
  }
})
</script>

<template>
  <NavBar>
    <RouterView />
  </NavBar>
</template>

<style scoped>

</style>
