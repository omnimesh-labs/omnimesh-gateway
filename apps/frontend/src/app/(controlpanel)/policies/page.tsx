import PoliciesView from './PoliciesView';

// Force dynamic rendering to prevent build-time evaluation
export const dynamic = 'force-dynamic';
export default function PoliciesPage() {
	return <PoliciesView />;
}
