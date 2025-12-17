<template>
  <div class="ai-sidebar" :class="{ hidden: !visible }" :style="{ width: sidebarWidth + 'px' }">
    <!-- 拖拽调整大小手柄 -->
    <div class="resize-handle" @mousedown="startResize"></div>

    <!-- 顶部工具栏 -->
    <div class="sidebar-header">
      <div class="header-left">
        <span class="model-name" :title="config.model">{{ config.model || "未配置模型" }}</span>
        <span class="status-dot" :class="statusClass"></span>
      </div>
      <div class="header-right">
        <div class="header-actions">
          <el-button text size="small" @click="exportHistory" title="导出对话" :disabled="tasks.length === 0 || !currentConversationId">
            <el-icon><Download /></el-icon>
          </el-button>
          <el-button text size="small" @click="deleteCurrentConversation" title="删除对话" :disabled="!currentConversationId">
            <el-icon><Delete /></el-icon>
          </el-button>
          <el-button text size="small" @click="emit('close')" title="关闭">
            <el-icon><Close /></el-icon>
          </el-button>
        </div>
      </div>
    </div>

    <!-- 对话选择栏 -->
    <div class="conversation-bar">
      <!-- 搜索按钮/搜索栏 -->
      <div class="search-area" ref="searchAreaRef">
        <template v-if="!searchExpanded">
          <el-button text size="small" @click="expandSearch" title="搜索对话">
            <el-icon><Search /></el-icon>
          </el-button>
        </template>
        <template v-else>
          <el-input
            ref="searchInputRef"
            v-model="searchKeyword"
            size="small"
            placeholder="搜索对话..."
            clearable
            @input="handleSearch"
            @clear="handleSearchClear"
            @keydown.esc="collapseSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          <el-button text size="small" @click="collapseSearch" title="取消搜索">
            <el-icon><Close /></el-icon>
          </el-button>
        </template>
      </div>
      <!-- 对话选择下拉列表 -->
      <el-select
        ref="conversationSelectRef"
        v-model="currentConversationId"
        size="small"
        placeholder="选择对话"
        class="conversation-select"
        @change="handleConversationChange"
        @visible-change="handleSelectVisibleChange"
        :disabled="sending"
      >
        <el-option label="+ 新建对话" value="" />
        <el-option
          v-for="conv in filteredConversations"
          :key="conv.id"
          :label="conv.title"
          :value="conv.id"
        >
          <div class="conversation-option" :class="{ 'is-current': conv.id === currentConversationId }">
            <span class="conv-title">{{ conv.title }}</span>
            <span class="conv-time">{{ formatTime(conv.updateTime) }}</span>
          </div>
        </el-option>
      </el-select>
    </div>

    <!-- 任务块区域 -->
    <div class="task-blocks" ref="taskBlocksRef" @scroll="handleTaskBlocksScroll">
      <div v-for="(task, taskIndex) in tasks" :key="task.id" class="task-block">
        <!-- 用户任务 -->
        <div class="user-task">
          <div class="task-label-row">
            <div class="task-label">Task</div>
            <el-button
              v-if="task.status !== 'processing'"
              text
              size="small"
              class="delete-task-btn"
              @click="deleteTask(taskIndex)"
              title="删除此对话"
            >
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>
          <div class="task-content">{{ task.userTask }}</div>
        </div>

        <!-- AI 推理过程（可折叠，默认收起，仅当有思考内容时显示） -->
        <div v-if="task.reasoning" class="reasoning-section">
          <div class="reasoning-header" @click="toggleReasoning(task.id)">
            <el-icon :class="{ 'is-expanded': expandedReasonings.has(task.id) }">
              <ArrowRight />
            </el-icon>
            <span class="reasoning-title">Thinking</span>
            <span class="reasoning-hint">点击展开/收起</span>
          </div>
          <div v-show="expandedReasonings.has(task.id)" class="reasoning-content">
            <pre>{{ task.reasoning }}</pre>
          </div>
        </div>

        <!-- AI 输出 -->
        <div class="ai-output" v-if="task.output || task.status === 'processing'">
          <div class="output-label-row">
            <div class="task-label">Output</div>
            <el-button
              v-if="task.output && task.status !== 'processing'"
              text
              size="small"
              class="copy-output-btn"
              @click="copyOutput(task.output)"
              title="复制回复"
            >
              <el-icon><CopyDocument /></el-icon>
            </el-button>
          </div>
          <div class="output-content">
            <MdPreview v-if="task.output" :modelValue="task.output" previewTheme="cyanosis" />
            <span v-else-if="task.status === 'processing'" class="cursor-blink">▋</span>
          </div>
        </div>

        <!-- 错误状态 -->
        <div v-if="task.status === 'error'" class="error-state">
          <el-icon><Warning /></el-icon>
          <span>{{ task.output }}</span>
        </div>
      </div>

      <el-empty v-if="tasks.length === 0" description="开始与 AI 对话" />
    </div>

    <!-- 底部输入区 -->
    <div class="input-area">
      <el-input
        v-model="inputText"
        type="textarea"
        :rows="3"
        :placeholder="inputPlaceholder"
        @keydown="handleKeydown"
        :disabled="sending"
      />
      <div class="input-actions">
        <div class="input-left">
          <span class="input-tip">Enter 发送，Shift+Enter 换行</span>
          <el-checkbox v-model="autoScrollUserEnabled" size="small" class="auto-scroll-checkbox">
            自动滚动
          </el-checkbox>
        </div>
        <el-button
          v-if="sending"
          type="danger"
          @click="stopGeneration"
        >
          <el-icon><VideoPause /></el-icon>
          停止
        </el-button>
        <el-button
          v-else
          type="primary"
          :disabled="!inputText.trim() || !config.model"
          @click="sendMessage"
        >
          发送
        </el-button>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, shallowRef, computed, watch, onMounted, onUnmounted, nextTick, triggerRef } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { Delete, Close, Warning, ArrowRight, VideoPause, Download, Search, CopyDocument } from "@element-plus/icons-vue";
