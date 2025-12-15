<template>
  <div class="ai-sidebar" :class="{ hidden: !visible }" :style="{ width: sidebarWidth + 'px' }">
    <!-- 拖拽调整大小手柄 -->
    <div class="resize-handle" @mousedown="startResize"></div>

    <!-- 顶部工具栏 -->
    <div class="sidebar-header">
      <div class="header-left">
        <el-select v-model="mode" size="small" style="width: 90px" @change="handleModeChange">
          <el-option label="Chat" value="chat" />
          <el-option label="Agent" value="agent" />
        </el-select>
      </div>
      <div class="header-right">
        <span class="model-name" :title="config.model">{{ config.model || "未配置模型" }}</span>
        <span class="status-dot" :class="statusClass"></span>
        <el-button text size="small" @click="clearHistory" title="清空会话">
          <el-icon><Delete /></el-icon>
        </el-button>
        <el-button text size="small" @click="emit('close')" title="关闭">
          <el-icon><Close /></el-icon>
        </el-button>
      </div>
    </div>

    <!-- 任务块区域 -->
    <div class="task-blocks" ref="taskBlocksRef">
      <div v-for="task in tasks" :key="task.id" class="task-block">
        <!-- 用户任务 -->
        <div class="user-task">
          <div class="task-label">Task</div>
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

        <!-- AI 输出（流式显示） -->
        <div class="ai-output" v-if="task.output || task.status === 'processing'">
          <div class="task-label">Output</div>
          <div class="output-content">
            <MdPreview v-if="task.output" :modelValue="getMarkdownContent(task.output)" previewTheme="cyanosis" />
            <span v-else-if="task.status === 'processing'" class="cursor-blink">▋</span>
          </div>
        </div>

        <!-- Agent 变更建议 -->
        <div v-if="task.agentResponse?.changes?.length" class="agent-changes">
          <div class="task-label">Proposed Changes</div>
          <div v-for="(change, idx) in task.agentResponse.changes" :key="idx" class="change-item">
            <div class="change-header">
              <el-tag :type="getChangeTagType(change.type)" size="small">{{ change.type }}</el-tag>
              <span class="change-position">{{ change.position }}</span>
              <el-tag v-if="change.applied" type="success" size="small" class="applied-tag">已应用</el-tag>
            </div>
            <div class="change-content">
              <pre>{{ change.content }}</pre>
            </div>
            <div class="change-actions">
              <el-button
                v-if="!change.applied"
                size="small"
                type="primary"
                @click="applyChange(change, idx)"
              >
                应用此修改
              </el-button>
              <el-button
                v-else
                size="small"
                type="warning"
                @click="undoChange(change, idx)"
              >
                撤回此修改
              </el-button>
            </div>
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
        <span class="input-tip">Enter 发送，Shift+Enter 换行</span>
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
import { ElMessage } from "element-plus";
import { Delete, Close, Warning, ArrowRight, VideoPause } from "@element-plus/icons-vue";
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
  (e: "applyChange", change: AgentChange, callback: (success: boolean, undoData?: { originalContent: string; appliedContent: string }) => void): void;
  (e: "undoChange", change: AgentChange, callback: (success: boolean) => void): void;
}>();

const mode = ref<"chat" | "agent">("chat");
const inputText = ref("");
const sending = ref(false);
const sidebarWidth = ref(380);
const tasks = shallowRef<TaskBlock[]>([]);
const taskBlocksRef = ref<HTMLElement | null>(null);
const expandedReasonings = ref<Set<string>>(new Set());
const abortController = ref<AbortController | null>(null);

const config = ref<AIConfig>({
  baseUrl: "",
  apiKey: "",
  model: "",
  systemPrompts: [],
  currentPromptId: "",
  agentEnabled: false,
  docContextEnabled: false,
  panelEnabled: false,
});

// 加载配置
const loadConfig = async () => {
  config.value = await AIConfigStore.getConfig();
  sidebarWidth.value = AIConfigStore.getSidebarWidth();

  // 加载聊天历史
  const history = await AIConfigStore.getChatHistory();
  tasks.value = history;
};

// 监听 AI 配置变化事件
const handleAiConfigChanged = (event: CustomEvent<AIConfig>) => {
  const newConfig = event.detail;

  // 如果文档上下文权限被关闭，自动切换到 Chat 模式
  if (mode.value === "agent" && !newConfig.docContextEnabled) {
    mode.value = "chat";
  }

  config.value = newConfig;
};

// 监听配置变化，确保模式与配置一致
watch(
  () => config.value,
  (newConfig) => {
    // 如果当前是 Agent 模式，但文档上下文权限被关闭，自动切换到 Chat 模式
    if (mode.value === "agent" && !newConfig.docContextEnabled) {
      mode.value = "chat";
    }
  },
  { deep: true }
);

