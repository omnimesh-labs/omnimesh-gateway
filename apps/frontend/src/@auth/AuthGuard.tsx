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
            if (pathname !== '/sign-in' && pathname !== '/sign-up') {
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
        return null;
    }

    return <>{children}</>;
}