<template>
  <el-dialog v-model="visible" title="AI 配置" width="600px" :close-on-click-modal="false" @close="handleClose">
    <el-tabs v-model="activeTab">
      <!-- 基础配置 -->
      <el-tab-pane label="基础配置" name="basic">
        <el-form label-width="100px" label-position="left">
          <el-form-item label="Base URL">
            <el-input v-model="form.baseUrl" placeholder="https://api.openai.com" clearable />
          </el-form-item>
          <el-form-item label="API Key">
            <el-input
              v-model="form.apiKey"
              :type="showApiKey ? 'text' : 'password'"
              placeholder="sk-..."
              clearable
              autocomplete="new-password"
              name="ai-api-key"
            >
              <template #suffix>
                <el-icon class="cursor-pointer" @click="showApiKey = !showApiKey">
                  <View v-if="showApiKey" />
                  <Hide v-else />
                </el-icon>
              </template>
            </el-input>
          </el-form-item>
          <el-form-item label="模型">
            <div class="model-select">
              <el-select v-if="!manualModelInput" v-model="form.model" placeholder="选择模型" filterable allow-create style="flex: 1">
                <el-option v-for="model in modelList" :key="model.id" :label="model.id" :value="model.id" />
              </el-select>
              <el-input v-else v-model="form.model" placeholder="输入模型名称" style="flex: 1" />
              <el-button :loading="modelsLoading" @click="fetchModels" style="margin-left: 8px">获取模型</el-button>
              <el-button text @click="manualModelInput = !manualModelInput" style="margin-left: 4px">
                {{ manualModelInput ? "选择" : "手动" }}
              </el-button>
            </div>
          </el-form-item>
          <el-form-item>
            <div class="check-config">
              <el-button :loading="checkLoading" @click="checkConfig">检查配置</el-button>
              <span v-if="checkResult" :class="['check-result', checkResult.valid ? 'success' : 'error']">
                {{ checkResult.valid ? "✓" : "✗" }} {{ checkResult.message }}
                <span v-if="checkResult.latency">({{ checkResult.latency }}ms)</span>
              </span>
            </div>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <!-- 系统提示词 -->
      <el-tab-pane label="系统提示词" name="prompts">
        <div class="prompts-header">
          <span class="tip">管理多个系统提示词，选择一个作为当前生效提示词</span>
          <el-button type="primary" size="small" @click="addPrompt">新增</el-button>
        </div>
        <div class="prompts-list">
          <div v-for="(prompt, index) in form.systemPrompts" :key="prompt.id" class="prompt-item" :class="{ active: form.currentPromptId === prompt.id }">
            <el-radio v-model="form.currentPromptId" :label="prompt.id" @change="handlePromptSelect">
              <span class="prompt-name">{{ prompt.name || "未命名提示词" }}</span>
            </el-radio>
            <div class="prompt-actions">
              <el-button text size="small" @click="editPrompt(index)">编辑</el-button>
              <el-button text size="small" type="danger" @click="deletePrompt(index)">删除</el-button>
            </div>
          </div>
          <el-empty v-if="form.systemPrompts.length === 0" description="暂无系统提示词" />
        </div>
      </el-tab-pane>

      <!-- 模式设置 -->
      <el-tab-pane label="模式设置" name="mode">
        <el-form label-width="160px" label-position="left">
          <el-form-item label="启用 AI 面板">
            <el-switch v-model="form.panelEnabled" />
            <span class="form-tip">开启后在编辑器工具栏显示 AI 按钮</span>
          </el-form-item>
          <el-form-item label="允许读取当前文档">
            <el-switch v-model="form.docContextEnabled" />
            <span class="form-tip">开启后 AI 可以访问当前编辑的文档内容，并支持 Agent 模式</span>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <!-- 同步设置 -->
      <el-tab-pane label="同步设置" name="sync">
        <el-form label-width="140px" label-position="left">
          <el-form-item label="与服务器同步">
            <el-switch v-model="saveToServer" @change="handleSyncChange" :loading="syncLoading" />
            <span class="form-tip">开启后配置将同步到服务器，支持多设备</span>
          </el-form-item>
        </el-form>
        <div class="sync-actions">
          <el-button @click="exportConfig">导出配置</el-button>
          <el-button @click="importConfigClick">导入配置</el-button>
          <input ref="importInput" type="file" accept=".json" style="display: none" @change="handleImportFile" />
        </div>
      </el-tab-pane>
    </el-tabs>

    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
    </template>

    <!-- 提示词编辑对话框 -->
    <el-dialog v-model="promptDialogVisible" :title="editingPromptIndex === -1 ? '新增提示词' : '编辑提示词'" width="500px" append-to-body>
      <el-form label-width="80px">
        <el-form-item label="名称">
          <el-input v-model="editingPrompt.name" placeholder="提示词名称" />
        </el-form-item>
        <el-form-item label="内容">
          <el-input v-model="editingPrompt.content" type="textarea" :rows="8" placeholder="系统提示词内容" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="promptDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="savePrompt">确定</el-button>
      </template>
    </el-dialog>

    <!-- 同步冲突对话框 -->
    <el-dialog v-model="conflictDialogVisible" title="配置冲突" width="400px" append-to-body :close-on-click-modal="false">
      <p>服务器上已存在配置，请选择处理方式：</p>
      <template #footer>
        <div class="conflict-buttons">
          <el-button @click="conflictDialogVisible = false">取消</el-button>
          <el-button type="warning" @click="handleConflict('overwrite')">覆盖服务器</el-button>
          <el-button type="primary" @click="handleConflict('use')">使用服务器</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 覆盖确认对话框 -->
    <el-dialog v-model="overwriteDialogVisible" title="覆盖服务器配置" width="400px" append-to-body :close-on-click-modal="false">
      <p>是否先导出服务器配置作为备份？</p>
      <template #footer>
        <el-button @click="overwriteDialogVisible = false">取消</el-button>
        <el-button @click="doOverwrite(false)">直接覆盖</el-button>
        <el-button type="primary" @click="doOverwrite(true)">导出后覆盖</el-button>
      </template>
    </el-dialog>

    <!-- 使用服务器确认对话框 -->
    <el-dialog v-model="useServerDialogVisible" title="使用服务器配置" width="400px" append-to-body :close-on-click-modal="false">
      <p>是否先导出本地配置作为备份？</p>
      <template #footer>
        <el-button @click="useServerDialogVisible = false">取消</el-button>
        <el-button @click="doUseServer(false)">直接使用</el-button>
        <el-button type="primary" @click="doUseServer(true)">导出后使用</el-button>
      </template>
    </el-dialog>
  </el-dialog>
