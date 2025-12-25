<template>
  <router-view></router-view>
</template>

<script lang="ts">
import { defineComponent, onMounted } from "vue";
import ThemeStore from "@/store/theme";

export default defineComponent({
  name: "App",
  setup() {
    onMounted(() => {
      // Initialize theme on app load
      const storedTheme = ThemeStore.getStoredTheme();
      ThemeStore.applyTheme(storedTheme);

      // Listen for system theme changes when in system mode
      ThemeStore.setupSystemThemeListener(() => {
        const currentTheme = ThemeStore.getStoredTheme();
        if (currentTheme === "system") {
          ThemeStore.applyTheme("system");
        }
      });
    });
  },
});
</script>

<style>
#app {
  position: fixed;
  left: 0;
  right: 0;
  top: 0;
  bottom: 0;
}
</style>