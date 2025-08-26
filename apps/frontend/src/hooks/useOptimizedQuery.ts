import { useQuery, UseQueryOptions, QueryKey } from '@tanstack/react-query';

interface OptimizedQueryOptions<TData> extends Omit<UseQueryOptions<TData>, 'queryKey' | 'queryFn'> {
	cacheKey?: string;
	enableMemoryCache?: boolean;
}

// Simple in-memory cache
const memoryCache = new Map<string, { data: unknown; timestamp: number }>();
const CACHE_TTL = 5 * 60 * 1000; // 5 minutes

function getCachedData<T>(key: string): T | null {
	const cached = memoryCache.get(key);

	if (cached && Date.now() - cached.timestamp < CACHE_TTL) {
		return cached.data as T;
	}

	if (cached) {
		memoryCache.delete(key);
	}

	return null;
}

function setCachedData<T>(key: string, data: T): void {
	memoryCache.set(key, { data, timestamp: Date.now() });
}

export function useOptimizedQuery<TData = unknown>(
	queryKey: QueryKey,
	queryFn: () => Promise<TData>,
	options?: OptimizedQueryOptions<TData>
) {
	const { cacheKey, enableMemoryCache = true, ...queryOptions } = options || {};

	const memCacheKey = cacheKey || JSON.stringify(queryKey);

	return useQuery<TData>({
		queryKey,
		queryFn: async () => {
			if (enableMemoryCache) {
				const cached = getCachedData<TData>(memCacheKey);

				if (cached) {
					return cached;
				}
			}

			// Fetch fresh data
			const data = await queryFn();

			// Store in memory cache
			if (enableMemoryCache) {
				setCachedData(memCacheKey, data);
			}

			return data;
		},
		// Optimize React Query settings
		staleTime: 1000 * 60 * 5, // Consider data fresh for 5 minutes
		gcTime: 1000 * 60 * 10, // Keep in cache for 10 minutes (renamed from cacheTime)
		refetchOnWindowFocus: false, // Don't refetch on window focus
		refetchOnMount: 'always', // Always check on mount
		retry: 1, // Only retry once on failure
		...queryOptions
	});
}
