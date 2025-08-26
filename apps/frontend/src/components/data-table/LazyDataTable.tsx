'use client';

import { lazy, Suspense, useState, useEffect, useRef } from 'react';
import { MaterialReactTableProps } from 'material-react-table';
import { CircularProgress, Box } from '@mui/material';

// Preload the DataTable component immediately
const DataTable = lazy(() => {
	const promise = import('./DataTable');
	// Start loading immediately
	promise.then(() => {
		// Component loaded
	});
	return promise;
});

// Trigger preload on module load
if (typeof window !== 'undefined') {
	// Preload the component after a short delay to not block initial render
	setTimeout(() => {
		import('./DataTable');
	}, 100);
}

// Lightweight loading fallback
const DataTableSkeleton = () => (
	<Box className="flex h-64 items-center justify-center">
		<CircularProgress size={40} />
	</Box>
);

function LazyDataTable<TData>(props: MaterialReactTableProps<TData>) {
	const [mounted, setMounted] = useState(false);
	const hasInteracted = useRef(false);

	useEffect(() => {
		// Quick mount for better perceived performance
		const timer = requestAnimationFrame(() => {
			setMounted(true);
		});

		// Track user interaction to prioritize loading
		const handleInteraction = () => {
			if (!hasInteracted.current) {
				hasInteracted.current = true;
				// Force immediate load on interaction
				setMounted(true);
			}
		};

		window.addEventListener('mousemove', handleInteraction, { once: true });
		window.addEventListener('touchstart', handleInteraction, { once: true });

		return () => {
			cancelAnimationFrame(timer);
			window.removeEventListener('mousemove', handleInteraction);
			window.removeEventListener('touchstart', handleInteraction);
		};
	}, []);

	if (!mounted) {
		return <DataTableSkeleton />;
	}

	return (
		<Suspense fallback={<DataTableSkeleton />}>
			<DataTable {...props} />
		</Suspense>
	);
}

export default LazyDataTable;
