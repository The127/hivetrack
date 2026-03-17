import pluginVue from "eslint-plugin-vue";
import prettierConfig from "@vue/eslint-config-prettier";

export default [
  {
    ignores: ["dist/**", "node_modules/**"],
  },
  ...pluginVue.configs["flat/recommended"],
  prettierConfig,
  {
    rules: {
      // Component names: allow single-word names for top-level views/layouts
      "vue/multi-word-component-names": "off",

      // Enforce Composition API — no Options API
      "vue/component-api-style": ["error", ["script-setup", "composition"]],

      // No v-html (XSS risk)
      "vue/no-v-html": "error",

      // Consistent attribute ordering in templates
      "vue/attributes-order": ["warn", { alphabetical: false }],

      // No unused vars
      "no-unused-vars": ["error", { argsIgnorePattern: "^_" }],

      // No console.log in committed code
      "no-console": ["warn", { allow: ["warn", "error"] }],

      // Prefer const
      "prefer-const": "error",
    },
  },
];
