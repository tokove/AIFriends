<script setup>
import {nextTick, onBeforeMount, onBeforeUnmount, ref, useTemplateRef, watch} from "vue";
import CemeraIcon from "@/views/user/profile/icons/CemeraIcon.vue";
import Croppie from 'croppie'
import 'croppie/croppie.css'


const props = defineProps(['photo'])
const myPhoto = ref(props.photo)

watch(() => props.photo, newValue => {
  myPhoto.value = newValue
})

const fileInputRef = useTemplateRef('file-input-ref')
const modalRef = useTemplateRef('modal-ref')
const croppieRef = useTemplateRef('croppie-ref')
let croppie = null

async function openModal(photo) {
  modalRef.value.show()
  await nextTick()

  if (!croppie) {
    croppie = new Croppie(croppieRef.value, {  // 创建croppie对象
      viewport: {width: 200, height: 200, type: 'square'},
      boundary: {width: 300, height: 300},
      enableOrientation: true,
      enforceBoundary: true,
    })
  }

  croppie.bind({
    url: photo,
  })
}

async function crop() {
  if (!croppie) return

  myPhoto.value = await croppie.result({
    type: 'base64',
    size: 'viewport',
  })

  modalRef.value.close()
}

function onFileChange(e) {
  const file = e.target.files[0]
  e.target.value = ''
  if (!file) return

  const reader = new FileReader()
  reader.onload = () => {
    openModal(reader.result)
  }

  reader.readAsDataURL(file)
}

onBeforeUnmount(() => {
  croppie?.destroy()
})

defineExpose({
  myPhoto,
})
</script>

<template>
  <div class="flex justify-center">
    <div class="avatar relative">
      <div class="w-28 rounded-full">
        <img :src="myPhoto" alt="">
      </div>
      <div @click="fileInputRef.click()" class="absolute left-0 top-0 w-28 h-28 flex justify-center items-center bg-black/20 rounded-full cursor-pointer">
        <CemeraIcon />
      </div>

      <input ref="file-input-ref" type="file" class="hidden" accept="image/*" @change="onFileChange">
      <dialog ref="modal-ref" class="modal">
        <div class="modal-box transition-none">
          <button @click="modalRef.close()" class="btn btn-circle btn-ghost btn-sm absolute right-2 top-2">✕</button>
          <div ref="croppie-ref" class="flex flex-col justify-center my-4"></div>

          <div class="modal-action">
            <button @click="modalRef.close()" class="btn">取消</button>
            <button @click="crop" class="btn btn-neutral">确定</button>
          </div>
        </div>
      </dialog>
    </div>
  </div>
</template>

<style scoped>

</style>