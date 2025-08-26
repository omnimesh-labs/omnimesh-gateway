import MainLayout from 'src/components/MainLayout';
import AuthGuard from '@auth/AuthGuard';

function Layout({ children }) {
	return (
		<AuthGuard>
			<MainLayout>{children}</MainLayout>
		</AuthGuard>
	);
}

export default Layout;
