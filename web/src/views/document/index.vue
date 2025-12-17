<template>
  <div class="page-document">
    <book @change="bookChange" @books="booksFetch" :onlyPreview="onlyPreview" :isStretch="isStretch" :loading="mdLoading"></book>
    <doc
      :onlyPreview="onlyPreview"
      :isStretch="isStretch"
      :currentBookId="currentBookId"
      :currentDoc="currentDoc"
      :books="books"
      @change="docChange"
      @loading="loadingChange"
      ref="docRef"
    ></doc>
    <div class="editor-container">
      <div class="codemirror-view" v-if="docType === 'openApi'">
        <open-api v-if="onlyPreview" :content="currentDoc.content"></open-api>
        <template v-else>
          <div class="codemirror-toolbar">
            <div class="icon-outer" title="保存" @click="saveDoc(currentDoc.content)">
              <svg-icon name="save" className="icon-save"></svg-icon>
            </div>
            <div class="icon-outer" title="导出" @click="exportOpenApi(currentDoc.name, currentDoc.content)">
              <svg-icon name="download" className="icon-download"></svg-icon>
            </div>
          </div>
          <div class="codemirror-inner">
            <codemirror-editor
              :style="{ visibility: codemirrorVisibility }"
              ref="codemirrorRef"
              v-model="currentDoc.content"
              :disabled="onlyPreview || mdLoading"
              noRadius
              @save="saveDoc(currentDoc.content)"
              @ready="codemirrorReady"
            />
          </div>
        </template>
      </div>
      <template v-else>
        <md-preview v-if="onlyPreview" :key="'preview' + mdKey" class="editor-view" :content="currentDoc.content" />
        <md-editor
          v-else
          ref="mdEditorRef"
          :key="'editor' + mdKey"
          class="editor-view"
          v-model="currentDoc.content"
          v-loading="mdLoading"
          @onUploadImg="uploadImage"
          @onSave="saveDoc"
          @export="exportMarkdown(currentDoc.name, currentDoc.content)"
          @aiToggle="toggleAiSidebar"
        />
      </template>
      <ai-sidebar
        :visible="aiSidebarVisible"
        :docTitle="currentDoc.name"
        :docContent="currentDoc.content"
        @close="aiSidebarVisible = false"
        @openConfig="openAiConfig"
      />
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, Ref, onMounted, onBeforeUnmount, nextTick, computed, watch } from "vue";
import { ElMessage } from "element-plus";
import MdEditor from "@/components/md-editor";
import MdPreview from "@/components/md-editor/preview";
import CodemirrorEditor from "@/components/codemirror-editor";
import OpenApi from "@/components/open-api/index.vue";
import SvgIcon from "@/components/svg-icon";
import AiSidebar from "@/components/ai-sidebar/index.vue";
import { uploadPicture } from "../picture/util";
import Book from "./components/book.vue";
import Doc from "./components/doc.vue";
import DocCache from "@/store/doc-cache";
import Token from "@/store/token";
import { host } from "@/config";
import crypto from "crypto-js";
import { exportMarkdown, exportOpenApi } from "./util";

defineProps({
  onlyPreview: {
    type: Boolean,
    default: true,
  },
  isStretch: {
    type: Boolean,
    default: true,
  },
});

const docRef = ref<InstanceType<typeof Doc>>();
const codemirrorRef = ref();
const mdEditorRef = ref();
const hostUrl = ref("");
const books: Ref<Book[]> = ref([]);
const currentBookId = ref("");
const currentDoc: Ref<CurrentDoc> = ref({
  id: "",
  name: "",
  content: "",
  originMD5: "",
  type: "",
  updateTime: "",
});
const mdLoading = ref(false);
const mdKey = ref(0);
const codemirrorVisibility = ref("hidden");
const aiSidebarVisible = ref(false);

const docType = computed(() => {
  return currentDoc.value.type;
});

watch(docType, (newVal, oldVal) => {
  if (oldVal && oldVal !== newVal && newVal === "openApi") {
    codemirrorVisibility.value = "hidden";
  }
});

