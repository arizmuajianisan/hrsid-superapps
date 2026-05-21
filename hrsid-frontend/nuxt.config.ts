// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  modules: ["@nuxt/eslint", "@nuxt/ui", "@pinia/nuxt"],

  ssr: false,

  devtools: {
    enabled: true,
  },

  css: ["~/assets/css/main.css"],

  runtimeConfig: {
    public: {
      apiBase: "http://localhost:4000",
    },
  },

  devServer: {
    port: 5173,
  },

  compatibilityDate: "2025-01-15",

  eslint: {
    config: {
      // This enables the stylistic rules automatically
      stylistic: {
        semi: true,
        commaDangle: "never",
        braceStyle: "1tbs",
      },
    },
  },
});
