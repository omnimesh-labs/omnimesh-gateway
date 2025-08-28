'use server';

import { revalidatePath } from 'next/cache';
import * as serverApi from '@/server/api';
import type { CreateEndpointRequest, UpdateEndpointRequest } from '@/lib/types';

export async function createEndpointAction(data: CreateEndpointRequest) {
	try {
		const endpoint = await serverApi.createEndpoint(data);
		revalidatePath('/endpoints');
		return { success: true, data: endpoint };
	} catch (error) {
		console.error('Failed to create endpoint:', error);
		return { success: false, error: error instanceof Error ? error.message : 'Failed to create endpoint' };
	}
}

export async function updateEndpointAction(id: string, data: Partial<UpdateEndpointRequest>) {
	try {
		const endpoint = await serverApi.updateEndpoint(id, data);
		revalidatePath('/endpoints');
		return { success: true, data: endpoint };
	} catch (error) {
		console.error('Failed to update endpoint:', error);
		return { success: false, error: error instanceof Error ? error.message : 'Failed to update endpoint' };
	}
}

export async function deleteEndpointAction(id: string) {
	try {
		await serverApi.deleteEndpoint(id);
		revalidatePath('/endpoints');
		return { success: true };
	} catch (error) {
		console.error('Failed to delete endpoint:', error);
		return { success: false, error: error instanceof Error ? error.message : 'Failed to delete endpoint' };
	}
}
