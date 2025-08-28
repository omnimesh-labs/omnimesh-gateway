/** @type {import('next').NextConfig} */
const nextConfig = {
	reactStrictMode: false,
	eslint: {
		ignoreDuringBuilds: true
	},
	typescript: {
		ignoreBuildErrors: true
	},
	generateBuildId: async () => {
		return 'build-' + Date.now()
	},
	experimental: {
		forceSwcTransforms: true
	}
};

export default nextConfig;
