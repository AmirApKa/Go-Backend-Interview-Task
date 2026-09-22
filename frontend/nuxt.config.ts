import tailwindcss from '@tailwindcss/vite'

export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },

  css: ['~/assets/css/main.css'],

  vite: {
    plugins: [tailwindcss()]
  },

  devServer: {
    host: '127.0.0.1',
    port: 3000
  },

  runtimeConfig: {
    public: {
      apiBase: 'http://localhost:8080'
    }
  }
})