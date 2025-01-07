import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    host: '0.0.0.0', // ทำให้สามารถเข้าถึงได้จากอุปกรณ์อื่นในเครือข่าย
    port: 5173,       // ใช้พอร์ตที่ต้องการ
  },
})


