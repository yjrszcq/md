import Token from "./token";
import localforage from "localforage";

const store = localforage.createInstance({
  name: "ai-config",
});

// 获取配置存储key
const getConfigKey = () => {
  return "AIConfig_" + Token.getName();
};

// 获取模型缓存key
const getModelsKey = () => {
  return "AIModels_" + Token.getName();
};

// 获取侧边栏宽度key
const getSidebarWidthKey = () => {
  return "AISidebarWidth";
};

// 获取会话历史key
const getChatHistoryKey = () => {
  return "AIChatHistory_" + Token.getName();
};

// 默认配置
const defaultConfig: AIConfig = {
  baseUrl: "",
  apiKey: "",
  model: "",
  systemPrompts: [],
  currentPromptId: "",
  systemPromptEnabled: false,
  docContextEnabled: false,
  syncEnabled: false,
};

class AIConfigStore {
  /**
   * 保存AI配置到本地
   */
  setConfig(config: AIConfig): Promise<AIConfig> {
    return store.setItem<AIConfig>(getConfigKey(), JSON.parse(JSON.stringify(config)));
  }

  /**
   * 获取本地AI配置
   */
  async getConfig(): Promise<AIConfig> {
    try {
      const config = await store.getItem<AIConfig>(getConfigKey());
      if (config) {
        // 确保systemPrompts是数组
        if (!config.systemPrompts) {
          config.systemPrompts = [];
        }
        return config;
      }
      return { ...defaultConfig };
    } catch {
      return { ...defaultConfig };
    }
  }

  /**
   * 删除本地AI配置
   */
  removeConfig(): Promise<void> {
    return store.removeItem(getConfigKey());
  }

  /**
   * 缓存模型列表
   */
  setModels(models: ModelInfo[]): Promise<{ models: ModelInfo[]; timestamp: number }> {
    // 深拷贝以避免 IndexedDB 克隆错误
    const data = JSON.parse(JSON.stringify({
      models,
      timestamp: Date.now(),
    }));
    return store.setItem(getModelsKey(), data);
  }

  /**
   * 获取缓存的模型列表（15分钟内有效）
   */
  async getModels(): Promise<ModelInfo[] | null> {
    try {
      const cached = await store.getItem<{ models: ModelInfo[]; timestamp: number }>(getModelsKey());
      if (cached && Date.now() - cached.timestamp < 15 * 60 * 1000) {
        return cached.models;
      }
      return null;
    } catch {
      return null;
    }
  }

  /**
   * 清除模型缓存
   */
  removeModels(): Promise<void> {
    return store.removeItem(getModelsKey());
  }

  /**
   * 获取侧边栏宽度
   */
  getSidebarWidth(): number {
    const width = localStorage.getItem(getSidebarWidthKey());
    return width ? parseInt(width) : 380;
  }

  /**
   * 设置侧边栏宽度
   */
  setSidebarWidth(width: number): void {
    localStorage.setItem(getSidebarWidthKey(), String(width));
  }

  /**
   * 保存聊天历史
   */
  async setChatHistory(history: TaskBlock[]): Promise<TaskBlock[]> {
    // 深拷贝以避免 IndexedDB 克隆错误
    const data = JSON.parse(JSON.stringify(history));
    return store.setItem(getChatHistoryKey(), data);
  }

  /**
   * 获取聊天历史
   */
  async getChatHistory(): Promise<TaskBlock[]> {
    try {
      const history = await store.getItem<TaskBlock[]>(getChatHistoryKey());
      return history || [];
    } catch {
      return [];
    }
  }

  /**
   * 清除聊天历史
   */
  removeChatHistory(): Promise<void> {
    return store.removeItem(getChatHistoryKey());
  }

  /**
   * 导出配置为JSON
   */
  async exportConfig(): Promise<AIConfigExport> {
    const config = await this.getConfig();
    return {
      version: "1.0",
      exportTime: Date.now(),
      config,
    };
  }

  /**
   * 从JSON导入配置
   */
  async importConfig(data: AIConfigExport): Promise<void> {
    if (data.version && data.config) {
      await this.setConfig(data.config);
    } else {
      throw new Error("无效的配置文件格式");
    }
  }

  /**
   * 清除所有AI相关缓存（用于登出/登录时）
   */
  async clearAll(): Promise<void> {
    // 清除 localforage 中的所有数据
    await store.clear();
    // 清除 localStorage 中的AI相关数据
    const keysToRemove: string[] = [];
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i);
      if (key && key === "AISidebarWidth") {
        keysToRemove.push(key);
      }
    }
    keysToRemove.forEach((key) => localStorage.removeItem(key));
  }
}

export default new AIConfigStore();