import { MdPreview } from "md-editor-v3";
import "md-editor-v3/lib/preview.css";
import AIApi from "@/api/ai";
import AIConfigStore from "@/store/ai-config";

const props = defineProps<{
  visible: boolean;
  docTitle?: string;
  docContent?: string;
}>();

const emit = defineEmits<{
  (e: "close"): void;
}>();

const inputText = ref("");
const sending = ref(false);
const sidebarWidth = ref(380);
const tasks = shallowRef<TaskBlock[]>([]);
const taskBlocksRef = ref<HTMLElement | null>(null);
const expandedReasonings = ref<Set<string>>(new Set());
const abortController = ref<AbortController | null>(null);

// Auto-scroll state
const autoScrollEnabled = ref(true);
const autoScrollUserEnabled = ref(true);
const lastScrollTop = ref(0);

// 对话记录相关
const conversations = ref<AIConversationListItem[]>([]);
const currentConversationId = ref<string>("");
const searchExpanded = ref(false);
const searchKeyword = ref("");
const searchInputRef = ref<HTMLInputElement | null>(null);
const searchAreaRef = ref<HTMLElement | null>(null);
const conversationSelectRef = ref<InstanceType<typeof import("element-plus").ElSelect> | null>(null);
const isFirstAiResponse = ref(false);
const searchDebounceTimer = ref<ReturnType<typeof setTimeout> | null>(null);
const savedConversationId = ref<string>(""); // 保存搜索前的对话ID

const config = ref<AIConfig>({
  baseUrl: "",
  apiKey: "",
  model: "",
  systemPrompts: [],
  currentPromptId: "",
  systemPromptEnabled: false,
  docContextEnabled: false,
  syncEnabled: false,
});

// 过滤后的对话列表
const filteredConversations = computed(() => {
  if (!searchKeyword.value) {
    return conversations.value;
  }
  const keyword = searchKeyword.value.toLowerCase();
  return conversations.value.filter((c) => c.title.toLowerCase().includes(keyword));
});

