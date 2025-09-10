import { useQuery } from '@tanstack/react-query';

/**
 * Placeholder hook for fetching resources.
 * TODO: Replace with actual API call.
 */
export function useResources(params?: any) {
	return useQuery({
		queryKey: ['resources', params],
		queryFn: async () => {
			return Promise.resolve([]);
		}
	});
}

export function useResource(id: string | null) {
	return useQuery({
		queryKey: ['resource', id],
		queryFn: async () => {
			console.warn('useResource is using placeholder data. API endpoint not implemented.');
			return Promise.resolve(null);
		},
		enabled: !!id
	});
}
