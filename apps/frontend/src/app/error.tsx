'use client';

import { useEffect } from 'react';
import { Typography, Button, Container } from '@mui/material';
import Link from '@fuse/core/Link';
import SvgIcon from '@fuse/core/SvgIcon';

type ErrorProps = {
	error: Error & { digest?: string };
	reset: () => void;
};

export default function Error({ error, reset }: ErrorProps) {
	useEffect(() => {
		console.error(error);
	}, [error]);

	return (
		<Container maxWidth="sm">
			<div className="flex min-h-screen flex-col items-center justify-center text-center">
				<SvgIcon
					className="mb-4"
					color="error"
					size={64}
				>
					lucide:info
				</SvgIcon>
				<Typography
					className="mb-4 text-xl lg:text-3xl"
					color="error.main"
				>
					Oops! Something went wrong
				</Typography>
				<Typography
					className="mb-8"
					color="text.secondary"
				>
					{error.message || 'An unexpected error occurred'}
				</Typography>
				<div className="flex gap-2">
					<Button
						component={Link}
						to="/"
						variant="contained"
						color="primary"
						size="small"
					>
						Go to homepage
					</Button>
					<Button
						onClick={() => reset()}
						variant="outlined"
						color="secondary"
						size="small"
					>
						Try again
					</Button>
				</div>
			</div>
		</Container>
	);
}