// 格式化时间
const formatTime = (timestamp: number) => {
  const date = new Date(timestamp);
  const now = new Date();
  const diff = now.getTime() - date.getTime();

  if (diff < 60000) return "刚刚";
  if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`;
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`;
  if (diff < 604800000) return `${Math.floor(diff / 86400000)}天前`;

  return `${date.getMonth() + 1}/${date.getDate()}`;
};

// 加载配置
const loadConfig = async () => {
  config.value = await AIConfigStore.getConfig();
  sidebarWidth.value = AIConfigStore.getSidebarWidth();
};

// 加载对话列表
const loadConversations = async () => {
  try {
    const res = await AIApi.getConversationList();
    conversations.value = res.data;
  } catch (err) {
    console.error("加载对话列表失败", err);
  }
};

// 加载对话详情
const loadConversation = async (id: string) => {
  if (!id) {
    // 新建对话模式
    tasks.value = [];
    expandedReasonings.value.clear();
    isFirstAiResponse.value = false;
    return;
  }

  try {
    const res = await AIApi.getConversation(id);
    const content = JSON.parse(res.data.content || "[]") as TaskBlock[];
    tasks.value = content;
    expandedReasonings.value.clear();
    isFirstAiResponse.value = content.length > 0;
  } catch (err) {
    console.error("加载对话详情失败", err);
    ElMessage.error("加载对话失败");
  }
};

// 处理对话切换
const handleConversationChange = async (id: string) => {
  // 如果是从搜索结果中选择的，先关闭搜索模式（防止 visible-change 重新打开下拉列表）
  const wasSearching = searchExpanded.value;
  if (wasSearching) {
    searchExpanded.value = false;
    searchKeyword.value = "";
  }

  await loadConversation(id);

  if (wasSearching) {
    await loadConversations();
  }
  scrollToBottom();
};

// 搜索相关
const expandSearch = () => {
  searchExpanded.value = true;
  // 保存当前对话ID，用于取消搜索时恢复
  savedConversationId.value = currentConversationId.value;
  nextTick(() => {
    // 先展开下拉列表，再聚焦搜索框
    (conversationSelectRef.value as any)?.toggleMenu?.();
    setTimeout(() => {
      searchInputRef.value?.focus();
    }, 50);
  });
};

const collapseSearch = () => {
  searchExpanded.value = false;
  searchKeyword.value = "";
  // 恢复到搜索前的对话
  currentConversationId.value = savedConversationId.value;
  // 关闭下拉列表
  (conversationSelectRef.value as any)?.blur?.();
  loadConversations();
};

const handleSearch = () => {
  // 清除之前的定时器
  if (searchDebounceTimer.value) {
    clearTimeout(searchDebounceTimer.value);
  }

  // 设置防抖延迟
  searchDebounceTimer.value = setTimeout(async () => {
    if (!searchKeyword.value) {
      await loadConversations();
      return;
    }

    try {
      const res = await AIApi.searchConversations(searchKeyword.value);
      conversations.value = res.data;
    } catch (err) {
      console.error("搜索失败", err);
    }
  }, 300);
};

const handleSearchClear = () => {
  loadConversations();
};

// 处理下拉列表显示/隐藏
const handleSelectVisibleChange = (visible: boolean) => {
  // 如果在搜索模式下，下拉列表被关闭
  if (!visible && searchExpanded.value) {
    // 检查当前焦点是否在搜索区域内，如果是则不取消搜索
    const activeElement = document.activeElement;
    if (searchAreaRef.value?.contains(activeElement)) {
      // 点击在搜索区域内，重新打开下拉列表
      nextTick(() => {
        (conversationSelectRef.value as any)?.toggleMenu?.();
      });
      return;
    }

    // 点击在其他地方，等待下拉列表动画结束后再收起搜索框
    setTimeout(() => {
      searchExpanded.value = false;
      searchKeyword.value = "";
      // 恢复到搜索前的对话
      currentConversationId.value = savedConversationId.value;
      loadConversations();
    }, 200); // Element Plus 下拉动画约 200ms
  }
};

