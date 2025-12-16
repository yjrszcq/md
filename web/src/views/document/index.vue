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
              @ready="codemirrorReday"
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
          :showAiButton="aiConfig.panelEnabled"
          @onUploadImg="uploadImage"
          @onSave="saveDoc"
          @export="exporMarkdown(currentDoc.name, currentDoc.content)"
          @aiToggle="toggleAiSidebar"
        />
      </template>
      <ai-sidebar
        :visible="aiSidebarVisible"
        :docTitle="currentDoc.name"
        :docContent="currentDoc.content"
        @close="aiSidebarVisible = false"
        @applyChange="handleAiChange"
        @undoChange="handleAiUndo"
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
import AIConfigStore from "@/store/ai-config";
import Token from "@/store/token";
import { host } from "@/config";
import crypto from "crypto-js";
import { exporMarkdown, exportOpenApi } from "./util";

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
const aiConfig = ref<AIConfig>({
  baseUrl: "",
  apiKey: "",
  model: "",
  systemPrompts: [],
  currentPromptId: "",
  agentEnabled: false,
  docContextEnabled: false,
  panelEnabled: false,
});

const docType = computed(() => {
  return currentDoc.value.type;
});

watch(docType, (newVal, oldVal) => {
  if (oldVal && oldVal !== newVal && newVal === "openApi") {
    codemirrorVisibility.value = "hidden";
  }
});

// 加载 AI 配置
const loadAiConfig = async () => {
  aiConfig.value = await AIConfigStore.getConfig();
};

// 监听 AI 配置变化事件
const handleAiConfigChanged = (event: CustomEvent<AIConfig>) => {
  aiConfig.value = event.detail;
};

onMounted(() => {
  hostUrl.value = process.env.NODE_ENV === "production" ? location.origin : host;
  DocCache.getDoc().then((res) => {
    if (res) {
      currentDoc.value = res;
    }
  });
  loadAiConfig();
  window.addEventListener("ai-config-changed", handleAiConfigChanged as EventListener);
});

