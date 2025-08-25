# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY apps/frontend/package.json ./
COPY apps/frontend/bun.lock* ./

# Install dependencies using npm install (since we don't have package-lock.json)
RUN npm install

# Copy source code
COPY apps/frontend/ .

# Build the application
RUN npm run build

# Production stage
FROM node:20-alpine

WORKDIR /app

# Copy built application
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/public ./public
COPY --from=builder /app/package.json ./
COPY --from=builder /app/node_modules ./node_modules

# Expose port
EXPOSE 3000

# Start the application
CMD ["npm", "start"]