// 监听 AI 配置变化事件
const handleAiConfigChanged = (event: CustomEvent<AIConfig>) => {
  config.value = event.detail;
};

onMounted(() => {
  loadConfig();
  loadConversations();
  window.addEventListener("ai-config-changed", handleAiConfigChanged as EventListener);
});

onUnmounted(() => {
  window.removeEventListener("ai-config-changed", handleAiConfigChanged as EventListener);
  abortController.value?.abort();
});

watch(
  () => props.visible,
  async (val) => {
    if (val) {
      await loadConfig();
      await loadConversations();
      scrollToBottom();
    }
  }
);

// 状态指示器样式
const statusClass = computed(() => {
  if (!config.value.baseUrl || !config.value.apiKey) return "status-error";
  if (!config.value.model) return "status-warning";
  return "status-ok";
});

// 输入框占位符
const inputPlaceholder = computed(() => {
  return "输入你的问题...";
});

// 切换思考内容展开/收起
const toggleReasoning = (taskId: string) => {
  if (expandedReasonings.value.has(taskId)) {
    expandedReasonings.value.delete(taskId);
  } else {
    expandedReasonings.value.add(taskId);
  }
};

// 处理键盘事件
const handleKeydown = (event: KeyboardEvent) => {
  if (event.key === "Enter" && !event.shiftKey) {
    event.preventDefault();
    if (inputText.value.trim() && !sending.value) {
      sendMessage();
    }
  }
};

// 停止生成
const stopGeneration = () => {
  abortController.value?.abort();
  abortController.value = null;
  sending.value = false;
  autoScrollEnabled.value = true;

  // 更新当前任务状态
  const currentTask = tasks.value[tasks.value.length - 1];
  if (currentTask && currentTask.status === "processing") {
    currentTask.status = "completed";
    saveConversation();
  }
};

// 生成对话标题（使用 AI）
const generateTitle = async (userMessage: string, aiResponse: string): Promise<string> => {
  try {
    const messages: ChatMessage[] = [
      {
        role: "system",
        content: "你是一个标题生成器。根据用户的问题和AI的回复，生成一个简短的对话标题（不超过20个字符）。只输出标题本身，不要有任何其他内容。"
      },
      {
        role: "user",
        content: `用户问题：${userMessage}\n\nAI回复开头：${aiResponse.slice(0, 200)}`
      }
    ];

    const response = await AIApi.chatDirect(config.value, messages);
    let title = response.choices[0]?.message?.content?.trim() || "新对话";
    // 限制长度
    if (title.length > 20) {
      title = title.slice(0, 20) + "...";
    }
    return title;
  } catch (err) {
    console.error("生成标题失败", err);
    // 使用用户消息的前20个字符作为标题
    return userMessage.slice(0, 20) + (userMessage.length > 20 ? "..." : "");
  }
};

// 保存对话
const saveConversation = async () => {
  // 如果没有对话ID，则不保存（新对话已在 sendMessage 中创建）
  if (!currentConversationId.value) return;

  const contentJson = JSON.stringify(tasks.value.map(t => ({
    id: t.id,
    userTask: t.userTask,
    reasoning: t.reasoning,
    output: t.output,
    status: t.status,
    timestamp: t.timestamp,
  })));

  try {
    await AIApi.updateConversation({
      id: currentConversationId.value,
      content: contentJson,
    });
  } catch (err) {
    console.error("保存对话失败", err);
  }
};

