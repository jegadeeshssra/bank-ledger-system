import path from "path"
import tailwindcss from "@tailwindcss/vite"
import { defineConfig, loadEnv } from "vite"
import react from "@vitejs/plugin-react"

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "")

  const frontendPort = Number(env.VITE_FRONTEND_PORT || 5173)
  const backendProtocol = env.VITE_BACKEND_PROTOCOL || "https"
  const backendDomain = env.VITE_BACKEND_DOMAIN
  const backendPort = env.VITE_BACKEND_PORT || "8081"
  const backendTarget = `${backendProtocol}://${backendDomain}:${backendPort}`

  return {
    plugins: [react(), tailwindcss()],
    resolve: {
      alias: {
        "@": path.resolve(__dirname, "./src"),
      },
    },
    server: {
      port: frontendPort,
      proxy: {
        "/api": {
          target: backendTarget,
          changeOrigin: true,
        },
      },
    },
  }
})
