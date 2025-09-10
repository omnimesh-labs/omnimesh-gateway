import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { promptsApi } from '@/lib/client-api';
import type { Prompt, CreatePromptRequest, UpdatePromptRequest } from '@/lib/types';

export function usePrompts(params?: {
	active?: boolean;
	category?: string;
	search?: string;
	popular?: boolean;
	limit?: number;
	offset?: number;
}) {
	return useQuery({
		queryKey: ['prompts', params],
		queryFn: () => promptsApi.listPrompts(params),
	});
}

export function usePrompt(id: string | null) {
	return useQuery({
		queryKey: ['prompt', id],
		queryFn: () => promptsApi.getPrompt(id!),
		enabled: !!id,
	});
}

export function useCreatePrompt() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (data: CreatePromptRequest) => promptsApi.createPrompt(data),
		onSuccess: () => {
			// Invalidate and refetch prompts list
			queryClient.invalidateQueries({ queryKey: ['prompts'] });
		},
	});
}

export function useUpdatePrompt() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: ({ id, data }: { id: string; data: UpdatePromptRequest }) =>
			promptsApi.updatePrompt(id, data),
		onSuccess: (updatedPrompt: Prompt) => {
			// Invalidate and refetch prompts list
			queryClient.invalidateQueries({ queryKey: ['prompts'] });
			// Update the specific prompt in cache
			queryClient.setQueryData(['prompt', updatedPrompt.id], updatedPrompt);
		},
	});
}

export function useDeletePrompt() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (id: string) => promptsApi.deletePrompt(id),
		onSuccess: () => {
			// Invalidate and refetch prompts list
			queryClient.invalidateQueries({ queryKey: ['prompts'] });
		},
	});
}

export function useUsePrompt() {
	const queryClient = useQueryClient();

	return useMutation({
		mutationFn: (id: string) => promptsApi.usePrompt(id),
		onSuccess: (usedPrompt: Prompt) => {
			// Update prompts list to reflect usage count change
			queryClient.invalidateQueries({ queryKey: ['prompts'] });
			// Update the specific prompt in cache
			queryClient.setQueryData(['prompt', usedPrompt.id], usedPrompt);
		},
	});
}
