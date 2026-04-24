<script setup>
import {ref} from "vue";
import api from "@/js/http/api.js";
import {useUserStore} from "@/stores/user.js";
import {useRouter} from "vue-router";
import {formRules, validatePassword, validateUsername} from "@/js/utils/validators.js";

const username = ref('')
const password = ref('')
const errorMessage = ref('')

const user = useUserStore()
const router = useRouter()

async function handleLogin() {
  errorMessage.value = ''
  const usernameError = validateUsername(username.value)
  const passwordError = validatePassword(password.value)

  if (usernameError) {
    errorMessage.value = usernameError
  } else if (passwordError) {
    errorMessage.value = passwordError
  } else {
    try {
      const res = await api.post('/api/user/account/login', {
        username: username.value.trim(),
        password: password.value
      })
      const data = res.data
      if (data.result === 'success') {
        user.setAccessToken(data.access)
        user.setUserInfo(data)
        await router.push({
          name: 'homepage-index'
        })
      } else {
        errorMessage.value = data.result
      }
    } catch (err) {
      errorMessage.value = err?.response?.data?.result || err?.message || '登录失败，请稍后重试'
    }
  }
}
</script>

<template>
  <div class="flex justify-center mt-30">
    <form @submit.prevent="handleLogin" class="fieldset bg-base-200 border-base-300 rounded-box w-xs border p-4">
      <label class="label">用户名</label>
      <input v-model.trim="username" type="text" class="input" placeholder="用户名" :maxlength="formRules.usernameMaxLength" />

      <label class="label">密码</label>
      <input v-model="password" type="password" class="input" placeholder="密码" :maxlength="formRules.passwordMaxLength" />

      <p v-if="errorMessage" class="text-sm text-red-500 mt-1">{{ errorMessage }}</p>

      <button class="btn btn-neutral mt-4">登录</button>
      <div class="flex justify-end">
        <RouterLink :to="{name: 'user-account-register-index'}" class="btn btn-sm btn-ghost">
          注册
        </RouterLink>
      </div>
    </form>
  </div>
</template>

<style scoped>

</style>
