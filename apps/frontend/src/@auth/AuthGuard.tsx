'use client';

import { ReactNode, useEffect } from 'react';
import { usePathname, useRouter } from 'next/navigation';
import Loading from '@fuse/core/Loading';
import { useAuth } from './AuthContext';

interface AuthGuardProps {
	children: ReactNode;
}

export default function AuthGuard({ children }: AuthGuardProps) {
	const { isAuthenticated, isLoading } = useAuth();
	const pathname = usePathname();
	const router = useRouter();

	useEffect(() => {
		if (!isLoading && !isAuthenticated) {
			// Store the current path to redirect back after login
			if (
				pathname &&
				pathname !== '/sign-in' &&
				pathname !== '/sign-up' &&
				pathname !== '/401' &&
				pathname !== '/404'
			) {
				localStorage.setItem('redirectUrl', pathname);
			}

			// Redirect to sign-in page
			router.replace('/sign-in');
		}
	}, [isAuthenticated, isLoading, pathname, router]);

	if (isLoading) {
		return <Loading />;
	}

	if (!isAuthenticated) {
		// Show loading while redirecting to prevent flash of protected content
		return <Loading />;
	}

	return <>{children}</>;
}
