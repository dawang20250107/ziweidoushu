import type { NextConfig } from "next";

// 开发期将 /api 代理到 Go 服务;生产由网关/Nginx 统一路由。
const apiBase = process.env.API_PROXY_TARGET ?? "http://localhost:8080";

const nextConfig: NextConfig = {
  reactStrictMode: true,
  async rewrites() {
    return [
      { source: "/api/:path*", destination: `${apiBase}/api/:path*` },
    ];
  },
};

export default nextConfig;
