<script setup>
import KeyboardIcon from "@/components/character/icons/KeyboardIcon.vue";
import {onBeforeUnmount, onMounted, ref} from "vue";
import {MicVAD} from "@ricky0123/vad-web";
import api from "@/js/http/api.js";
import CONFIG_API from "@/js/config/config.js";

const emit = defineEmits(['close', 'send', 'stop'])
const isSpeaking = ref(false)
let vadInstance = null;
const AUDIO_SAMPLE_RATE = 16000

const startRecording = async () => {
  const baseUrl = CONFIG_API.VAD_URL;
  try {
    vadInstance = await MicVAD.new({
      baseAssetPath: baseUrl,
      onnxWASMBasePath: baseUrl,
      onSpeechStart: () => {
        isSpeaking.value = true;
        emit("stop")
      },
      onSpeechEnd: (audio) => {
        isSpeaking.value = false;
        sendToBackend(audio);
      },
      ortConfig: (ort) => {
        ort.env.logLevel = "error";
      },
      positiveSpeechThreshold: 0.8,
      negativeSpeechThreshold: 0.65,
      minSpeechFrames: 5,
      redemptionFrames: 5,
    });

    await vadInstance.start();
  } catch (e) {
    console.error('[vad] start failed', e)
  }
};
const float32ToInt16 = (float32Array) => {
  const buffer = new Int16Array(float32Array.length);
  for (let i = 0; i < float32Array.length; i++) {
    let s = Math.max(-1, Math.min(1, float32Array[i]));
    buffer[i] = s < 0 ? s * 0x8000 : s * 0x7fff;
  }
  return buffer;
}

const encodeWav = (pcm16Array, sampleRate) => {
  const dataSize = pcm16Array.length * 2
  const buffer = new ArrayBuffer(44 + dataSize)
  const view = new DataView(buffer)
  const writeString = (offset, value) => {
    for (let i = 0; i < value.length; i++) {
      view.setUint8(offset + i, value.charCodeAt(i))
    }
  }

  writeString(0, 'RIFF')
  view.setUint32(4, 36 + dataSize, true)
  writeString(8, 'WAVE')
  writeString(12, 'fmt ')
  view.setUint32(16, 16, true)
  view.setUint16(20, 1, true)
  view.setUint16(22, 1, true)
  view.setUint32(24, sampleRate, true)
  view.setUint32(28, sampleRate * 2, true)
  view.setUint16(32, 2, true)
  view.setUint16(34, 16, true)
  writeString(36, 'data')
  view.setUint32(40, dataSize, true)

  let offset = 44
  for (let i = 0; i < pcm16Array.length; i++, offset += 2) {
    view.setInt16(offset, pcm16Array[i], true)
  }
  return new Blob([buffer], { type: 'audio/wav' })
}

const sendToBackend = async (float32Audio) => {
  const pcm16 = float32ToInt16(float32Audio)
  const pcmBlob = new Blob([pcm16.buffer], { type: "audio/pcm" })
  const wavBlob = encodeWav(pcm16, AUDIO_SAMPLE_RATE)
  const durationMs = Math.round(float32Audio.length / AUDIO_SAMPLE_RATE * 1000)
  const formData = new FormData()
  formData.append("audio", pcmBlob, "voice.pcm")
  formData.append("display_audio", wavBlob, "voice.wav")
  formData.append("duration_ms", String(durationMs))
  try {
    const res = await api.post("/api/friend/message/asr", formData)
    const data = res.data
    if (data.result === "success") {
      emit("send", {
        text: data.text,
        audioUrl: data.audio_url || URL.createObjectURL(wavBlob),
        durationMs: data.duration_ms || durationMs,
      })
    }
  } catch (err) {
    console.error(err)
  }
}

onMounted(() => {
  startRecording()
})

onBeforeUnmount(() => {
  if (vadInstance) {
    vadInstance.destroy()
    vadInstance = null
  }
})
</script>

<template>
  <div class="absolute w-92 h-12 left-2 right-2 bottom-4 flex items-center bg-black/30 backdrop-blur-sm rounded-2xl">
    <div v-if="isSpeaking" class="flex items-center justify-center gap-1 h-6 flex-1">
      <div
        v-for="i in 32" :key="i"
        class="w-0.5 bg-blue-400 rounded-full animate-wave"
        :style="{ animationDelay: `${i * 0.1}s` }"
      ></div>
    </div>
    <div v-else class="text-white/50 text-base w-full text-center"> 心里的话，尽情说</div>
    <div @click="emit('close')" class="absolute right-2 w-8 h-8 flex justify-center items-center cursor-pointer">
      <KeyboardIcon />
    </div>
  </div>
</template>

<style scoped>
.animate-wave {
  height: 4px;
  animation: wave-animation 0.6s ease-in-out infinite alternate;
}

@keyframes wave-animation {
  0% { height: 4px; opacity: 0.3; }
  100% { height: 20px; opacity: 1; }
}
</style>
