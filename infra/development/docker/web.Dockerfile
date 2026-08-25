# 1. Build stage
FROM oven/bun:1 AS builder
WORKDIR /app

COPY web/package.json web/bun.lock ./
RUN bun install --frozen-lockfile

COPY web/ .
ENV NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY=pk_test_51U5HNiPPY9UamPfE5gw4HtKkhRAIx8q1jaBjyni8zZpJmldTyClYxLOy6gGqkTKSXT761VLlL9XaGbnxRj4kCnOA00bxvthtKo
RUN bun run build

# 2. Release stage
FROM node:20-alpine
WORKDIR /app

ENV NODE_ENV=production \
    HOSTNAME=0.0.0.0 \
    PORT=3000

COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static

EXPOSE 3000

CMD ["node", "server.js"]
