/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
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
  // Ensure all static assets are properly served
  assetPrefix: ".",
  trailingSlash: true,

  // Add rewrites for proxying API requests during development
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: 'http://localhost:3000/api/:path*', // Proxy to Backend
      },
    ]
  },
}

module.exports = nextConfig
