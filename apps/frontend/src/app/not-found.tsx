'use client';

export const dynamic = 'force-dynamic';

import Link from 'next/link';
import { Typography, Button, Container } from '@mui/material';

export default function NotFound() {
	return (
		<Container maxWidth="sm">
			<div className="flex min-h-screen flex-col items-center justify-center text-center">
				<Typography
					className="mb-4 text-4xl font-bold"
					color="text.primary"
				>
					404
				</Typography>
				<Typography
					className="mb-4 text-xl"
					color="text.secondary"
				>
					Page not found
				</Typography>
				<Typography
					className="mb-8"
					color="text.secondary"
				>
					The page you are looking for does not exist.
				</Typography>
				<Button
					component={Link}
					href="/"
					variant="contained"
					color="primary"
					size="small"
				>
					Go to homepage
				</Button>
			</div>
		</Container>
	);
}
