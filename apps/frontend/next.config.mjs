const isTurbopack = process.env.TURBOPACK === '1';

// Conditionally add webpack configuration only when NOT using turbopack
const nextConfig = {
	reactStrictMode: false,
	eslint: {
		// Only enable ESLint in development
		ignoreDuringBuilds: process.env.NODE_ENV === 'production'
	},
	typescript: {
		// Dangerously allow production builds to successfully complete even if
		// your project has type errors.
		// ignoreBuildErrors: true
	},
	// Enable experimental features for better performance
	experimental: {
		// Optimize for faster refreshes
		optimizePackageImports: ['@mui/material', '@mui/icons-material', 'lodash', 'date-fns'],
		// Preload all pages on start in development
		preloadEntriesOnStart: true,
	},
	// Compiler optimizations
	compiler: {
		// Remove console logs in production
		removeConsole: process.env.NODE_ENV === 'production',
	},
	modularizeImports: {
		'@mui/material': {
			transform: '@mui/material/{{member}}',
		},
		'@mui/icons-material': {
			transform: '@mui/icons-material/{{member}}',
		},
		lodash: {
			transform: 'lodash/{{member}}',
		},
	},
	// Remove turbopack rules for now - let Next.js handle SVGs by default
	// turbopack: {
	// 	rules: {}
	// },
	...(!isTurbopack && {
		webpack: (config, { dev, isServer }) => {
			if (config.module && config.module.rules) {
				config.module.rules.push({
					test: /\.(json|js|ts|tsx|jsx)$/,
					resourceQuery: /raw/,
					use: 'raw-loader'
				});
			}

			// Optimize webpack for development
			if (dev && !isServer) {
				// Use fastest source maps in development
				config.devtool = 'eval';
				
				// Optimize module resolution
				config.resolve = {
					...config.resolve,
					// Cache module resolutions
					unsafeCache: true,
					// Skip symlink resolution for faster builds
					symlinks: false,
				};

				// Optimize for faster rebuilds
				config.optimization = {
					...config.optimization,
					removeAvailableModules: false,
					removeEmptyChunks: false,
					splitChunks: false,
					// Skip minimization in dev
					minimize: false,
					// Use deterministic module ids for caching
					moduleIds: 'deterministic',
				};

				// Use faster hashing algorithm in dev
				config.output.hashFunction = 'xxhash64';
				
				// Cache webpack modules
				config.cache = {
					type: 'filesystem',
					allowCollectingMemory: true,
					buildDependencies: {
						config: [require.resolve('./next.config.mjs')],
					},
				};
				
				// Ignore large modules that slow down builds
				config.watchOptions = {
					ignored: ['**/node_modules', '**/.next'],
				};
			}

			return config;
		}
	})
};

export default nextConfig;
