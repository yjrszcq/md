<template>
  <div class="page-open-document">
    <md-preview v-if="docType === 'md'" class="md-view" :content="content" />
    <open-api-preview v-if="docType === 'openApi'" :content="content" mixUrl></open-api-preview>
    <div v-if="docType === 'error'" class="error-view">文档加载失败</div>
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted } from "vue";
import MdPreview from "@/components/md-editor/preview";
import OpenApiPreview from "@/components/open-api/index.vue";
import { useRoute } from "vue-router";
import OpenApi from "@/api/open";

const content = ref("");
const docType = ref("");

onMounted(() => {
  OpenApi.getDoc(String(useRoute().query.id))
    .then((res) => {
      content.value = res.data.content;
      docType.value = res.data.type!;
    })
    .catch((err) => {
      console.error(err);
      docType.value = "error";
    });
});
</script>

<style lang="scss">
.page-open-document {
  height: 100%;
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background-color: var(--bg-secondary);
  .md-view {
    height: 100%;
    width: 100%;
    overflow: auto;
  }
  .error-view {
    font-size: 16px;
    color: var(--text-secondary);
    user-select: none;
  }
}
@media (max-width: 720px) {
  .page-open-document {
    .md-view {
      .catalog-view {
        display: none;
      }
    }
  }
}
</style>