// 发送消息
const sendMessage = async () => {
  if (!inputText.value.trim() || sending.value) return;

  // 检查配置
  if (!config.value.baseUrl || !config.value.apiKey || !config.value.model) {
    ElMessage.warning("请先配置 AI API");
    return;
  }

  const userInput = inputText.value.trim();
  inputText.value = "";

  // 判断是否需要在 AI 回复后生成标题：新对话且是第一条消息
  const isNewConversation = !currentConversationId.value;

  // 如果是新对话，先创建对话记录
  if (isNewConversation) {
    try {
      const res = await AIApi.addConversation({
        title: "新对话",
        content: "[]",
      });
      currentConversationId.value = res.data.id;
      await loadConversations();
    } catch (err) {
      console.error("创建对话失败", err);
      ElMessage.error("创建对话失败");
      return;
    }
  }

  const task: TaskBlock = {
    id: Date.now().toString(),
    userTask: userInput,
    reasoning: "",
    output: "",
    status: "processing",
    timestamp: Date.now(),
  };

  tasks.value.push(task);
  await nextTick();
  // Reset auto-scroll when AI starts generating
  autoScrollEnabled.value = true;
  scrollToBottom(true);

  sending.value = true;
  abortController.value = new AbortController();

  try {
    // 构建消息历史
    const messages: ChatMessage[] = [];

    // 添加历史消息（最近5轮）
    const recentTasks = tasks.value.slice(-6, -1);
    for (const t of recentTasks) {
      if (t.status === "completed") {
        messages.push({ role: "user", content: t.userTask });
        messages.push({ role: "assistant", content: t.output });
      }
    }

    // 添加当前消息
    messages.push({ role: "user", content: userInput });

    // 使用流式请求
    await AIApi.chatStream(config.value, messages, {
      docTitle: config.value.docContextEnabled ? props.docTitle : undefined,
      docContent: config.value.docContextEnabled ? props.docContent : undefined,
      signal: abortController.value.signal,
      onReasoning: (chunk) => {
        task.reasoning += chunk;
        triggerRef(tasks);
        scrollToBottom();
      },
      onContent: (chunk) => {
        task.output += chunk;
        triggerRef(tasks);
        scrollToBottom();
      },
      onDone: () => {
        task.status = "completed";
        triggerRef(tasks);
      },
      onError: (err) => {
        task.status = "error";
        task.output = err.message || "请求失败";
        triggerRef(tasks);
      },
    });
  } catch (err: any) {
    task.status = "error";
    task.output = err.message || "请求失败";
    triggerRef(tasks);
  } finally {
    sending.value = false;
    autoScrollEnabled.value = true;
    abortController.value = null;
    await saveConversation();

    // 如果是新对话的第一次AI回复成功，生成标题
    if (isNewConversation && task.status === "completed" && currentConversationId.value) {
      const title = await generateTitle(userInput, task.output);
      try {
        await AIApi.updateConversationTitle(currentConversationId.value, title);
        await loadConversations();
      } catch (err) {
        console.error("更新标题失败", err);
      }
    }

    scrollToBottom();
  }
};

// Check if user is near bottom of scroll area
const isNearBottom = () => {
  if (!taskBlocksRef.value) return true;
  const { scrollTop, scrollHeight, clientHeight } = taskBlocksRef.value;
  return scrollHeight - scrollTop - clientHeight < 180;
};

// Handle scroll event to detect user scroll direction
const handleTaskBlocksScroll = () => {
  // Only track scroll state while AI is generating
  if (!sending.value || !taskBlocksRef.value) return;

  const { scrollTop } = taskBlocksRef.value;
  const scrollDelta = scrollTop - lastScrollTop.value;

  // User scrolled up - disable auto scroll (any upward movement)
  if (scrollDelta < -5) {
    autoScrollEnabled.value = false;
  }

  // User actively scrolled down and reached near bottom - re-enable
  if (scrollDelta > 5 && isNearBottom()) {
    autoScrollEnabled.value = true;
  }

  lastScrollTop.value = scrollTop;
};

// Scroll to bottom only if auto-scroll is enabled and AI is generating
const scrollToBottom = (force = false) => {
  nextTick(() => {
    if (taskBlocksRef.value && (force || (autoScrollUserEnabled.value && autoScrollEnabled.value && sending.value))) {
      taskBlocksRef.value.scrollTop = taskBlocksRef.value.scrollHeight;
      lastScrollTop.value = taskBlocksRef.value.scrollTop;
    }
  });
};

