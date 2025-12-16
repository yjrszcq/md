<template>
  <div class="page-login">
    <text-rain class="text-rain-background"></text-rain>
    <div class="title-view" title="查看公开文档" @click="publishClick">云文档</div>
    <transition name="fade">
      <div class="content-view" v-if="isLogin">
        <form>
          <el-input
            class="input-view"
            v-model.trim="inputData.name"
            size="large"
            clearable
            placeholder="请输入用户名"
            @keyup.enter.native="loginClick"
          ></el-input>
          <el-input
            class="input-view"
            v-model.trim="inputData.password"
            size="large"
            type="password"
            clearable
            placeholder="请输入密码"
            @keyup.enter.native="loginClick"
          ></el-input>
        </form>
        <el-button class="register-button" type="primary" link text size="small" @click="registerClick" :disabled="loading">注册</el-button>
        <el-button class="login-button" size="large" type="primary" @click="loginClick" :disabled="loading">登录</el-button>
      </div>
      <div class="content-view" v-else>
        <form>
          <el-input class="input-view" v-model.trim="inputData.name" size="large" clearable placeholder="请输入用户名"></el-input>
          <el-input class="input-view" v-model.trim="inputData.password" size="large" type="password" clearable placeholder="请输入密码"></el-input>
          <el-input
            class="input-view"
            v-model.trim="inputData.confirmPassword"
            size="large"
            type="password"
            clearable
            placeholder="请再次输入密码"
          ></el-input>
        </form>
        <el-button class="register-button" type="primary" link text size="small" @click="registerClick" :disabled="loading">返回登录</el-button>
        <el-button class="login-button" size="large" type="primary" @click="loginClick" :disabled="loading">注册</el-button>
      </div>
    </transition>
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted } from "vue";
import Token from "@/store/token";
import TokenApi from "@/api/token";
import { ElMessage } from "element-plus";
import { useRouter } from "vue-router";
import TextRain from "@/components/text-rain/index.vue";
import AIConfigStore from "@/store/ai-config";

const hostUrl = ref(location.origin);
const router = useRouter();
const loading = ref(false);
// 登录/注册
const isLogin = ref(true);
// 输入框数据
const inputData = ref({
  name: "",
  password: "",
  confirmPassword: "",
});

onMounted(() => {
  let nameCache = Token.getName();
  if (nameCache) {
    inputData.value.name = nameCache;
  }
});

/**
 * 点击注册切换按钮
 */
const registerClick = () => {
  inputData.value.name = "";
  inputData.value.password = "";
  inputData.value.confirmPassword = "";
  isLogin.value = !isLogin.value;
};

/**
 * 点击登录按钮
 */
const loginClick = () => {
  if (!inputData.value.name) {
    ElMessage.warning("请输入用户名");
    return;
  }
  if (!inputData.value.password) {
    ElMessage.warning("请输入密码");
    return;
  }
  if (isLogin.value) {
    // 登录
    loading.value = true;
    TokenApi.signIn(inputData.value.name, inputData.value.password)
      .then(async (res) => {
        // 登录时清除AI缓存，避免残留其他用户的数据
        await AIConfigStore.clearAll();
        Token.setToken(res.data);
        router.push({ name: "layout" });
      })
      .finally(() => {
        loading.value = false;
      });
  } else {
    // 注册
    if (inputData.value.password !== inputData.value.confirmPassword) {
      ElMessage.warning("两次密码不一致");
      return;
    }
    loading.value = true;
    TokenApi.signUp(inputData.value.name, inputData.value.password)
      .then(() => {
        ElMessage.success("注册成功");
        inputData.value.password = "";
        inputData.value.confirmPassword = "";
        isLogin.value = true;
      })
      .finally(() => {
        loading.value = false;
      });
  }
};

/**
 * 点击公开文档
 */
const publishClick = () => {
  let url = hostUrl.value + "/#/open/publish";
  window.open(url, "_blank");
};
</script>

<style lang="scss" scoped>
.page-login {
  position: fixed;
  width: 100%;
  height: 100%;
  left: 0;
  right: 0;
  top: 0;
  bottom: 0;
  background: linear-gradient(0deg, rgba(255, 238, 213, 1) 0%, rgba(148, 210, 233, 1) 70%);
  display: flex;
  flex-direction: column;
  align-items: center;
  user-select: none;
  overflow: auto;
  .title-view {
    margin-top: 10vh;
    font-size: 24px;
    font-weight: bold;
    color: #3f3f3f;
    cursor: pointer;
  }
  .content-view {
    margin-top: 20px;
    width: 300px;
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    background: rgba(255, 255, 255, 0.2);
    padding: 30px 20px;
    .input-view {
      width: 100%;
      margin-top: 10px;
    }
    .register-button {
      margin: 5px 0;
    }
    .login-button {
      width: 100%;
    }
    .register-button + .login-button {
      margin-left: 0;
    }
  }
  .text-rain-background {
    position: absolute;
    z-index: -1;
  }
}
</style>
