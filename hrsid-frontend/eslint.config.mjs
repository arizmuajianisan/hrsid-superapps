// eslint.config.mjs
// @ts-check
import withNuxt from "./.nuxt/eslint.config.mjs";

export default withNuxt(
  // Your custom manual overrides go here, not inside a 'stylistic' object
  {
    rules: {
      // Add any specific rule overrides here if necessary
      // Example: 'no-console': 'warn'
    },
  },
);