</template>

<script lang="ts" setup>
import { ref, watch, onMounted } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { View, Hide } from "@element-plus/icons-vue";
import AIApi from "@/api/ai";
import AIConfigStore from "@/store/ai-config";

const props = defineProps<{
  modelValue: boolean;
}>();

const emit = defineEmits<{
  (e: "update:modelValue", value: boolean): void;
  (e: "configChanged", config: AIConfig): void;
}>();

const visible = ref(props.modelValue);
const activeTab = ref("basic");
const saving = ref(false);
const showApiKey = ref(false);
const manualModelInput = ref(false);
const modelsLoading = ref(false);
const checkLoading = ref(false);
const syncLoading = ref(false);
const saveToServer = ref(false);

const modelList = ref<ModelInfo[]>([]);
const checkResult = ref<CheckConfigResponse | null>(null);
const serverConfig = ref<AIConfig | null>(null);

// 表单数据
const form = ref<AIConfig>({
  baseUrl: "",
  apiKey: "",
  model: "",
  systemPrompts: [],
  currentPromptId: "",
  agentEnabled: false,
  docContextEnabled: false,
  panelEnabled: false,
});

// 提示词编辑
const promptDialogVisible = ref(false);
const editingPromptIndex = ref(-1);
const editingPrompt = ref<SystemPrompt>({ id: "", name: "", content: "", isActive: false });

// 冲突处理
const conflictDialogVisible = ref(false);
const overwriteDialogVisible = ref(false);
const useServerDialogVisible = ref(false);

// 导入
const importInput = ref<HTMLInputElement | null>(null);

watch(
  () => props.modelValue,
  (val) => {
    visible.value = val;
    if (val) {
      loadConfig();
    }
  }
);

watch(visible, (val) => {
  emit("update:modelValue", val);
});

onMounted(() => {
  if (props.modelValue) {
    loadConfig();
  }
});

// 加载配置
const loadConfig = async () => {
  saveToServer.value = AIConfigStore.getSyncToServer();
  const localConfig = await AIConfigStore.getConfig();
  form.value = { ...localConfig };

  // 加载缓存的模型列表
  const cachedModels = await AIConfigStore.getModels();
  if (cachedModels) {
    modelList.value = cachedModels;
  }
};