onMounted(() => {
  loadConfig();
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
  if (mode.value === "agent") {
    return "描述你想要对文档进行的修改...";
  }
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

// 处理模式切换
const handleModeChange = (newMode: "chat" | "agent") => {
  if (newMode === "agent" && !config.value.docContextEnabled) {
    // 如果选择 Agent 模式但未开启文档上下文权限，提示用户
    ElMessage.warning("Agent 模式需要先在 AI 配置中开启「允许读取当前文档」权限");
    // 切回 Chat 模式
    nextTick(() => {
      mode.value = "chat";
    });
  }
};

// 渲染 Markdown - 提取需要渲染的内容
const getMarkdownContent = (content: string) => {
  try {
    // 尝试解析为 Agent JSON 格式
    if (mode.value === "agent" && content.startsWith("{")) {
      const parsed = JSON.parse(content);
      if (parsed.explanation) {
        return parsed.explanation;
      }
    }
    return content;
  } catch {
    return content;
  }
};

// 获取变更标签类型
const getChangeTagType = (type: string) => {
  switch (type) {
    case "insert":
      return "success";
    case "delete":
      return "danger";
    default:
      return "warning";
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

  // 更新当前任务状态
  const currentTask = tasks.value[tasks.value.length - 1];
  if (currentTask && currentTask.status === "processing") {
    currentTask.status = "completed";
    saveChatHistory();
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
  scrollToBottom();

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

    // 流式调用 AI API
    await AIApi.chatStream(config.value, messages, {
      agentMode: mode.value === "agent" && config.value.docContextEnabled,
      docTitle: config.value.docContextEnabled ? props.docTitle : undefined,
      docContent: config.value.docContextEnabled ? props.docContent : undefined,
      signal: abortController.value.signal,
      onReasoning: (chunk) => {
        task.reasoning += chunk;
        triggerRef(tasks); // 强制触发视图更新
        scrollToBottom();
      },
      onContent: (chunk) => {
        task.output += chunk;
        triggerRef(tasks); // 强制触发视图更新
        scrollToBottom();
      },
      onDone: () => {
        // 如果是 Agent 模式，尝试解析结构化响应
        if (mode.value === "agent" && config.value.docContextEnabled) {
          try {
            const parsed = JSON.parse(task.output);
            if (parsed.plan || parsed.changes) {
              task.agentResponse = parsed as AgentResponse;
            }
          } catch {
            // 不是有效的 JSON，作为普通输出处理
          }
        }
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
    abortController.value = null;
    await saveChatHistory();
    scrollToBottom();
  }
};

// 滚动到底部
const scrollToBottom = () => {
  nextTick(() => {
    if (taskBlocksRef.value) {
      taskBlocksRef.value.scrollTop = taskBlocksRef.value.scrollHeight;
    }
  });
};

// 清空历史
const clearHistory = async () => {
  if (sending.value) {
    stopGeneration();
  }
  tasks.value = [];
  expandedReasonings.value.clear();
  await AIConfigStore.removeChatHistory();
  ElMessage.success("会话已清空");
};

// 保存聊天历史
const saveChatHistory = async () => {
  // 只保留最近20条
  const toSave = tasks.value.slice(-20);
  await AIConfigStore.setChatHistory(toSave);
};

// 应用变更
const applyChange = (change: AgentChange, index: number) => {
  emit("applyChange", change, (success, undoData) => {
    if (success && undoData) {
      change.applied = true;
      change.undoData = undoData;
      triggerRef(tasks);
    }
  });
};

// 撤回变更
const undoChange = (change: AgentChange, index: number) => {
  emit("undoChange", change, (success) => {
    if (success) {
      change.applied = false;
      change.undoData = undefined;
      triggerRef(tasks);
    }
  });
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
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 8px;

    .model-name {
      font-size: 12px;
      color: #606266;
      max-width: 100px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .status-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;

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

.agent-changes {
  .change-item {
    background: #fafafa;
    border: 1px solid #ebeef5;
    border-radius: 6px;
    padding: 12px;
    margin-bottom: 10px;

    .change-header {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-bottom: 8px;

      .change-position {
        font-size: 13px;
        color: #606266;
      }

      .applied-tag {
        margin-left: auto;
      }
    }

    .change-content {
      margin-bottom: 10px;

      pre {
        background: #fff;
        border: 1px solid #e4e7ed;
        padding: 10px;
        border-radius: 4px;
        font-size: 13px;
        margin: 0;
        white-space: pre-wrap;
        word-break: break-all;
      }
    }
  }
}

.error-state {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px;
  border-radius: 4px;
  font-size: 13px;
  background: #fff2f0;
  color: #ff4d4f;
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

    .input-tip {
      font-size: 12px;
      color: #909399;
    }
  }
}
</style>