// 删除当前对话
const deleteCurrentConversation = async () => {
  if (!currentConversationId.value) return;

  try {
    await ElMessageBox.confirm("确定要删除当前对话吗？", "提示", {
      confirmButtonText: "确定",
      cancelButtonText: "取消",
      type: "warning",
    });

    await AIApi.deleteConversation(currentConversationId.value);
    currentConversationId.value = "";
    tasks.value = [];
    expandedReasonings.value.clear();
    await loadConversations();
    ElMessage.success("对话已删除");
  } catch (err: any) {
    if (err !== "cancel") {
      console.error("删除对话失败", err);
      ElMessage.error("删除失败");
    }
  }
};

// 导出对话历史
const exportHistory = () => {
  if (tasks.value.length === 0) {
    ElMessage.warning("没有可导出的对话");
    return;
  }

  // 构建导出数据，移除运行时状态
  const exportData = {
    version: "1.0",
    exportTime: Date.now(),
    model: config.value.model,
    tasks: tasks.value.map((task) => ({
      id: task.id,
      userTask: task.userTask,
      reasoning: task.reasoning,
      output: task.output,
      status: task.status,
      timestamp: task.timestamp,
    })),
  };

  const json = JSON.stringify(exportData, null, 2);
  const blob = new Blob([json], { type: "application/json" });
  const url = URL.createObjectURL(blob);

  const a = document.createElement("a");
  a.href = url;
  a.download = `ai-chat-${new Date().toISOString().slice(0, 10)}.json`;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);

  ElMessage.success("对话已导出");
};

// 删除单条对话
const deleteTask = async (index: number) => {
  const task = tasks.value[index];
  if (task) {
    // 移除对应的推理展开状态
    expandedReasonings.value.delete(task.id);
    // 从数组中删除
    tasks.value = tasks.value.filter((_, i) => i !== index);
    await saveConversation();
    ElMessage.success("已删除该对话");
  }
};

// 复制AI回复内容
const copyOutput = async (content: string) => {
  try {
    await navigator.clipboard.writeText(content);
    ElMessage.success("已复制到剪贴板");
  } catch (err) {
    console.error("复制失败", err);
    ElMessage.error("复制失败");
  }
};

// 拖拽调整大小
let isResizing = false;
let startX = 0;
let startWidth = 0;

const startResize = (e: MouseEvent) => {
  isResizing = true;
  startX = e.clientX;
  startWidth = sidebarWidth.value;
  document.addEventListener("mousemove", doResize);
  document.addEventListener("mouseup", stopResize);
};

const doResize = (e: MouseEvent) => {
  if (!isResizing) return;
  const diff = startX - e.clientX;
  const newWidth = Math.max(300, Math.min(600, startWidth + diff));
  sidebarWidth.value = newWidth;
};

const stopResize = () => {
  isResizing = false;
  AIConfigStore.setSidebarWidth(sidebarWidth.value);
  document.removeEventListener("mousemove", doResize);
  document.removeEventListener("mouseup", stopResize);
};
</script>

<style lang="scss" scoped>
.ai-sidebar {
  height: 100%;
  border-left: 1px solid #e4e7ed;
  background: #fff;
  display: flex;
  flex-direction: column;
  position: relative;
  transition: margin-right 0.3s;

  &.hidden {
    display: none;
  }
}

.resize-handle {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 4px;
  cursor: ew-resize;
  background: transparent;
  z-index: 10;

  &:hover {
    background: #409eff;
  }
}

.sidebar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #e4e7ed;
  flex-shrink: 0;

  .header-left {
    display: flex;
    align-items: center;
    gap: 8px;

    .model-name {
      font-size: 14px;
      font-weight: 500;
      color: #303133;
      max-width: 180px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .status-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      flex-shrink: 0;

      &.status-ok {
        background: #67c23a;
      }
      &.status-warning {
        background: #e6a23c;
      }
      &.status-error {
        background: #f56c6c;
      }
    }
  }

  .header-right {
    display: flex;
    align-items: center;

    .header-actions {
      display: flex;
      align-items: center;

      :deep(.el-button) {
        padding: 4px;
        margin: 0;
      }
    }
  }
}

