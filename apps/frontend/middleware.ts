import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

// Define public routes that don't require authentication
const publicRoutes = ['/sign-in', '/sign-up', '/401', '/404'];
const authRoutes = ['/sign-in', '/sign-up'];

export function middleware(request: NextRequest) {
	const { pathname } = request.nextUrl;

	// Allow access to static files and API routes
	if (
		pathname.startsWith('/_next') ||
		pathname.startsWith('/api') ||
		pathname.startsWith('/static') ||
		pathname.includes('.') // static files with extensions
	) {
		return NextResponse.next();
	}

	// Get the access token from cookies or headers
	const accessToken = request.cookies.get('access_token')?.value;
	const authHeader = request.headers.get('authorization');
	const hasToken = accessToken || authHeader?.startsWith('Bearer ');

	// Check if the route is public
	const isPublicRoute = publicRoutes.some((route) => pathname === route);
	const isAuthRoute = authRoutes.some((route) => pathname === route);

	// If user is authenticated and trying to access auth routes, redirect to dashboard
	if (hasToken && isAuthRoute) {
		return NextResponse.redirect(new URL('/', request.url));
	}

	// If user is not authenticated and trying to access protected routes, redirect to sign-in
	if (!hasToken && !isPublicRoute) {
		const signInUrl = new URL('/sign-in', request.url);
		// Save the original URL to redirect back after login
		signInUrl.searchParams.set('redirect', pathname);
		return NextResponse.redirect(signInUrl);
	}

	return NextResponse.next();
}

export const config = {
	matcher: [
		/*
		 * Match all request paths except for the ones starting with:
		 * - _next/static (static files)
		 * - _next/image (image optimization files)
		 * - favicon.ico (favicon file)
		 * - public folder
		 */
		'/((?!_next/static|_next/image|favicon.ico|public|assets|.*\\..*|api/auth).*)'
	]
};
