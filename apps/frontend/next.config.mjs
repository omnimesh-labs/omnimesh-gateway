/** @type {import('next').NextConfig} */
const nextConfig = {
	reactStrictMode: false,
	output: 'standalone',
	eslint: {
		ignoreDuringBuilds: true
	},
	typescript: {
		ignoreBuildErrors: true
	},
	generateBuildId: async () => {
		return 'build-' + Date.now()
	},
	images: {
		unoptimized: true
	}
};

export default nextConfig;
