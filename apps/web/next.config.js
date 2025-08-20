/** @type {import('next').NextConfig} */
const nextConfig = {
    typescript: {
        // Enable TypeScript error checking during builds
        ignoreBuildErrors: false,
    },
    eslint: {
        // Enable ESLint error checking during builds
        ignoreDuringBuilds: false,
    },
}

module.exports = nextConfig
