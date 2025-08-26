import { useQuery, UseQueryOptions, QueryKey } from '@tanstack/react-query';
import { getCachedData, setCachedData } from '@/lib/performance';

interface OptimizedQueryOptions<TData> extends Omit<UseQueryOptions<TData>, 'queryKey' | 'queryFn'> {
  cacheKey?: string;
  enableMemoryCache?: boolean;
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
      // Check memory cache first if enabled
      if (enableMemoryCache) {
        const cached = getCachedData<TData>(memCacheKey);
        if (cached) {
          console.log('[Cache] Using memory cached data for:', memCacheKey);
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
    ...queryOptions,
  });
}