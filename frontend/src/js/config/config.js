const platform = 'vue' // vue, gin, cloud

const CONFIG_API = {
    HTTP_URL: '',
    VAD_URL: '',
}

if (platform === 'vue') {
    CONFIG_API.HTTP_URL = ''
    CONFIG_API.VAD_URL = '/vad/'
} else if (platform === 'gin') {
    CONFIG_API.HTTP_URL = 'http://127.0.0.1:8080'
    CONFIG_API.VAD_URL = 'http://127.0.0.1:8080/static/frontend/vad/'
} else if (platform === 'cloud') {
    CONFIG_API.HTTP_URL = 'https://app7727.acapp.acwing.com.cn'
    CONFIG_API.VAD_URL = 'https://app7727.acapp.acwing.com.cn/static/frontend/vad/'
}

export default CONFIG_API