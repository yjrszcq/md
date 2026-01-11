import { defineComponent, ref, computed, onMounted, onUnmounted } from "vue";
import { Codemirror } from "vue-codemirror";
import { json } from "@codemirror/lang-json";
import { keymap, EditorView, highlightSpecialChars, drawSelection } from "@codemirror/view";
import { HighlightStyle, syntaxHighlighting } from "@codemirror/language";
import { tags as t } from "@lezer/highlight";
import "./index.scss";

export default defineComponent({
  name: "CodemirrorEditor",
  props: {
    noRadius: {
      type: Boolean,
      default: false,
    },
    extensions: {
      type: Array as () => any[],
      default: () => [],
    },
  },
  emits: ["save"],
  setup(props, { emit }) {
    const isDark = ref(document.documentElement.getAttribute("data-theme") === "dark");
    const observer = new MutationObserver((mutations) => {
      mutations.forEach((mutation) => {
        if (mutation.attributeName === "data-theme") {
          isDark.value = document.documentElement.getAttribute("data-theme") === "dark";
        }
      });
    });

    onMounted(() => observer.observe(document.documentElement, { attributes: true }));
    onUnmounted(() => observer.disconnect());

    const extensions = computed(() => {
      const warmHighlight = HighlightStyle.define([
        { tag: t.string, color: "rgb(229, 192, 123)" }, // strings
        { tag: t.number, color: "rgb(224, 108, 117)" }, // numbers
        { tag: t.bool, color: "rgb(78, 125, 228)" }, // booleans
        { tag: t.null, color: "rgb(198, 120, 221)" }, // null
        { tag: t.propertyName, color: "#c2c2c2" }, // field names
        { tag: t.punctuation, color: "#a9b7c6" },
      ]);

      const tokenTheme = EditorView.theme({
        ".cmt-string, .cm-string": { color: "rgb(229, 192, 123)" },
        ".cmt-number, .cm-number": { color: "rgb(224, 108, 117)" },
        ".cmt-atom, .cm-atom": { color: "rgb(78, 125, 228)" }, // bool/null both map to atom in some modes
        ".cmt-bool, .cm-bool": { color: "rgb(78, 125, 228)" },
        ".cmt-null": { color: "rgb(198, 120, 221)" },
        ".cmt-propertyName, .cm-property": { color: "#c2c2c2" },
        ".cmt-punctuation, .cm-punctuation, .cmt-operator, .cm-operator": { color: "#a9b7c6" },
      });

      const userExtensions = props.extensions && props.extensions.length > 0 ? props.extensions : [json()];

      const base = [
        ...userExtensions,
        drawSelection(),
        keymap.of([
          {
            key: "Ctrl-s",
            run: () => {
              emit("save");
              return true;
            },
          },
        ]),
      ];

      if (isDark.value) {
        const darkTheme = EditorView.theme(
          {
            "&": { color: "#d4d4d4", backgroundColor: "#1e1e1e" },
            ".cm-content": { caretColor: "#e6c07b" },
            ".cm-cursor, .cm-dropCursor": { borderLeftColor: "#e6c07b" },
            ".cm-activeLine": { backgroundColor: "#2a2a2a" },
            ".cm-gutters": { backgroundColor: "#1e1e1e", color: "#6b6b6b", borderRight: "1px solid #333333" },
          },
          { dark: true }
        );

        base.push(syntaxHighlighting(warmHighlight), tokenTheme, darkTheme, highlightSpecialChars());
      }

      return base;
    });

    return () => (
      <Codemirror
        style={{ width: "100%" }}
        class={props.noRadius ? "no-radius" : ""}
        autofocus
        indent-with-tab
        tab-size={2}
        extensions={extensions.value}
        key={isDark.value ? "cm-dark" : "cm-light"}
      />
    );
  },
});
