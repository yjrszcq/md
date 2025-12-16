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
  agentEnabled: boolean;
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
  agentMode?: boolean;
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

// Agent 变更
interface AgentChange {
  type: "replace" | "insert" | "delete";
  position: string;
  oldText?: string;
  content: string;
  // 应用状态（运行时使用，不持久化）
  applied?: boolean;
  // 撤回所需的原始内容（运行时使用）
  undoData?: {
    originalContent: string;
    appliedContent: string;
  };
}

// Agent 响应
interface AgentResponse {
  plan: string[];
  changes: AgentChange[];
  explanation: string;
}

// 任务块
interface TaskBlock {
  id: string;
  userTask: string;
  reasoning: string;
  output: string;
  agentResponse?: AgentResponse;
  status: "pending" | "processing" | "completed" | "error";
  timestamp: number;
  mode?: "chat" | "agent";
}

// 配置导出格式
interface AIConfigExport {
  version: string;
  exportTime: number;
  config: AIConfig;
}
