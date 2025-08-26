'use client';

import AuthGuardRedirect from '@auth/AuthGuardRedirect';
import SignOutPageView from '../../components/views/SignOutPageView';

function Page() {
	return (
		<AuthGuardRedirect auth={null}>
			<SignOutPageView />
		</AuthGuardRedirect>
	);
}

export default Page;
