import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  /* config options here */
  output: 'standalone',
  // Enable runtime configuration for environment variables
  publicRuntimeConfig: {
    apiUrl: process.env.API_URL || 'http://localhost:8080/api',
  },
  serverRuntimeConfig: {
    // Server-side only environment variables
  },
};

export default nextConfig;
