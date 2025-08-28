import DashboardView from './(controlpanel)/dashboard/DashboardView';
import AuthGuard from '@auth/AuthGuard';
import MainLayout from 'src/components/MainLayout';

function MainPage() {
	// Show the dashboard directly at the root path
	return (
		<AuthGuard>
			<MainLayout>
				<DashboardView />
			</MainLayout>
		</AuthGuard>
	);
}

export default MainPage;
