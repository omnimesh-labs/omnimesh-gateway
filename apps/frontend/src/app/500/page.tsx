'use client';

import Link from 'next/link';

export default function Page500() {
	return (
		<div
			style={{
				display: 'flex',
				flexDirection: 'column',
				alignItems: 'center',
				justifyContent: 'center',
				minHeight: '100vh',
				fontFamily: 'system-ui, -apple-system, sans-serif',
				textAlign: 'center',
				padding: '2rem'
			}}
		>
			<h1 style={{ fontSize: '2rem', marginBottom: '0.5rem' }}>500 - Server error</h1>
			<p style={{ color: '#666', marginBottom: '1.5rem' }}>Please try again or return to the homepage.</p>
			<Link
				href="/"
				style={{
					padding: '0.5rem 1rem',
					backgroundColor: '#3b82f6',
					color: 'white',
					textDecoration: 'none',
					borderRadius: '0.25rem',
					border: 'none',
					cursor: 'pointer'
				}}
			>
				Go to homepage
			</Link>
		</div>
	);
}
