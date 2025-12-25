import { config } from "md-editor-v3";
import highlight from "highlight.js";
// Removed fixed github.css import - let md-editor-v3 manage code theme via codeTheme prop
import prettier from "prettier";
import parserMarkdown from "prettier/parser-markdown";
import cropper from "cropperjs";
import "cropperjs/dist/cropper.css";
import screenfull from "screenfull";
import katex from "katex";
import "katex/dist/katex.css";
import "./index.scss";
import { lineNumbers } from "@codemirror/view";

/**
 * 扩展：链接打开新窗口
 * @param md
 */
const targetBlankExtension = (md: any) => {
  const defaultRender =
    md.renderer.rules.link_open ||
    function (tokens: any, idx: any, options: any, env: any, self: any) {
      return self.renderToken(tokens, idx, options);
    };
  md.renderer.rules.link_open = function (tokens: any, idx: any, options: any, env: any, self: any) {
    const aIndex = tokens[idx].attrIndex("target");
    if (aIndex < 0) {
      tokens[idx].attrPush(["target", "_blank"]);
    } else {
      tokens[idx].attrs[aIndex][1] = "_blank";
    }
    return defaultRender(tokens, idx, options, env, self);
  };
};

config({
  editorConfig: {
    renderDelay: 500,
  },
  editorExtensions: {
    iconfont: "/static/md-editor/iconfont.js",
    highlight: {
      instance: highlight,
    },
    prettier: {
      prettierInstance: prettier,
      parserMarkdownInstance: parserMarkdown,
    },
    cropper: {
      instance: cropper,
    },
    screenfull: {
      instance: screenfull,
    },
    katex: {
      instance: katex,
    },
  },
  codeMirrorExtensions(_theme: any, extensions: any) {
    return [...extensions, lineNumbers()];
  },
  markdownItConfig(md) {
    md.use(targetBlankExtension);
  },
});

export const generateId = (text: any, level: any, index: any) => `heading-${index}`;
