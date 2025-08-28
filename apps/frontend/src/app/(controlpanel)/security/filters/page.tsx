import ContentFiltersView from './ContentFiltersView';

// Force dynamic rendering to prevent build-time evaluation
export const dynamic = 'force-dynamic';
export default function SecurityFiltersPage() {
	return <ContentFiltersView />;
}
