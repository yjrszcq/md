import { defineComponent, Fragment } from "vue";
import { MdEditor, NormalToolbar } from "md-editor-v3";
import "md-editor-v3/lib/style.css";
import { generateId } from "./config";
import SvgIcon from "@/components/svg-icon";
import { ChatDotSquare } from "@element-plus/icons-vue";

export default defineComponent({
  name: "MdEditor",
  props: {
    showAiButton: {
      type: Boolean,
      default: false,
    },
  },
  emits: ["export", "aiToggle"],
  setup(props, { emit }) {
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
            ...(props.showAiButton ? [1] : []),
          ]}
          previewTheme="cyanosis"
          codeTheme="github"
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
              {props.showAiButton && (
                <NormalToolbar
                  title="AI 助手"
                  onClick={aiClick}
                  trigger={
                    <div class="md-editor-icon">
                      <ChatDotSquare />
                    </div>
                  }
                ></NormalToolbar>
              )}
            </Fragment>
          }
        />
      );
    };
  },
});
