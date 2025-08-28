'use client';

import Error500PageView from '../../components/views/Error500PageView';

// Force dynamic rendering for this error page
export const dynamic = 'force-dynamic';

function Page() {
	return <Error500PageView />;
}

export default Page;
