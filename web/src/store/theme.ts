// Theme store for managing application theme

export type ThemeMode = "light" | "dark" | "system";

const THEME_KEY = "md-theme-mode";

/**
 * Get stored theme mode from localStorage
 */
export function getStoredTheme(): ThemeMode {
  const stored = localStorage.getItem(THEME_KEY);
  if (stored === "light" || stored === "dark" || stored === "system") {
    return stored;
  }
  return "system";
}

/**
 * Store theme mode to localStorage
 */
export function setStoredTheme(mode: ThemeMode): void {
  localStorage.setItem(THEME_KEY, mode);
}

/**
 * Get the effective theme (resolves system preference)
 */
export function getEffectiveTheme(mode: ThemeMode): "light" | "dark" {
  if (mode === "system") {
    return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
  }
  return mode;
}

/**
 * Apply theme to document
 */
export function applyTheme(mode: ThemeMode): void {
  const effectiveTheme = getEffectiveTheme(mode);
  document.documentElement.setAttribute("data-theme", effectiveTheme);
}

/**
 * Cycle to next theme mode: system -> light -> dark -> system
 */
export function cycleTheme(current: ThemeMode): ThemeMode {
  switch (current) {
    case "system":
      return "light";
    case "light":
      return "dark";
    case "dark":
      return "system";
    default:
      return "system";
  }
}

/**
 * Get display name for theme mode
 */
export function getThemeLabel(mode: ThemeMode): string {
  switch (mode) {
    case "light":
      return "亮色主题";
    case "dark":
      return "暗色主题";
    case "system":
      return "系统主题";
    default:
      return "系统主题";
  }
}

/**
 * Setup system theme change listener
 */
export function setupSystemThemeListener(callback: () => void): () => void {
  const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
  const handler = () => callback();
  mediaQuery.addEventListener("change", handler);
  return () => mediaQuery.removeEventListener("change", handler);
}

export default {
  getStoredTheme,
  setStoredTheme,
  getEffectiveTheme,
  applyTheme,
  cycleTheme,
  getThemeLabel,
  setupSystemThemeListener,
};
