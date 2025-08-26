import { CircularProgress, Box } from '@mui/material';

export default function Loading() {
	return (
		<Box className="flex h-64 items-center justify-center">
			<CircularProgress size={40} />
		</Box>
	);
}
