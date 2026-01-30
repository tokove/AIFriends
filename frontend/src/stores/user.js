import {defineStore} from "pinia";
import {ref} from "vue";


export const useUserStore = defineStore('user', () => {
    const id = ref(1)
    const username = ref('zzh')
    const photo = ref('https://cdn.acwing.com/media/article/image/2026/01/30/535111_a1c749a5fd-default_at.jpg')
    const profile = ref('e')
    const accessToken = ref('e')

    function isLogin() {
        return !!accessToken.value // 必须带value
    }

    function setAccessToken(token) {
        accessToken.value = token
    }

    function setUserInfo(data) {
        id.value = data.user_id
        username.value = data.username
        photo.value = data.photo
        profile.value = data.profile
    }

    function logout() {
        id.value = 0
        username.value = ''
        photo.value = ''
        profile.value = ''
        accessToken.value = ''
    }

    return {
        id,
        username,
        photo,
        profile,
        accessToken,
        setAccessToken,
        setUserInfo,
        isLogin,
        logout,
    }
})