import NamespaceDetailView from './NamespaceDetailView';

// Configure route as dynamic for client-side rendering
export const dynamic = 'force-dynamic';

// Generate static params for static export
export async function generateStaticParams() {
	// For static export, we can't pre-generate dynamic routes
	// Return empty array and let the route be client-side rendered
	return [];
}

export default NamespaceDetailView;
