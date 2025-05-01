/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  swcMinify: true,
  eslint: {
    ignoreDuringBuilds: true,
  },
  typescript: {
    ignoreBuildErrors: true,
  },
  images: {
    unoptimized: true,
  },
  // This ensures Next.js doesn't interfere with the Vite build
  output: "export",
  // Ensure all static assets are properly served
  assetPrefix: ".",
  trailingSlash: true,
}

module.exports = nextConfig
