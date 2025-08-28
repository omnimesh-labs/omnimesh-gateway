import ToolsView from './components/views/ToolsView';

// Force dynamic rendering to prevent build-time evaluation
export const dynamic = 'force-dynamic';

export default function ToolsPage() {
	return <ToolsView />;
}
