const isTurbopack = process.env.TURBOPACK === '1';

// Bundle analyzer for production builds
const withBundleAnalyzer =
	process.env.ANALYZE === 'true'
		? (await import('@next/bundle-analyzer').then((m) => m.default))({
				enabled: true,
				openAnalyzer: false
			})
		: (config) => config;

// Conditionally add webpack configuration only when NOT using turbopack
const baseConfig = {
	reactStrictMode: false,
	skipMiddlewareUrlNormalize: true,
	output: 'standalone',
	eslint: {
		// Only enable ESLint in development
		ignoreDuringBuilds: process.env.NODE_ENV === 'production'
	},
	typescript: {
		// Dangerously allow production builds to successfully complete even if
		// your project has type errors.
		ignoreBuildErrors: true
	},
	// Enable experimental features for better performance
	experimental: {
		// Optimize for faster refreshes - extensive package imports optimization
		optimizePackageImports: [
			'@mui/material',
			'@mui/icons-material',
			'@mui/lab',
			'@mui/x-date-pickers',
			'@mui/system',
			'lodash',
			'date-fns',
			'@tanstack/react-query',
			'notistack',
			'react-hook-form',
			'material-react-table',
			'@fuse/core',
			'@fuse/hooks'
		],
		// Enable optimizations for better performance in dev
		optimizeServerReact: true,
		forceSwcTransforms: true
	},
	// Force dynamic rendering for problematic pages
	async generateBuildId() {
		return 'janex-build';
	},
	// Compiler optimizations
	compiler: {
		// Remove console logs in production
		removeConsole:
			process.env.NODE_ENV === 'production'
				? {
						exclude: ['error', 'warn']
					}
				: false,
		// Enable React compiler optimizations
		reactRemoveProperties:
			process.env.NODE_ENV === 'production'
				? {
						properties: ['^data-testid$']
					}
				: false
	},
	modularizeImports: {
		'@mui/material': {
			transform: '@mui/material/{{member}}'
		},
		'@mui/icons-material': {
			transform: '@mui/icons-material/{{member}}'
		},
		lodash: {
			transform: 'lodash/{{member}}'
		}
	},

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
				config.devtool = 'eval-cheap-source-map';

				// Optimize module resolution
				config.resolve = {
					...config.resolve,
					// Cache module resolutions
					unsafeCache: true,
					// Skip symlink resolution for faster builds
					symlinks: false,
					// Prefer ESM modules for better tree shaking
					mainFields: ['module', 'main'],
					// Optimize extensions resolution order
					extensions: ['.ts', '.tsx', '.js', '.jsx']
				};

				// Optimize for faster rebuilds
				config.optimization = {
					...config.optimization,
					removeAvailableModules: false,
					removeEmptyChunks: false,
					splitChunks: {
						// Optimize chunking for better performance
						chunks: 'all',
						minSize: 20000,
						maxSize: 244000,
						cacheGroups: {
							default: false,
							vendors: false,
							// Framework chunk (React, Next.js core)
							framework: {
								name: 'framework',
								test: /[\\/]node_modules[\\/](react|react-dom|scheduler|next)[\\/]/,
								chunks: 'all',
								priority: 40,
								reuseExistingChunk: true
							},
							// Create separate chunk for MUI to reduce compilation time
							mui: {
								name: 'mui',
								test: /[\\/]node_modules[\\/]@mui[\\/]/,
								chunks: 'all',
								priority: 30,
								reuseExistingChunk: true
							},
							// Create separate chunk for React Query
							reactQuery: {
								name: 'react-query',
								test: /[\\/]node_modules[\\/]@tanstack[\\/]/,
								chunks: 'all',
								priority: 25,
								reuseExistingChunk: true
							},
							// Create separate chunk for Material React Table
							materialReactTable: {
								name: 'material-react-table',
								test: /[\\/]node_modules[\\/]material-react-table[\\/]/,
								chunks: 'all',
								priority: 35,
								enforce: true,
								reuseExistingChunk: true
							},
							// Fuse components
							fuse: {
								name: 'fuse',
								test: /[\\/]@fuse[\\/]/,
								chunks: 'all',
								priority: 20,
								reuseExistingChunk: true
							},
							// Common chunks for shared modules
							common: {
								name: 'common',
								minChunks: 2,
								priority: 10,
								reuseExistingChunk: true
							}
						}
					},
					// Skip minimization in dev
					minimize: false,
					// Use deterministic module ids for caching
					moduleIds: 'deterministic'
				};

				// Use faster hashing algorithm in dev
				config.output.hashFunction = 'xxhash64';

				// Enhanced caching
				config.cache = {
					type: 'filesystem',
					allowCollectingMemory: true,
					maxMemoryGenerations: 10,
					buildDependencies: {
						config: [require.resolve('./next.config.mjs')]
					},
					// Add cache invalidation for better performance
					version: '2.0',
					// Store cache in memory for faster access
					store: 'pack'
				};

				// Optimized watch options
				config.watchOptions = {
					ignored: ['**/node_modules', '**/.next', '**/logs/**'],
					aggregateTimeout: 300,
					poll: false
				};

				// Add module concatenation for better performance
				config.optimization.concatenateModules = true;

				// Optimize for development
				config.optimization.usedExports = false;
				config.optimization.sideEffects = false;

				// Disable runtime chunk in dev for faster rebuilds
				config.optimization.runtimeChunk = false;
			}

			return config;
		}
	})
};

const nextConfig = withBundleAnalyzer(baseConfig);

export default nextConfig;
