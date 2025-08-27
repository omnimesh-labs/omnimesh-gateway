/** @type {import('next').NextConfig} */
const nextConfig = {
	reactStrictMode: false,
	eslint: {
		ignoreDuringBuilds: true
	},
	typescript: {
		ignoreBuildErrors: true
	},
	experimental: {
		optimizePackageImports: [
			'@mui/material',
			'@mui/icons-material'
		]
	},
	compiler: {
		removeConsole: process.env.NODE_ENV === 'production' ? {
			exclude: ['error', 'warn']
		} : false
	}
};

export default nextConfig;