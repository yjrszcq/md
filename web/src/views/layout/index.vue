<template>
  <div class="page-layout">
    <div class="top-view">
      <div class="left-view">
        <div class="title-view">
          <span :style="isDocument ? 'cursor: pointer' : ''" :title="isDocument ? (isStretch ? '收起侧栏' : '弹出侧栏') : ''" @click="iconClick">
            <svg-icon name="md" customStyle="width: 20px; height: 20px; margin: 5px 5px 0 0"></svg-icon>
          </span>
          <span :style="isDocument ? 'cursor: pointer' : ''" :title="isDocument ? (onlyPreview ? '编辑模式' : '预览模式') : ''" @click="titleClick">
            云文档
          </span>
        </div>
        <div class="menu-view">
          <router-link to="/document">文档</router-link>
          <router-link to="/picture">图片</router-link>
          <router-link to="/tool">工具</router-link>
        </div>
      </div>
      <el-dropdown class="right-view">
        <div class="text-view">{{ name }}</div>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item style="user-select: none" @click="toggleTheme">{{ themeLabel }}</el-dropdown-item>
            <el-dropdown-item style="user-select: none" @click="publishClick">公开文档</el-dropdown-item>
            <el-dropdown-item style="user-select: none" @click="dialogVisible = true">修改密码</el-dropdown-item>
            <el-dropdown-item style="user-select: none" @click="logout">退出登录</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
    <router-view class="content-view" :onlyPreview="onlyPreview" :isStretch="isStretch"></router-view>
    <el-dialog v-model="dialogVisible" title="修改密码" width="400px" :show-close="false" :before-close="dialogClose">
      <form>
        <el-input v-model.trim="form.password" size="large" type="password" clearable placeholder="请输入原密码"></el-input>
        <el-input style="margin: 10px 0" v-model.trim="form.newPassword" size="large" type="password" clearable placeholder="请输入新密码"></el-input>
        <el-input v-model.trim="form.confirmPassword" size="large" type="password" clearable placeholder="请再次输入密码"></el-input>
      </form>
      <template #footer>
        <span class="dialog-footer">
          <el-button :loading="dialogLoading" @click="dialogClose">取消</el-button>
          <el-button type="primary" :loading="dialogLoading" @click="updatePassword">保存</el-button>
        </span>
      </template>
    </el-dialog>
    <ai-config-dialog v-model="aiConfigVisible" @configChanged="onAiConfigChanged" />
  </div>
</template>

<script lang="ts" setup>
import { ref, watch, onMounted, onUnmounted } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import SvgIcon from "@/components/svg-icon";
import AiConfigDialog from "@/components/ai-config-dialog/index.vue";
import Token from "@/store/token";
import TokenApi from "@/api/token";
import UserApi from "@/api/user";
import AIApi from "@/api/ai";
import router from "@/router";
import DocCache from "@/store/doc-cache";
import AIConfigStore from "@/store/ai-config";
import ThemeStore, { type ThemeMode } from "@/store/theme";

const hostUrl = ref(location.origin);
const name = ref(Token.getName());
const dialogVisible = ref(false);
const dialogLoading = ref(false);
const aiConfigVisible = ref(false);
const form = ref({ password: "", newPassword: "", confirmPassword: "" });
const onlyPreview = ref(true);
const isStretch = ref(true);
const isDocument = ref(router.currentRoute.value.name === "document");

// Theme state
const themeMode = ref<ThemeMode>(ThemeStore.getStoredTheme());
const themeLabel = ref(ThemeStore.getThemeLabel(themeMode.value));
let removeSystemThemeListener: (() => void) | null = null;

// Listen for open-ai-config event from ai-sidebar
const handleOpenAiConfig = () => {
  aiConfigVisible.value = true;
};

// Sync AI config from server on page load/refresh
const syncAiConfigFromServer = async () => {
  try {
    const serverRes = await AIApi.getConfigFull();
    const serverConfig = serverRes.data;
    const localConfig = await AIConfigStore.getConfig();

    if (serverConfig && serverConfig.syncEnabled) {
      // Server sync is enabled, update local config
      await AIConfigStore.setConfig(serverConfig);
      window.dispatchEvent(new CustomEvent("ai-config-changed", { detail: serverConfig }));
    } else if (localConfig.syncEnabled && (!serverConfig || !serverConfig.syncEnabled)) {
      // Server sync is disabled but local is enabled, disable local sync
      localConfig.syncEnabled = false;
      await AIConfigStore.setConfig(localConfig);
      window.dispatchEvent(new CustomEvent("ai-config-changed", { detail: localConfig }));
    }
  } catch {
    // Ignore errors, server might not have config yet
  }
};

onMounted(() => {
  window.addEventListener("open-ai-config", handleOpenAiConfig);
  syncAiConfigFromServer();

  // Setup system theme listener for auto-switching when in system mode
  removeSystemThemeListener = ThemeStore.setupSystemThemeListener(() => {
    if (themeMode.value === "system") {
      ThemeStore.applyTheme("system");
    }
  });
});

