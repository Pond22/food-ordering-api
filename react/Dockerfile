# Build stage
FROM node:18-alpine as build
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

# Production stage
FROM nginx:alpine
# คัดลอกไฟล์ static จาก build stage
COPY --from=build /app/dist /usr/share/nginx/html
# คัดลอกไฟล์ nginx.conf ที่ปรับแต่งไว้
#COPY nginx/nginx.conf /etc/nginx/conf.d/default.conf
# เปิดพอร์ตทั้ง HTTP และ HTTPS
EXPOSE 80 443
CMD ["nginx", "-g", "daemon off;"]