onBeforeUnmount(() => {
  if (Token.getAccessToken()) {
    DocCache.setDoc(currentDoc.value);
  }
  window.removeEventListener("ai-config-changed", handleAiConfigChanged as EventListener);
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
const codemirrorReday = () => {
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
 * Scroll editor to position and highlight lines
 */
let highlightStyleEl: HTMLStyleElement | null = null;

const scrollToAndSelect = (startIndex: number, endIndex: number) => {
  // Wait for editor to fully re-render after mdKey change
  setTimeout(() => {
    const content = currentDoc.value.content;

    // Boundary check
    const safeStart = Math.min(startIndex, content.length);
    const safeEnd = Math.min(endIndex, content.length);

    const beforeText = content.slice(0, safeStart);
    const totalLines = content.split('\n').length;
    // Clamp startLine to actual line count (handles EOF insert undo)
    const startLine = Math.min(beforeText.split('\n').length, totalLines);

    // Remove previous highlight style if exists
    if (highlightStyleEl) {
      highlightStyleEl.remove();
      highlightStyleEl = null;
    }

    // Apply highlight style first (survives image re-renders)
    const style = document.createElement('style');
    if (safeStart === safeEnd) {
      // Point indicator for insert undo / delete apply
      style.textContent = `
        .editor-view .cm-content .cm-line:nth-child(${startLine})::after {
          content: '';
          display: block;
          height: 3px;
          background: linear-gradient(90deg, #f0ad4e, #ec971f);
          margin-top: 2px;
          border-radius: 2px;
        }
      `;
    } else {
      // Normal range highlight
      const changeText = content.slice(safeStart, safeEnd);
      const lineCount = changeText.split('\n').length;
      const endLine = startLine + lineCount - 1;

      const selectors: string[] = [];
      for (let i = startLine; i <= endLine; i++) {
        selectors.push(`.editor-view .cm-content .cm-line:nth-child(${i})`);
      }
      style.textContent = `
        ${selectors.join(',\n')} {
          background-color: #fff3cd !important;
        }
      `;
    }
    document.head.appendChild(style);
    highlightStyleEl = style;

    // Scroll function - re-query elements as DOM may have changed
    const doScroll = () => {
      const freshEditorEl = document.querySelector('.editor-view .cm-content') as HTMLElement;
      if (!freshEditorEl) return;
      const freshLines = freshEditorEl.querySelectorAll('.cm-line');
      // If target line doesn't exist (e.g., insert at EOF undone), scroll to last line
      const targetIndex = Math.min(startLine - 1, freshLines.length - 1);
      const freshTarget = freshLines[targetIndex] as HTMLElement;
      if (freshTarget) {
        freshTarget.scrollIntoView({ behavior: 'smooth', block: 'center' });
      }
    };

    // Multiple scroll attempts to handle async rendering
    doScroll();
    setTimeout(doScroll, 300);
    setTimeout(doScroll, 600);

    // Remove highlight after delay
    setTimeout(() => {
      if (highlightStyleEl) {
        highlightStyleEl.remove();
        highlightStyleEl = null;
      }
    }, 2500);
  }, 100);
};

/**
 * 处理 AI 变更建议
 */
const handleAiChange = (
  change: AgentChange,
  callback: (success: boolean, undoData?: { originalContent: string; appliedContent: string; undoStart: number; undoEnd: number }) => void
) => {
  const originalContent = currentDoc.value.content;
  let success = false;
  let changeStart = -1;
  let changeEnd = -1;
  let undoStart = -1;
  let undoEnd = -1;

  switch (change.type) {
    case "replace":
      if (change.oldText) {
        const index = originalContent.indexOf(change.oldText);
        if (index !== -1) {
          currentDoc.value.content =
            originalContent.slice(0, index) + change.content + originalContent.slice(index + change.oldText.length);
          ElMessage.success("已替换指定内容");
          mdKey.value++;
          success = true;
          changeStart = index;
          changeEnd = index + change.content.length;
          // Undo position: where oldText was
          undoStart = index;
          undoEnd = index + change.oldText.length;
        } else {
          ElMessage.warning("未找到要替换的内容，请手动修改");
        }
      } else {
        ElMessage.warning("缺少原文定位信息，请手动修改");
      }
      break;

    case "insert":
      if (change.oldText) {
        // Insert after specified text
        const index = originalContent.indexOf(change.oldText);
        if (index !== -1) {
          const insertPos = index + change.oldText.length;
          currentDoc.value.content =
            originalContent.slice(0, insertPos) + change.content + originalContent.slice(insertPos);
          ElMessage.success("已插入内容");
          mdKey.value++;
          success = true;
          changeStart = insertPos;
          changeEnd = insertPos + change.content.length;
          // Undo: content removed, show line at insert position
          undoStart = insertPos;
          undoEnd = insertPos;
        } else {
          ElMessage.warning("未找到插入位置，请手动修改");
        }
      } else {
        // No anchor text, append to end
        const insertPos = originalContent.length + 1;
        currentDoc.value.content = originalContent + "\n" + change.content;
        ElMessage.success("已追加内容到文档末尾");
        mdKey.value++;
        success = true;
        changeStart = insertPos;
        changeEnd = currentDoc.value.content.length;
        // Undo: show line at original end position
        undoStart = originalContent.length;
        undoEnd = originalContent.length;
      }
      break;

    case "delete":
      if (change.oldText) {
        const index = originalContent.indexOf(change.oldText);
        if (index !== -1) {
          currentDoc.value.content =
            originalContent.slice(0, index) + originalContent.slice(index + change.oldText.length);
          ElMessage.success("已删除指定内容");
          mdKey.value++;
          success = true;
          changeStart = index;
          changeEnd = index;
          // Undo: oldText restored at index
          undoStart = index;
          undoEnd = index + change.oldText.length;
        } else {
          ElMessage.warning("未找到要删除的内容，请手动修改");
        }
      } else if (change.content) {
        // 使用 content 字段作为要删除的内容
        const index = originalContent.indexOf(change.content);
        if (index !== -1) {
          currentDoc.value.content =
            originalContent.slice(0, index) + originalContent.slice(index + change.content.length);
          ElMessage.success("已删除指定内容");
          mdKey.value++;
          success = true;
          changeStart = index;
          changeEnd = index;
          // Undo: content restored at index
          undoStart = index;
          undoEnd = index + change.content.length;
        } else {
          ElMessage.warning("未找到要删除的内容，请手动修改");
        }
      } else {
        ElMessage.warning("缺少删除定位信息，请手动修改");
      }
      break;

    default:
      ElMessage.warning("未知的修改类型，请手动修改");
  }

  // 回调通知结果
  if (success) {
    callback(true, {
      originalContent,
      appliedContent: currentDoc.value.content,
      undoStart,
      undoEnd,
    });
    // Scroll to change and highlight
    if (changeStart >= 0) {
      scrollToAndSelect(changeStart, changeEnd);
    }
  } else {
    callback(false);
  }
};

/**
 * 处理 AI 变更撤回
 */
const handleAiUndo = (change: AgentChange, callback: (success: boolean) => void) => {
  if (change.undoData) {
    // Check if content matches applied content
    if (currentDoc.value.content === change.undoData.appliedContent) {
      currentDoc.value.content = change.undoData.originalContent;
      // Save scroll position before callback clears undoData
      const scrollStart = change.undoData.undoStart;
      const scrollEnd = change.undoData.undoEnd;
      ElMessage.success("已撤回修改");
      mdKey.value++;
      callback(true);
      // Scroll to undo position
      if (scrollStart >= 0) {
        scrollToAndSelect(scrollStart, scrollEnd);
      }
    } else {
      ElMessage.warning("文档已被修改，无法撤回");
      callback(false);
    }
  } else {
    ElMessage.warning("无法撤回：缺少撤回数据");
    callback(false);
  }
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
