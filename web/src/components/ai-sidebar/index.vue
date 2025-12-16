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
      <div class="search-area">
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
    <div class="task-blocks" ref="taskBlocksRef">
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

        <!-- AI 输出（Chat模式流式显示，Agent模式仅显示说明） -->
        <div class="ai-output" v-if="task.output || (task.status === 'processing' && task.mode !== 'agent')">
          <div class="task-label">Output</div>
          <div class="output-content">
            <MdPreview v-if="task.output" :modelValue="getMarkdownContent(task)" previewTheme="cyanosis" />
            <span v-else-if="task.status === 'processing'" class="cursor-blink">▋</span>
          </div>
        </div>

        <!-- Agent 模式加载状态 -->
        <div v-if="task.status === 'processing' && task.mode === 'agent'" class="agent-loading">
          <div class="loading-spinner"></div>
          <span class="loading-text">AI 正在分析文档并生成修改建议...</span>
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
import { ElMessage, ElMessageBox } from "element-plus";
import { Delete, Close, Warning, ArrowRight, VideoPause, Download, Search } from "@element-plus/icons-vue";
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

// 对话记录相关
const conversations = ref<AIConversationListItem[]>([]);
const currentConversationId = ref<string>("");
const searchExpanded = ref(false);
const searchKeyword = ref("");
const searchInputRef = ref<HTMLInputElement | null>(null);
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
  agentEnabled: false,
  docContextEnabled: false,
  panelEnabled: false,
});

