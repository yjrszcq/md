import request from "@/utils/request";

// Agent模式预设系统提示词
const AgentSystemPrompt = `你是一个专业的文档编辑AI助手。

## 重要行为准则

**只有当用户明确请求对文档进行修改、编辑、优化等操作时**，你才应该分析文档并返回结构化的修改建议。

对于普通对话（如打招呼、闲聊、与文档无关的问题），请正常回复，不要主动分析或评价文档内容，也不要输出JSON格式。

## 输出格式要求

当用户明确请求修改文档时，你的回复必须是一个有效的JSON对象，格式如下：
{
  "plan": ["步骤1说明", "步骤2说明"],
  "changes": [
    {
      "type": "replace",
      "position": "修改位置描述（如：第2段、标题下方、代码块等）",
      "oldText": "被替换的原文（用于定位）",
      "content": "新的内容"
    }
  ],
  "explanation": "修改原因说明"
}

## 修改类型说明
- replace: 替换指定位置的内容
- insert: 在指定位置插入新内容
- delete: 删除指定位置的内容

## 重要约束
1. 只有用户明确请求修改时才输出JSON格式
2. position字段要准确描述修改位置
3. oldText字段要包含足够的原文用于精确定位
4. 保持文档的整体风格和格式
5. 不要进行未请求的结构性重组`;

class AIApi {
  /**
   * 获取AI配置（从服务器）
   */
  getConfig() {
    return request<AIConfig>({
      method: "get",
      url: "/ai/config",
    });
  }

  /**
   * 获取AI配置（完整版，包含完整API Key，用于同步）
   */
  getConfigFull() {
    return request<AIConfig>({
      method: "get",
      url: "/ai/config/full",
    });
  }

  /**
   * 保存AI配置（到服务器）
   */
  saveConfig(config: AIConfig) {
    return request({
      method: "post",
      url: "/ai/config",
      data: config,
    });
  }

  /**
   * 删除AI配置（从服务器）
   */
  deleteConfig() {
    return request({
      method: "post",
      url: "/ai/config/delete",
    });
  }

  /**
   * 检查服务器AI配置是否存在
   */
  configExists() {
    return request<{ exists: boolean }>({
      method: "get",
      url: "/ai/config/exists",
    });
  }

  /**
   * 直接获取模型列表（前端直接调用AI API）
   */
  async getModelsDirect(baseUrl: string, apiKey: string): Promise<ModelListResponse> {
    const url = baseUrl.replace(/\/$/, "") + "/v1/models";
    const response = await fetch(url, {
      method: "GET",
      headers: {
        Authorization: `Bearer ${apiKey}`,
      },
    });

    if (!response.ok) {
      const text = await response.text();
      throw new Error(`获取模型列表失败: ${response.status} - ${text}`);
    }

    return response.json();
  }

  /**
   * 直接检查配置有效性（前端直接调用AI API）
   */
  async checkConfigDirect(baseUrl: string, apiKey: string): Promise<CheckConfigResponse> {
    const startTime = Date.now();
    const url = baseUrl.replace(/\/$/, "") + "/v1/models";

    try {
      const response = await fetch(url, {
        method: "GET",
        headers: {
          Authorization: `Bearer ${apiKey}`,
        },
        signal: AbortSignal.timeout(15000),
      });

      const latency = Date.now() - startTime;

      if (response.ok) {
        return { valid: true, message: "配置有效", latency };
      }

      switch (response.status) {
        case 401:
        case 403:
          return { valid: false, message: "API Key 无效或无权限", latency };
        case 404:
          return { valid: false, message: "Base URL 不正确或不兼容", latency };
        case 429:
          return { valid: false, message: "请求频率过高或额度不足", latency };
        default:
          const text = await response.text();
          return { valid: false, message: `上游错误 (${response.status}): ${text}`, latency };
      }
    } catch (err: any) {
      const latency = Date.now() - startTime;
      if (err.name === "TimeoutError" || err.message?.includes("timeout")) {
        return { valid: false, message: "请求超时，请检查网络或代理设置", latency };
      }
      return { valid: false, message: `请求失败: ${err.message}`, latency };
    }
  }

