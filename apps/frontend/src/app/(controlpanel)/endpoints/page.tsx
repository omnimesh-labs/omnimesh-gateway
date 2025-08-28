import * as serverApi from '@/server/api';
import EndpointsView from './EndpointsView';

// Force dynamic rendering to prevent build-time evaluation
export const dynamic = 'force-dynamic';

export default async function EndpointsPage() {
	// Fetch data on the server
	const [initialEndpoints, initialNamespaces] = await Promise.all([
		serverApi.listEndpoints(),
		serverApi.listNamespaces()
	]);

	return (
		<EndpointsView
			initialEndpoints={initialEndpoints}
			initialNamespaces={initialNamespaces}
		/>
	);
}