// 获取模型列表
const fetchModels = async () => {
  if (!form.value.baseUrl || !form.value.apiKey) {
    ElMessage.warning("请先填写 Base URL 和 API Key");
    return;
  }
  modelsLoading.value = true;
  try {
    const res = await AIApi.getModelsDirect(form.value.baseUrl, form.value.apiKey);
    modelList.value = res.data || [];
    await AIConfigStore.setModels(modelList.value);
    ElMessage.success(`获取到 ${modelList.value.length} 个模型`);
  } catch (err: any) {
    ElMessage.error(err.message || "获取模型列表失败");
  } finally {
    modelsLoading.value = false;
  }
};

// 检查配置
const checkConfig = async () => {
  if (!form.value.baseUrl || !form.value.apiKey) {
    ElMessage.warning("请先填写 Base URL 和 API Key");
    return;
  }
  checkLoading.value = true;
  checkResult.value = null;
  try {
    checkResult.value = await AIApi.checkConfigDirect(form.value.baseUrl, form.value.apiKey);
  } catch (err: any) {
    checkResult.value = {
      valid: false,
      message: err.message || "检查失败",
      latency: 0,
    };
  } finally {
    checkLoading.value = false;
  }
};

// 新增提示词
const addPrompt = () => {
  editingPromptIndex.value = -1;
  editingPrompt.value = {
    id: Date.now().toString(),
    name: "",
    content: "",
    isActive: false,
  };
  promptDialogVisible.value = true;
};

// 编辑提示词
const editPrompt = (index: number) => {
  editingPromptIndex.value = index;
  editingPrompt.value = { ...form.value.systemPrompts[index] };
  promptDialogVisible.value = true;
};

// 保存提示词
const savePrompt = () => {
  if (!editingPrompt.value.name.trim()) {
    ElMessage.warning("请输入提示词名称");
    return;
  }
  if (editingPromptIndex.value === -1) {
    form.value.systemPrompts.push({ ...editingPrompt.value });
    // 如果是第一个提示词，自动设为当前
    if (form.value.systemPrompts.length === 1) {
      form.value.currentPromptId = editingPrompt.value.id;
    }
  } else {
    form.value.systemPrompts[editingPromptIndex.value] = { ...editingPrompt.value };
  }
  promptDialogVisible.value = false;
};

// 删除提示词
const deletePrompt = (index: number) => {
  ElMessageBox.confirm("确定删除该提示词？", "提示", {
    confirmButtonText: "删除",
    cancelButtonText: "取消",
    type: "warning",
  }).then(() => {
    const deleted = form.value.systemPrompts.splice(index, 1)[0];
    if (form.value.currentPromptId === deleted.id) {
      form.value.currentPromptId = form.value.systemPrompts[0]?.id || "";
    }
  });
};

// 选择提示词
const handlePromptSelect = () => {
  // 已通过 v-model 处理
};

// 处理同步开关变化
const handleSyncChange = async (enabled: boolean) => {
  if (enabled) {
    syncLoading.value = true;
    try {
      // 检查服务器是否有配置
      const existsRes = await AIApi.configExists();
      if (existsRes.data.exists) {
        // 获取服务器配置（完整版，包含完整API Key）
        const serverRes = await AIApi.getConfigFull();
        serverConfig.value = serverRes.data;
        // 显示冲突对话框
        conflictDialogVisible.value = true;
        saveToServer.value = false; // 暂时关闭，等用户确认
      } else {
        // 直接保存到服务器
        await saveConfigToServer();
        AIConfigStore.setSyncToServer(true);
        ElMessage.success("已开启服务器同步");
      }
    } catch (err: any) {
      ElMessage.error(err.message || "检查服务器配置失败");
      saveToServer.value = false;
    } finally {
      syncLoading.value = false;
    }
  } else {
    // 关闭同步
    ElMessageBox.confirm("是否删除服务器上的备份？", "关闭同步", {
      confirmButtonText: "删除备份",
      cancelButtonText: "保留备份",
      distinguishCancelAndClose: true,
      type: "info",
    })
      .then(async () => {
        // 删除服务器备份
        try {
          await AIApi.deleteConfig();
          ElMessage.success("已删除服务器备份");
        } catch (err: any) {
          ElMessage.error(err.message || "删除失败");
        }
        AIConfigStore.setSyncToServer(false);
      })
      .catch((action) => {
        if (action === "cancel") {
          // 保留备份
          AIConfigStore.setSyncToServer(false);
          ElMessage.info("已关闭同步，服务器备份已保留");
        } else {
          // 取消操作
          saveToServer.value = true;
        }
      });
  }
};

