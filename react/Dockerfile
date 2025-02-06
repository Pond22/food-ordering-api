FROM node:18-alpine

WORKDIR /app

COPY package*.json ./

RUN npm install

COPY . .

EXPOSE 5173

# เพิ่ม --host=0.0.0.0 เพื่อให้รับการเชื่อมต่อจากภายนอก
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]