onMounted(() => {
  hostUrl.value = process.env.NODE_ENV === "production" ? location.origin : host;
  DocCache.getDoc().then((res) => {
    if (res) {
      currentDoc.value = res;
    }
  });
});

onBeforeUnmount(() => {
  if (Token.getAccessToken()) {
    DocCache.setDoc(currentDoc.value);
  }
});

window.onbeforeunload = () => {
  if (Token.getAccessToken()) {
    DocCache.setDoc(currentDoc.value);
  }
};

/**
 * 文档loading变化
 */
const loadingChange = (val: boolean) => {
  mdLoading.value = val;
};

/**
 * 文集选择变化
 */
const bookChange = (bookId: string) => {
  currentBookId.value = bookId;
};

/**
 * 文集列表变化
 */
const booksFetch = (bookList: Book[]) => {
  books.value = bookList;
};

/**
 * 文档选择变化
 */
const docChange = (id: string, name: string, content: string, type: string, updateTime: string, noRender?: boolean) => {
  currentDoc.value.id = id;
  currentDoc.value.name = name;
  currentDoc.value.content = content;
  currentDoc.value.type = type;
  currentDoc.value.originMD5 = crypto.MD5(content).toString();
  currentDoc.value.updateTime = updateTime;
  if (!noRender) {
    mdKey.value++;
    nextTick(() => {
      if (codemirrorRef.value) {
        codemirrorRef.value.$el.getElementsByClassName("cm-scroller")[0].scrollTop = 0;
      }
    });
  }
  DocCache.setDoc(currentDoc.value);
};

/**
 * codemirror加载完成
 */
const codemirrorReady = () => {
  setTimeout(() => {
    if (codemirrorRef.value) {
      codemirrorRef.value.$el.getElementsByClassName("cm-scroller")[0].scrollTop = 0;
      codemirrorVisibility.value = "unset";
    }
  }, 100);
};

/**
 * 上传图片
 */
const uploadImage = async (files: File[], callback: (urls: string[]) => void) => {
  const pathList: string[] = [];
  for (let file of files) {
    try {
      pathList.push(hostUrl.value + (await uploadPicture(file)));
    } catch (e) {}
  }
  callback(pathList);
};

/**
 * 保存文档
 */
const saveDoc = (content: string) => {
  if (mdLoading.value) {
    return;
  }
  docRef.value?.saveDoc(content);
};

/**
 * 切换 AI 侧边栏
 */
const toggleAiSidebar = () => {
  aiSidebarVisible.value = !aiSidebarVisible.value;
};

/**
 * 打开 AI 配置（通过全局事件通知 layout）
 */
const openAiConfig = () => {
  window.dispatchEvent(new CustomEvent("open-ai-config"));
};
</script>

<style lang="scss">
.page-document {
  display: flex;
  overflow: auto;
  .editor-container {
    display: flex;
    flex: 1;
    height: 100%;
    overflow: hidden;
  }
  .editor-view {
    height: 100%;
    flex: 1;
    min-width: 720px;
  }
  .editor-view.md-fullscreen {
    min-width: unset;
  }
  .codemirror-view {
    height: 100%;
    flex: 1;
    min-width: 720px;
    overflow: hidden;
  }
  .codemirror-toolbar {
    height: 34px;
    display: flex;
    align-items: center;
    justify-content: flex-end;
    border: #e6e6e6 1px solid;
    border-bottom: none;
    padding-right: 10px;
    .icon-outer {
      width: 30px;
      height: 24px;
      color: #3f4a54;
      cursor: pointer;
      .icon-save {
        width: 16px;
        height: 16px;
        margin: 4px 7px;
      }
      .icon-download {
        width: 20px;
        height: 20px;
        margin: 2px 5px;
      }
    }
    .icon-outer:hover {
      background: #f6f6f6;
    }
  }
  .codemirror-inner {
    height: calc(100% - 35px);
    overflow: hidden;
  }
}
@media (max-width: 720px) {
  .page-document {
    .editor-view {
      min-width: 100%;
      .catalog-view {
        display: none;
      }
    }
    .codemirror-view {
      min-width: 100%;
    }
  }
}
</style>
