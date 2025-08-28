import PromptsView from './components/views/PromptsView';

// Force dynamic rendering to prevent build-time evaluation
export const dynamic = 'force-dynamic';

export default function PromptsPage() {
	return <PromptsView />;
}
