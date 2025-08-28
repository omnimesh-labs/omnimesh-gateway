import ResourcesView from './components/views/ResourcesView';

// Force dynamic rendering to prevent build-time evaluation
export const dynamic = 'force-dynamic';

export default function ResourcesPage() {
	return <ResourcesView />;
}
