'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@auth/AuthContext';
import Loading from '@fuse/core/Loading';
import SignInPageView from '../../components/views/SignInPageView';

function Page() {
	const { isAuthenticated, isLoading } = useAuth();
	const router = useRouter();

	useEffect(() => {
		if (!isLoading && isAuthenticated) {
			// If user is already authenticated, redirect to dashboard
			// Check for any saved redirect URL first
			const redirectUrl = localStorage.getItem('redirectUrl');
			if (redirectUrl && redirectUrl !== '/sign-in') {
				localStorage.removeItem('redirectUrl');
				router.replace(redirectUrl);
			} else {
				router.replace('/');
			}
		}
	}, [isAuthenticated, isLoading, router]);

	if (isLoading) {
		return <Loading />;
	}

	if (isAuthenticated) {
		// Show loading while redirecting
		return <Loading />;
	}

	return <SignInPageView />;
}

export default Page;