// 处理冲突
const handleConflict = (action: "overwrite" | "use") => {
  conflictDialogVisible.value = false;
  if (action === "overwrite") {
    overwriteDialogVisible.value = true;
  } else {
    useServerDialogVisible.value = true;
  }
};

// 执行覆盖服务器
const doOverwrite = async (exportFirst: boolean) => {
  if (exportFirst && serverConfig.value) {
    downloadConfig({
      version: "1.0",
      exportTime: Date.now(),
      config: serverConfig.value,
    });
  }
  overwriteDialogVisible.value = false;
  await saveConfigToServer();
  AIConfigStore.setSyncToServer(true);
  saveToServer.value = true;
  ElMessage.success("已覆盖服务器配置");
};

// 执行使用服务器配置
const doUseServer = async (exportFirst: boolean) => {
  if (exportFirst) {
    const localExport = await AIConfigStore.exportConfig();
    downloadConfig(localExport);
  }
  useServerDialogVisible.value = false;
  if (serverConfig.value) {
    form.value = { ...serverConfig.value };
    await AIConfigStore.setConfig(form.value);
  }
  AIConfigStore.setSyncToServer(true);
  saveToServer.value = true;
  ElMessage.success("已使用服务器配置");
};

// 导出配置
const exportConfig = async () => {
  const exportData = await AIConfigStore.exportConfig();
  downloadConfig(exportData);
  ElMessage.success("配置已导出");
};

// 下载配置文件
const downloadConfig = (data: AIConfigExport) => {
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: "application/json" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = `ai-config-${new Date().toISOString().slice(0, 10)}.json`;
  a.click();
  URL.revokeObjectURL(url);
};

// 导入配置点击
const importConfigClick = () => {
  importInput.value?.click();
};

// 处理导入文件
const handleImportFile = async (event: Event) => {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;

  try {
    const text = await file.text();
    const data = JSON.parse(text) as AIConfigExport;
    await AIConfigStore.importConfig(data);
    form.value = { ...data.config };
    ElMessage.success("配置已导入");
  } catch (err) {
    ElMessage.error("导入失败：无效的配置文件");
  }
  input.value = "";
};

// 保存配置到服务器
const saveConfigToServer = async () => {
  await AIApi.saveConfig(form.value);
};

// 保存
const handleSave = async () => {
  saving.value = true;
  try {
    // 保存到本地
    await AIConfigStore.setConfig(form.value);

    // 如果开启了同步，同时保存到服务器
    if (saveToServer.value) {
      await saveConfigToServer();
    }

    emit("configChanged", form.value);
    ElMessage.success("保存成功");
    handleClose();
  } catch (err: any) {
    ElMessage.error(err.message || "保存失败");
  } finally {
    saving.value = false;
  }
};

// 关闭
const handleClose = () => {
  visible.value = false;
  checkResult.value = null;
};
</script>

<style lang="scss" scoped>
.model-select {
  display: flex;
  width: 100%;
}

.check-config {
  display: flex;
  align-items: center;
  gap: 12px;
}

.check-result {
  font-size: 13px;
  &.success {
    color: #67c23a;
  }
  &.error {
    color: #f56c6c;
  }
}

.prompts-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  .tip {
    color: #909399;
    font-size: 13px;
  }
}

.prompts-list {
  max-height: 300px;
  overflow-y: auto;
}

.prompt-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  margin-bottom: 8px;
  transition: all 0.2s;

  &:hover {
    border-color: #c0c4cc;
  }

  &.active {
    border-color: #409eff;
    background-color: #ecf5ff;
  }

  .prompt-name {
    margin-left: 8px;
  }

  .prompt-actions {
    display: flex;
    gap: 4px;
  }
}

.form-tip {
  color: #909399;
  font-size: 12px;
  margin-left: 12px;
}

.sync-actions {
  margin-top: 20px;
  display: flex;
  gap: 12px;
}

.conflict-buttons {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

.cursor-pointer {
  cursor: pointer;
}
</style>
