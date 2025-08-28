import AuthMethodsView from './AuthMethodsView';

// Force dynamic rendering to prevent build-time evaluation
export const dynamic = 'force-dynamic';
export default function SecurityAuthPage() {
	return <AuthMethodsView />;
}
