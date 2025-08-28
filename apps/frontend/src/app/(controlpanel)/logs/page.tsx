import LogsView from './LogsView';

// Force dynamic rendering to prevent build-time evaluation
export const dynamic = 'force-dynamic';

export default function LogsPage() {
	return <LogsView />;
}
