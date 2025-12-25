import { defineComponent, ref, onMounted, onUnmounted, computed } from "vue";
import { MdPreview, MdCatalog } from "md-editor-v3";
import "md-editor-v3/lib/preview.css";
import { generateId } from "./config";

export default defineComponent({
  name: "MdPreview",
  props: {
    content: {
      type: String,
      default: "",
    },
  },
  setup(props) {
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

    return () => {
      return (
        <div class="md-preview_component">
          <MdPreview
            class="preview-view"
            modelValue={props.content}
            editorId="MdPreview"
            theme={theme.value}
            previewTheme={previewTheme.value}
            codeTheme={codeTheme.value}
            mdHeadingId={generateId}
            noMermaid
          />
          <el-scrollbar class="catalog-view">
            <MdCatalog editorId="MdPreview" mdHeadingId={generateId} theme={theme.value} />
          </el-scrollbar>
        </div>
      );
    };
  },
});
