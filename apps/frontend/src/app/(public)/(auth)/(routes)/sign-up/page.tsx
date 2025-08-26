'use client';

import authRoles from '@auth/authRoles';
import AuthGuardRedirect from '@auth/AuthGuardRedirect';
import SignUpPageView from '../../components/views/SignUpPageView';

function Page() {
	return (
		<AuthGuardRedirect auth={authRoles.onlyGuest}>
			<SignUpPageView />
		</AuthGuardRedirect>
	);
}

export default Page;
