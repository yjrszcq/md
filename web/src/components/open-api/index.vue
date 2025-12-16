<template>
  <iframe class="open-api_component" ref="openApiRef" :src="hostUrl + '/api.html?' + current + sharpUrl"></iframe>
</template>

<script lang="ts" setup>
import { ref, watch, onMounted, onBeforeUnmount, onBeforeMount } from "vue";

const props = defineProps({
  content: {
    type: String,
    default: "",
  },
  mixUrl: {
    type: Boolean,
    default: false,
  },
});

const hostUrl = ref(location.origin);
const openApiRef = ref();
const sharpUrl = ref("");
const current = ref(new Date().getTime());

onBeforeMount(() => {
  if (props.mixUrl) {
    let pathArr = window.location.href.split("#");
    if (pathArr.length > 2) {
      sharpUrl.value = "/#" + pathArr[2];
    }
  }
});

onMounted(() => {
  openApiRef.value.onload = () => {
    postMessage({ type: "content", val: props.content });
  };
  if (props.mixUrl) {
    window.addEventListener("message", receiveMessage, false);
  }
});

onBeforeUnmount(() => {
  if (props.mixUrl) {
    window.removeEventListener("message", receiveMessage, false);
  }
});

watch(
  () => props.content,
  (val) => {
    postMessage({ type: "content", val: val });
  }
);

/**
 * 向iframe发送消息
 * @param val
 */
const postMessage = (val: object) => {
  openApiRef.value?.contentWindow?.postMessage(JSON.stringify(val), hostUrl.value);
};

/**
 * 接收iframe消息
 * @param e
 */
const receiveMessage = (e: MessageEvent) => {
  // Verify message origin
  if (e.origin !== hostUrl.value) {
    return;
  }

  try {
    const data = JSON.parse(e.data);
    if (data.type === "url" && typeof data.val === "string") {
      const pathArr = window.location.href.split("#");
      let path = pathArr[0] + "#" + pathArr[1];
      if (pathArr.length >= 2) {
        if (!path.endsWith("/")) {
          path += "/";
        }
        path = path + "#" + data.val;
        window.location.href = path;
      }
    }
  } catch {
    // Ignore invalid messages
  }
};
</script>

<style lang="scss" scoped>
.open-api_component {
  border: none;
  width: 100%;
  height: 100%;
}
</style>
