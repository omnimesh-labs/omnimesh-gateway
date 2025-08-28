import EndpointsView from './EndpointsView';

// Force dynamic rendering to prevent build-time evaluation
export const dynamic = 'force-dynamic';

export default function EndpointsPage() {
	return <EndpointsView />;
}