.conversation-bar {
  display: flex;
  align-items: center;
  padding: 8px 16px;
  gap: 8px;
  border-bottom: 1px solid #e4e7ed;
  flex-shrink: 0;

  .search-area {
    display: flex;
    align-items: center;
    gap: 4px;

    .el-input {
      width: 120px;
    }
  }

  .conversation-select {
    flex: 1;
    min-width: 0;
  }
}

.task-blocks {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.task-block {
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #ebeef5;

  &:last-child {
    border-bottom: none;
  }
}

.task-label {
  font-size: 12px;
  font-weight: 500;
  color: #909399;
  margin-bottom: 6px;
  text-transform: uppercase;
}

.user-task {
  margin-bottom: 12px;

  .task-label-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 6px;

    .task-label {
      margin-bottom: 0;
    }

    .delete-task-btn {
      padding: 2px 4px;
      color: #909399;

      &:hover {
        color: #f56c6c;
      }
    }
  }

  .task-content {
    background: #f5f7fa;
    padding: 10px 12px;
    border-radius: 6px;
    font-size: 14px;
    line-height: 1.6;
    white-space: pre-wrap;
  }
}

.reasoning-section {
  margin-bottom: 12px;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  overflow: hidden;

  .reasoning-header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 12px;
    background: #fafafa;
    cursor: pointer;
    user-select: none;

    &:hover {
      background: #f5f7fa;
    }

    .el-icon {
      transition: transform 0.2s;
      color: #909399;

      &.is-expanded {
        transform: rotate(90deg);
      }
    }

    .reasoning-title {
      font-size: 13px;
      font-weight: 500;
      color: #606266;
    }

    .reasoning-hint {
      font-size: 12px;
      color: #c0c4cc;
      margin-left: auto;
    }
  }

  .reasoning-content {
    padding: 12px;
    background: #fff;
    border-top: 1px solid #e4e7ed;

    pre {
      margin: 0;
      font-size: 13px;
      color: #909399;
      line-height: 1.6;
      white-space: pre-wrap;
      word-break: break-word;
      font-family: inherit;
    }
  }
}

.ai-output {
  margin-bottom: 12px;

  .output-label-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 6px;

    .task-label {
      margin-bottom: 0;
    }

    .copy-output-btn {
      padding: 2px 4px;
      color: #909399;

      &:hover {
        color: #409eff;
      }
    }
  }

  .output-content {
    font-size: 14px;
    line-height: 1.8;

    :deep(pre) {
      background: #f5f7fa;
      padding: 12px;
      border-radius: 4px;
      overflow-x: auto;
    }

    :deep(code) {
      background: #f5f7fa;
      padding: 2px 6px;
      border-radius: 3px;
      font-size: 13px;
    }

    :deep(pre code) {
      background: transparent;
      padding: 0;
    }

    .cursor-blink {
      animation: blink 1s step-end infinite;
      color: #409eff;
    }

    @keyframes blink {
      50% {
        opacity: 0;
      }
    }
  }
}

.error-state {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 10px;
  border-radius: 4px;
  font-size: 13px;
  background: #fff2f0;
  color: #ff4d4f;

  .el-icon {
    flex-shrink: 0;
    margin-top: 2px;
  }

  span {
    word-break: break-word;
    white-space: pre-wrap;
    overflow-wrap: break-word;
  }
}

.input-area {
  padding: 16px;
  border-top: 1px solid #e4e7ed;
  flex-shrink: 0;

  .input-actions {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 10px;

    .input-left {
      display: flex;
      align-items: center;
      gap: 12px;
    }

    .input-tip {
      font-size: 12px;
      color: #909399;
    }

    .auto-scroll-checkbox {
      :deep(.el-checkbox__label) {
        font-size: 12px;
        color: #909399;
      }
    }
  }
}
</style>

<!-- 全局样式，用于下拉弹出层（弹出层挂载在body上，scoped样式无法影响） -->
<style lang="scss">
.conversation-option {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;

  .conv-title {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    margin-right: 12px;
  }

  .conv-time {
    font-size: 11px;
    color: #909399;
    flex-shrink: 0;
    text-align: right;
  }

  &.is-current {
    .conv-time {
      font-weight: 600;
    }
  }
}
</style>
