'use client';

import Error404PageView from '../../components/views/Error404PageView';

// Force dynamic rendering for this error page
export const dynamic = 'force-dynamic';

function Page() {
	return <Error404PageView />;
}

export default Page;
