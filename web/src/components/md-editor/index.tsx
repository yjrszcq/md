import { defineComponent, Fragment, ref, onMounted, onUnmounted, computed } from "vue";
import { MdEditor, NormalToolbar } from "md-editor-v3";
import "md-editor-v3/lib/style.css";
import { generateId } from "./config";
import SvgIcon from "@/components/svg-icon";
import { ChatDotSquare } from "@element-plus/icons-vue";

export default defineComponent({
  name: "MdEditor",
  emits: ["export", "aiToggle"],
  setup(props, { emit }) {
    const isDark = ref(document.documentElement.getAttribute("data-theme") === "dark");

    // Watch for theme changes
    const observer = new MutationObserver((mutations) => {
      mutations.forEach((mutation) => {
        if (mutation.attributeName === "data-theme") {
          isDark.value = document.documentElement.getAttribute("data-theme") === "dark";
        }
      });
    });

    onMounted(() => {
      observer.observe(document.documentElement, { attributes: true });
    });

    onUnmounted(() => {
      observer.disconnect();
    });

    const theme = computed(() => (isDark.value ? "dark" : "light"));
    const codeTheme = computed(() => (isDark.value ? "atom" : "github"));
    const previewTheme = computed(() => (isDark.value ? "default" : "cyanosis"));

    const exportClick = () => {
      emit("export");
    };
    const aiClick = () => {
      emit("aiToggle");
    };
    return () => {
      return (
        <MdEditor
          toolbars={[
            "bold",
            "underline",
            "italic",
            "-",
            "title",
            "strikeThrough",
            "sub",
            "sup",
            "quote",
            "unorderedList",
            "orderedList",
            "task",
            "-",
            "codeRow",
            "code",
            "link",
            "image",
            "table",
            "katex",
            "-",
            "revoke",
            "next",
            "save",
            0,
            "=",
            "prettier",
            "pageFullscreen",
            "preview",
            "htmlPreview",
            "catalog",
            1,
          ]}
          theme={theme.value}
          previewTheme={previewTheme.value}
          codeTheme={codeTheme.value}
          codeFoldable={false}
          showCodeRowNumber
          mdHeadingId={generateId}
          noMermaid
          defToolbars={
            <Fragment>
              <NormalToolbar
                title="导出"
                onClick={exportClick}
                trigger={
                  <div class="md-editor-icon">
                    <SvgIcon className="icon-download" name="download"></SvgIcon>
                  </div>
                }
              ></NormalToolbar>
              <NormalToolbar
                title="AI 助手"
                onClick={aiClick}
                trigger={
                  <div class="md-editor-icon">
                    <ChatDotSquare />
                  </div>
                }
              ></NormalToolbar>
            </Fragment>
          }
        />
      );
    };
  },
});