// 过滤后的对话列表
const filteredConversations = computed(() => {
  if (!searchKeyword.value) {
    return conversations.value;
  }
  return conversations.value;
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
  // 如果在搜索模式下，下拉列表被关闭（非选择导致），重新打开
  if (!visible && searchExpanded.value) {
    nextTick(() => {
      (conversationSelectRef.value as any)?.toggleMenu?.();
      // 保持搜索框焦点
      setTimeout(() => {
        searchInputRef.value?.focus();
      }, 50);
    });
  }
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
const getMarkdownContent = (task: TaskBlock) => {
  // Agent 模式显示 explanation
  if (task.mode === "agent" && task.agentResponse?.explanation) {
    return task.agentResponse.explanation;
  }
  return task.output;
};

// 尝试增量解析 Agent 响应的 JSON
const tryParseAgentResponse = (
  task: TaskBlock,
  lastParsedChangesCount: number,
  onNewChanges: (newCount: number) => void
) => {
  const content = task.output;

  // 尝试提取 changes 数组中已完成的对象
  const changesMatch = content.match(/"changes"\s*:\s*\[/);
  if (!changesMatch) return;

  const changesStartIndex = changesMatch.index! + changesMatch[0].length;
  const changesContent = content.slice(changesStartIndex);

  // 逐个解析 change 对象
  const changes: AgentChange[] = [];
  let depth = 0;
  let objectStart = -1;
  let inString = false;
  let escapeNext = false;

  for (let i = 0; i < changesContent.length; i++) {
    const char = changesContent[i];

    if (escapeNext) {
      escapeNext = false;
      continue;
    }

    if (char === "\\") {
      escapeNext = true;
      continue;
    }

    if (char === '"' && !escapeNext) {
      inString = !inString;
      continue;
    }

    if (inString) continue;

    if (char === "{") {
      if (depth === 0) {
        objectStart = i;
      }
      depth++;
    } else if (char === "}") {
      depth--;
      if (depth === 0 && objectStart !== -1) {
        // 找到一个完整的对象
        const objectStr = changesContent.slice(objectStart, i + 1);
        try {
          const change = JSON.parse(objectStr) as AgentChange;
          changes.push(change);
        } catch {
          // 解析失败，忽略
        }
        objectStart = -1;
      }
    } else if (char === "]" && depth === 0) {
      // changes 数组结束
      break;
    }
  }

  // 如果解析到新的 changes，更新 task
  if (changes.length > lastParsedChangesCount) {
    if (!task.agentResponse) {
      task.agentResponse = { plan: [], changes: [], explanation: "" };
    }
    task.agentResponse.changes = changes;
    onNewChanges(changes.length);
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
  const contentJson = JSON.stringify(tasks.value.map(t => ({
    id: t.id,
    userTask: t.userTask,
    reasoning: t.reasoning,
    output: t.output,
    agentResponse: t.agentResponse ? {
      plan: t.agentResponse.plan,
      changes: t.agentResponse.changes.map(c => ({
        type: c.type,
        position: c.position,
        oldText: c.oldText,
        content: c.content,
      })),
      explanation: t.agentResponse.explanation,
    } : undefined,
    status: t.status,
    timestamp: t.timestamp,
    mode: t.mode,
  })));

  try {
    if (currentConversationId.value) {
      // 更新现有对话
      await AIApi.updateConversation({
        id: currentConversationId.value,
        content: contentJson,
      });
    } else if (tasks.value.length > 0) {
      // 创建新对话
      const res = await AIApi.addConversation({
        title: "新对话",
        content: contentJson,
      });
      currentConversationId.value = res.data.id;
      isFirstAiResponse.value = false;
      await loadConversations();
    }
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

  const isAgentMode = mode.value === "agent" && config.value.docContextEnabled;
  const needGenerateTitle = !currentConversationId.value && tasks.value.length === 0;

  const task: TaskBlock = {
    id: Date.now().toString(),
    userTask: userInput,
    reasoning: "",
    output: "",
    status: "processing",
    timestamp: Date.now(),
    mode: mode.value,
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

    if (isAgentMode) {
      // Agent 模式使用流式请求，实时解析 JSON
      let lastParsedChangesCount = 0;

      await AIApi.chatStream(config.value, messages, {
        agentMode: true,
        docTitle: props.docTitle,
        docContent: props.docContent,
        signal: abortController.value.signal,
        onReasoning: (chunk) => {
          task.reasoning += chunk;
          triggerRef(tasks);
          scrollToBottom();
        },
        onContent: (chunk) => {
          task.output += chunk;

          // 尝试增量解析 JSON，提取已完成的 changes
          tryParseAgentResponse(task, lastParsedChangesCount, (newCount) => {
            lastParsedChangesCount = newCount;
          });

          triggerRef(tasks);
          scrollToBottom();
        },
        onDone: () => {
          // 最终解析完整 JSON
          try {
            const parsed = JSON.parse(task.output);
            if (parsed.plan || parsed.changes) {
              task.agentResponse = parsed as AgentResponse;
            }
          } catch {
            // 不是有效的 JSON，作为普通输出处理
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
    } else {
      // Chat 模式使用流式请求
      await AIApi.chatStream(config.value, messages, {
        agentMode: false,
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
    }
  } catch (err: any) {
    task.status = "error";
    task.output = err.message || "请求失败";
    triggerRef(tasks);
  } finally {
    sending.value = false;
    abortController.value = null;
    await saveConversation();

    // 如果是新对话的第一次AI回复，生成标题
    if (needGenerateTitle && task.status === "completed" && currentConversationId.value) {
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

// 滚动到底部
const scrollToBottom = () => {
  nextTick(() => {
    if (taskBlocksRef.value) {
      taskBlocksRef.value.scrollTop = taskBlocksRef.value.scrollHeight;
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
      agentResponse: task.agentResponse
        ? {
            plan: task.agentResponse.plan,
            changes: task.agentResponse.changes.map((c) => ({
              type: c.type,
              position: c.position,
              oldText: c.oldText,
              content: c.content,
            })),
            explanation: task.agentResponse.explanation,
          }
        : undefined,
      status: task.status,
      timestamp: task.timestamp,
      mode: task.mode,
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

    .header-actions {
      display: flex;
      align-items: center;
      margin-left: 4px;

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

.agent-loading {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background: #f0f9ff;
  border: 1px solid #bae0ff;
  border-radius: 6px;
  margin-bottom: 12px;

  .loading-spinner {
    width: 20px;
    height: 20px;
    border: 2px solid #91caff;
    border-top-color: #1677ff;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  .loading-text {
    font-size: 13px;
    color: #1677ff;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
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
