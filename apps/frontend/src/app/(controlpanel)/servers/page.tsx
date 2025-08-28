import ServersView from './ServersView';

// Force dynamic rendering to prevent build-time evaluation
export const dynamic = 'force-dynamic';

export default function ServersPage() {
	return <ServersView />;
}
