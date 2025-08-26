'use client';

import React, { useCallback, useEffect, useState } from 'react';
import Utils from '@fuse/utils';
import {
	getSessionRedirectUrl,
	resetSessionRedirectUrl,
	setSessionRedirectUrl
} from '@fuse/core/Authorization/sessionRedirectUrl';
import { RouteObjectType } from '@fuse/core/Layout/Layout';
import usePathname from '@fuse/hooks/usePathname';
import Loading from '@fuse/core/Loading';
import useNavigate from '@fuse/hooks/useNavigate';
import useUser from './useUser';

type AuthGuardProps = {
	auth: RouteObjectType['auth'];
	children: React.ReactNode;
	loginRedirectUrl?: string;
};

function AuthGuardRedirect({ auth, children, loginRedirectUrl = '/' }: AuthGuardProps) {
	const { data: user, isGuest } = useUser();
	const userRole = user?.role;
	const navigate = useNavigate();

	const [accessGranted, setAccessGranted] = useState<boolean>(false);
	const pathname = usePathname();

	// Function to handle redirection
	const handleRedirection = useCallback(() => {
		const redirectUrl = getSessionRedirectUrl() || loginRedirectUrl;

		if (isGuest) {
			navigate('/sign-in');
		} else {
			navigate(redirectUrl);
			resetSessionRedirectUrl();
		}
	}, [isGuest, loginRedirectUrl, navigate]);

	// Check user's permissions and set access granted state
	useEffect(() => {
		const isOnlyGuestAllowed = Array.isArray(auth) && auth.length === 0;
		const userHasPermission = Utils.hasPermission(auth, userRole);
		const ignoredPaths = ['/', '/callback', '/sign-in', '/sign-out', '/logout', '/404'];

		if (!auth || (auth && userHasPermission) || (isOnlyGuestAllowed && isGuest)) {
			setAccessGranted(true);
			return;
		}

		if (!userHasPermission) {
			if (isGuest && !ignoredPaths.includes(pathname)) {
				setSessionRedirectUrl(pathname);
			} else if (!isGuest && !ignoredPaths.includes(pathname)) {
				/**
				 * If user is member but don't have permission to view the route
				 * redirected to main route '/'
				 */
				if (isOnlyGuestAllowed) {
					setSessionRedirectUrl('/');
				} else {
					setSessionRedirectUrl('/401');
				}
			}
		}

		handleRedirection();
	}, [auth, userRole, isGuest, pathname, handleRedirection]);

	// Return children if access is granted, otherwise null
	return accessGranted ? children : <Loading />;
}

// the landing page "/" redirected to /example but the example npt

export default AuthGuardRedirect;
