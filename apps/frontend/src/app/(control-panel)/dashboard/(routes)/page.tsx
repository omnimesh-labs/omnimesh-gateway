import { redirect } from 'next/navigation';

// Redirect /dashboard to / since the dashboard is at the root
function DashboardRedirect() {
	redirect('/');
	return null;
}

export default DashboardRedirect;