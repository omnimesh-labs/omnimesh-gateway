'use client';

import Link from 'next/link';

export default function GlobalError({
	error,
	reset,
}: {
	error: Error & { digest?: string };
	reset: () => void;
}) {
	return (
		<html>
			<body>
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
					<h2 style={{ color: '#dc2626', marginBottom: '1rem' }}>Something went wrong!</h2>
					<p style={{ color: '#666', marginBottom: '2rem' }}>{error.message || 'An unexpected error occurred'}</p>
					<div style={{ display: 'flex', gap: '1rem' }}>
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
						<button
							onClick={() => reset()}
							style={{
								padding: '0.5rem 1rem',
								backgroundColor: 'transparent',
								color: '#3b82f6',
								border: '1px solid #3b82f6',
								borderRadius: '0.25rem',
								cursor: 'pointer'
							}}
						>
							Try again
						</button>
					</div>
				</div>
			</body>
		</html>
	);
}
