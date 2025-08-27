'use client';

import { useEffect } from 'react';

type ErrorProps = {
	error: Error & { digest?: string };
	reset: () => void;
};

export default function Error({ error, reset }: ErrorProps) {
	useEffect(() => {
		console.error(error);
	}, [error]);

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
			<h1 style={{ color: '#dc2626', marginBottom: '1rem', fontSize: '2rem' }}>
				Oops! Something went wrong
			</h1>
			<p style={{ color: '#666', marginBottom: '2rem', fontSize: '1rem' }}>
				{error.message || 'An unexpected error occurred'}
			</p>
			<div style={{ display: 'flex', gap: '1rem' }}>
				<a
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
				</a>
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
	);
}
