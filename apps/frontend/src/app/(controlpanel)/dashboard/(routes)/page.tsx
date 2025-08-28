import DashboardView from '../DashboardView';

// Force dynamic rendering to prevent build-time evaluation
export const dynamic = 'force-dynamic';

export default function DashboardPage() {
	return <DashboardView />;
}
