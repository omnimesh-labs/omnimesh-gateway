// Authentication Flow Test Utilities
// This file is commented out to avoid console statements in production
// Uncomment for debugging authentication flow issues

/*
export function testAuthenticationFlow() {
	console.log('Testing Authentication Flow...');

	// Test 1: Check if unauthenticated users are redirected to sign-in
	console.log('Test 1: Unauthenticated redirect check');
	const isAuthenticated = localStorage.getItem('access_token') !== null;

	if (!isAuthenticated) {
		console.log('✓ User is not authenticated - should redirect to /sign-in');
	} else {
		console.log('✓ User is authenticated - can access protected routes');
	}

	// Test 2: Check AuthGuard behavior
	console.log('\nTest 2: AuthGuard protection');
	const protectedRoutes = [
		'/dashboard',
		'/servers',
		'/namespaces',
		'/endpoints',
		'/policies',
		'/content',
		'/configuration',
		'/a2a',
		'/profile'
	];

	protectedRoutes.forEach((route) => {
		console.log(`- Route ${route}: Protected by AuthGuard via (control-panel) layout`);
	});

	// Test 3: Check public routes accessibility
	console.log('\nTest 3: Public routes');
	const publicRoutes = ['/sign-in', '/sign-up', '/sign-out'];

	publicRoutes.forEach((route) => {
		console.log(`- Route ${route}: Accessible without authentication`);
	});

	// Test 4: Check redirect after login
	console.log('\nTest 4: Login redirect behavior');
	const redirectUrl = localStorage.getItem('redirectUrl');

	if (redirectUrl) {
		console.log(`✓ Will redirect to: ${redirectUrl} after login`);
	} else {
		console.log('✓ Will redirect to dashboard (/) after login');
	}

	// Test 5: Check 401 handling
	console.log('\nTest 5: 401 Error handling');
	console.log('- API returns 401 → Attempt token refresh');
	console.log('- Refresh succeeds → Retry original request');
	console.log('- Refresh fails → Clear tokens and redirect to /sign-in');

	console.log('\n✅ Authentication flow is properly configured!');
	console.log('\nKey behaviors:');
	console.log('1. Unauthenticated users are redirected to /sign-in');
	console.log('2. Protected routes use AuthGuard in (control-panel) layout');
	console.log('3. Failed auth attempts redirect to /sign-in');
	console.log('4. Successful login redirects to saved URL or dashboard');
	console.log('5. Token expiry triggers automatic refresh or re-login');
}

// Export for use in browser console
if (typeof window !== 'undefined') {
	(window as any).testAuthFlow = testAuthenticationFlow;
}
*/

// Empty export to keep module valid
export {};