onUnmounted(() => {
  window.removeEventListener("open-ai-config", handleOpenAiConfig);
  if (removeSystemThemeListener) {
    removeSystemThemeListener();
  }
});

watch(
  () => router.currentRoute.value.name,
  (val) => {
    if (val === "document") {
      isDocument.value = true;
    } else {
      isDocument.value = false;
    }
  }
);

/**
 * 点击退出登录
 */
const logout = () => {
  ElMessageBox.confirm("是否退出登录？", "提示", {
    confirmButtonText: "退出登录",
    cancelButtonText: "取消",
    type: "info",
  })
    .then(async () => {
      DocCache.removeDoc();
      await AIConfigStore.clearAll();
      TokenApi.signOut();
      Token.removeToken();
    })
    .catch(() => {});
};

/**
 * 更新密码
 */
const updatePassword = () => {
  if (form.value.password === "" || form.value.newPassword === "" || form.value.confirmPassword === "") {
    ElMessage.warning("请填写密码");
    return;
  }
  if (form.value.newPassword !== form.value.confirmPassword) {
    ElMessage.warning("两次密码不一致");
    return;
  }

  dialogLoading.value = true;
  UserApi.updatePassword(form.value.password, form.value.newPassword)
    .then((res) => {
      ElMessage.success("修改成功");
      dialogLoading.value = false;
      dialogClose();
    })
    .catch(() => {
      dialogLoading.value = false;
    });
};

/**
 * 弹窗关闭
 */
const dialogClose = () => {
  if (dialogLoading.value) {
    return;
  }
  dialogVisible.value = false;
  form.value.password = "";
  form.value.newPassword = "";
  form.value.confirmPassword = "";
};

/**
 * 点击标题
 */
const titleClick = () => {
  if (isDocument.value) {
    onlyPreview.value = !onlyPreview.value;
  }
};

/**
 * 点击标题图标
 */
const iconClick = () => {
  if (isDocument.value) {
    isStretch.value = !isStretch.value;
  }
};

/**
 * 点击公开文档
 */
const publishClick = () => {
  let url = hostUrl.value + "/#/open/publish";
  window.open(url, "_blank");
};

/**
 * Toggle theme mode: system -> light -> dark -> system
 */
const toggleTheme = () => {
  themeMode.value = ThemeStore.cycleTheme(themeMode.value);
  themeLabel.value = ThemeStore.getThemeLabel(themeMode.value);
  ThemeStore.setStoredTheme(themeMode.value);
  ThemeStore.applyTheme(themeMode.value);
};

/**
 * AI 配置变化
 */
const onAiConfigChanged = (config: AIConfig) => {
  // 通过自定义事件通知其他组件
  window.dispatchEvent(new CustomEvent("ai-config-changed", { detail: config }));
};
</script>

<style lang="scss" scoped>
.page-layout {
  position: fixed;
  width: 100%;
  height: 100%;
  left: 0;
  right: 0;
  top: 0;
  bottom: 0;
  background-color: var(--bg-primary);
}
.top-view {
  height: 49px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  user-select: none;
  border-bottom: 1px var(--header-border) solid;
  padding: 0 2%;
  white-space: nowrap;
  background-color: var(--header-bg);
  .left-view {
    display: flex;
    align-items: center;
    height: 100%;
    flex: 1;
    .title-view {
      font-size: 18px;
      font-weight: bold;
      display: flex;
      align-items: center;
      color: var(--text-primary);
    }
    .menu-view {
      margin-left: 2%;
      display: flex;
      align-items: center;
      height: 100%;
      font-size: 14px;
      a {
        text-decoration: none;
        color: var(--text-primary);
        height: 100%;
        display: flex;
        align-items: center;
        justify-content: center;
        border-bottom: 2px solid transparent;
        box-sizing: border-box;
        padding: 4px 20px 0 20px;
        transition: all 0.3s;
      }
      a:hover {
        background: rgba(0, 148, 193, 0.1);
      }
      .router-link-active {
        color: #0094c1;
        border-color: #0094c1;
      }
    }
  }
  .right-view {
    height: 100%;
    .text-view {
      height: 100%;
      display: flex;
      align-items: center;
      justify-content: center;
      cursor: pointer;
      padding: 0 20px;
      transition: all 0.3s;
      color: var(--text-primary);
      outline: none;
    }
    .text-view:hover {
      color: #0094c1;
    }
  }
}
.content-view {
  width: 100%;
  height: calc(100% - 50px);
  overflow: auto;
  background-color: var(--content-bg);
}
</style>

<style lang="scss">
// Dark theme: menu active color should be yellow (unscoped to work with data-theme)
[data-theme="dark"] {
  .top-view {
    .left-view .menu-view {
      a:hover {
        background: rgba(230, 162, 60, 0.1) !important;
      }
      .router-link-active {
        color: #f0c060 !important;
        border-color: #f0c060 !important;
      }
    }
    .right-view .text-view:hover {
      color: #f0c060 !important;
    }
  }
}
</style>
