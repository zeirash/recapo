const createNextIntlPlugin = require('next-intl/plugin');
const withNextIntl = createNextIntlPlugin('./src/i18n.ts');

/** @type {import('next').NextConfig} */
const nextConfig = {
  experimental: {
    appDir: true,
  },
  sassOptions: {
    includePaths: ['./src/styles'],
  },
  env: {
    API_BASE_URL: process.env.API_BASE_URL || 'http://localhost:3000',
  },
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: `${process.env.API_BASE_URL || 'http://localhost:3000'}/:path*`,
      },
    ];
  },
};

module.exports = withNextIntl(nextConfig);
