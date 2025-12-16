// AI 系统提示词
interface SystemPrompt {
  id: string;
  name: string;
  content: string;
  isActive: boolean;
}

// AI 配置
interface AIConfig {
  baseUrl: string;
  apiKey: string;
  model: string;
  systemPrompts: SystemPrompt[];
  currentPromptId: string;
  systemPromptEnabled: boolean;
  docContextEnabled: boolean;
  panelEnabled: boolean;
}

// 聊天消息
interface ChatMessage {
  role: "user" | "assistant" | "system";
  content: string;
}

// 聊天请求（前端直接调用 AI API 使用）
interface ChatRequest {
  messages: ChatMessage[];
  docTitle?: string;
  docContent?: string;
}

// 聊天响应选项
interface ChatChoice {
  index: number;
  message: {
    role: string;
    content: string;
  };
  finish_reason: string;
}

// 聊天响应
interface ChatResponse {
  id: string;
  object: string;
  created: number;
  model: string;
  choices: ChatChoice[];
  usage: {
    prompt_tokens: number;
    completion_tokens: number;
    total_tokens: number;
  };
}

// 模型信息
interface ModelInfo {
  id: string;
  object: string;
  created: number;
  owned_by: string;
}

// 模型列表响应
interface ModelListResponse {
  object: string;
  data: ModelInfo[];
}

// 检查配置请求
interface CheckConfigRequest {
  baseUrl: string;
  apiKey: string;
}

// 检查配置响应
interface CheckConfigResponse {
  valid: boolean;
  message: string;
  latency: number;
}

// 任务块
interface TaskBlock {
  id: string;
  userTask: string;
  reasoning: string;
  output: string;
  status: "pending" | "processing" | "completed" | "error";
  timestamp: number;
}

// 配置导出格式
interface AIConfigExport {
  version: string;
  exportTime: number;
  config: AIConfig;
}

// AI对话记录
interface AIConversation {
  id: string;
  title: string;
  content: string; // JSON存储的TaskBlock[]
  createTime: number;
  updateTime: number;
}

// AI对话列表项
interface AIConversationListItem {
  id: string;
  title: string;
  createTime: number;
  updateTime: number;
}

// 对话请求条件
interface AIConversationCondition {
  id?: string;
  title?: string;
  content?: string;
  keyword?: string;
}
