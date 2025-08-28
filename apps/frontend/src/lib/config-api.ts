const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';

class ConfigAPI {
	public async exportConfiguration(options: Record<string, unknown>): Promise<Blob> {
		const response = await fetch(`${API_BASE_URL}/admin/export`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(options)
		});

		if (!response.ok) {
			throw new Error(`Export failed: ${response.status} ${response.statusText}`);
		}

		return response.blob();
	}

	public async importConfiguration(data: FormData): Promise<void> {
		const response = await fetch(`${API_BASE_URL}/admin/import`, {
			method: 'POST',
			body: data
		});

		if (!response.ok) {
			throw new Error(`Import failed: ${response.status} ${response.statusText}`);
		}
	}
}

export const configApi = new ConfigAPI();