  /**
   * 直接发送聊天请求（前端直接调用AI API）
   */
  async chatDirect(
    config: AIConfig,
    messages: ChatMessage[],
    options?: {
      agentMode?: boolean;
      docTitle?: string;
      docContent?: string;
      signal?: AbortSignal;
    }
  ): Promise<ChatResponse> {
    const { agentMode, docTitle, docContent, signal } = options || {};

    // 构建完整的消息列表
    const fullMessages: ChatMessage[] = [];

    // 注入系统提示词
    let systemContent = "";
    if (agentMode) {
      // Agent模式使用预设提示词
      systemContent = AgentSystemPrompt;
    } else {
      // 普通模式使用用户配置的提示词
      const activePrompt = config.systemPrompts.find((p) => p.id === config.currentPromptId);
      if (activePrompt?.content) {
        systemContent = activePrompt.content;
      }
    }

    // 注入文档上下文
    if (config.docContextEnabled && docContent) {
      let docContext = "\n\n---\n## 文档上下文\n\n";
      docContext += "以下是用户当前正在编辑的文档内容，仅供参考。\n";
      docContext += "**重要**：只有当用户明确提及文档、询问文档相关问题、或请求对文档进行操作时，你才应该引用或分析文档内容。";
      docContext += "对于普通对话（如打招呼、闲聊、与文档无关的问题），请正常回复，不要主动提及或评价文档内容。\n\n";
      if (docTitle) {
        docContext += `文档标题：${docTitle}\n`;
      }
      docContext += `文档内容：\n\`\`\`markdown\n${docContent}\n\`\`\``;
      systemContent += docContext;
    }

    if (systemContent) {
      fullMessages.push({ role: "system", content: systemContent });
    }

    // 添加用户消息
    fullMessages.push(...messages);

    // 发送请求
    const url = config.baseUrl.replace(/\/$/, "") + "/v1/chat/completions";
    const response = await fetch(url, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${config.apiKey}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        model: config.model,
        messages: fullMessages,
        stream: false,
      }),
    });

    if (!response.ok) {
      const text = await response.text();
      throw new Error(`AI请求失败 (${response.status}): ${text}`);
    }

    return response.json();
  }

  /**
   * 流式发送聊天请求（前端直接调用AI API）
   */
  async chatStream(
    config: AIConfig,
    messages: ChatMessage[],
    options: {
      agentMode?: boolean;
      docTitle?: string;
      docContent?: string;
      onContent?: (content: string) => void;
      onReasoning?: (reasoning: string) => void;
      onDone?: () => void;
      onError?: (error: Error) => void;
      signal?: AbortSignal;
    }
  ): Promise<void> {
    const { agentMode, docTitle, docContent, onContent, onReasoning, onDone, onError, signal } = options;

    // 构建完整的消息列表
    const fullMessages: ChatMessage[] = [];

    // 注入系统提示词
    let systemContent = "";
    if (agentMode) {
      systemContent = AgentSystemPrompt;
    } else {
      const activePrompt = config.systemPrompts.find((p) => p.id === config.currentPromptId);
      if (activePrompt?.content) {
        systemContent = activePrompt.content;
      }
    }

    // 注入文档上下文
    if (config.docContextEnabled && docContent) {
      let docContext = "\n\n---\n## 文档上下文\n\n";
      docContext += "以下是用户当前正在编辑的文档内容，仅供参考。\n";
      docContext += "**重要**：只有当用户明确提及文档、询问文档相关问题、或请求对文档进行操作时，你才应该引用或分析文档内容。";
      docContext += "对于普通对话（如打招呼、闲聊、与文档无关的问题），请正常回复，不要主动提及或评价文档内容。\n\n";
      if (docTitle) {
        docContext += `文档标题：${docTitle}\n`;
      }
      docContext += `文档内容：\n\`\`\`markdown\n${docContent}\n\`\`\``;
      systemContent += docContext;
    }

    if (systemContent) {
      fullMessages.push({ role: "system", content: systemContent });
    }

    fullMessages.push(...messages);

    const url = config.baseUrl.replace(/\/$/, "") + "/v1/chat/completions";

    try {
      const response = await fetch(url, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${config.apiKey}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          model: config.model,
          messages: fullMessages,
          stream: true,
        }),
        signal,
      });

      if (!response.ok) {
        const text = await response.text();
        throw new Error(`AI请求失败 (${response.status}): ${text}`);
      }

      const reader = response.body?.getReader();
      if (!reader) {
        throw new Error("无法获取响应流");
      }

      const decoder = new TextDecoder();
      let buffer = "";

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split("\n");
        buffer = lines.pop() || "";

        for (const line of lines) {
          const trimmed = line.trim();
          if (!trimmed || trimmed === "data: [DONE]") continue;
          if (!trimmed.startsWith("data: ")) continue;

          try {
            const json = JSON.parse(trimmed.slice(6));
            const delta = json.choices?.[0]?.delta;

            if (delta) {
              // 处理思考内容 (reasoning_content 是一些模型如 DeepSeek 的字段)
              if (delta.reasoning_content && onReasoning) {
                onReasoning(delta.reasoning_content);
              }
              // 处理正常内容
              if (delta.content && onContent) {
                onContent(delta.content);
              }
            }
          } catch {
            // 忽略解析错误
          }
        }
      }

      onDone?.();
    } catch (err: any) {
      if (err.name === "AbortError") {
        return;
      }
      onError?.(err);
    }
  }
}

export default new AIApi();
