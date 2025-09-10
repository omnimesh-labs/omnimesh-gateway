import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toolsApi } from '@/lib/client-api';
import type { Tool, CreateToolRequest, UpdateToolRequest } from '@/lib/types';

export function useTools(params?: {
	active?: boolean;
	category?: string;
	search?: string;
	popular?: boolean;
	include_public?: boolean;
	limit?: number;
	offset?: number;
}) {
	return useQuery({
		queryKey: ['tools', params],
		queryFn: () => toolsApi.listTools(params),
	});
}

export function useTool(id: string | null) {
	return useQuery({
		queryKey: ['tool', id],
		queryFn: () => toolsApi.getTool(id!),
		enabled: !!id,
	});
}

export function useCreateTool() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (data: CreateToolRequest) => toolsApi.createTool(data),
		onSuccess: () => {
			// Invalidate and refetch tools list
			queryClient.invalidateQueries({ queryKey: ['tools'] });
		},
	});
}

export function useUpdateTool() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: ({ id, data }: { id: string; data: UpdateToolRequest }) =>
			toolsApi.updateTool(id, data),
		onSuccess: (updatedTool: Tool) => {
			// Invalidate and refetch tools list
			queryClient.invalidateQueries({ queryKey: ['tools'] });
			// Update the specific tool in cache
			queryClient.setQueryData(['tool', updatedTool.id], updatedTool);
		},
	});
}

export function useDeleteTool() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (id: string) => toolsApi.deleteTool(id),
		onSuccess: () => {
			// Invalidate and refetch tools list
			queryClient.invalidateQueries({ queryKey: ['tools'] });
		},
	});
}

export function useExecuteTool() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (id: string) => toolsApi.executeTool(id),
		onSuccess: (executedTool: Tool) => {
			// Update tools list to reflect usage count change
			queryClient.invalidateQueries({ queryKey: ['tools'] });
			// Update the specific tool in cache
			queryClient.setQueryData(['tool', executedTool.id], executedTool);
		},
	});
}

export function usePublicTools(params?: { limit?: number; offset?: number }) {
	return useQuery({
		queryKey: ['public-tools', params],
		queryFn: () => toolsApi.getPublicTools(params),
	});
}
