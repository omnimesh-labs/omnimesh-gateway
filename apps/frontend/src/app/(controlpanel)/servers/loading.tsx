import CircularProgress from '@mui/material/CircularProgress';
import Box from '@mui/material/Box';

export default function Loading() {
	return (
		<Box className="flex h-64 items-center justify-center">
			<CircularProgress size={40} />
		</Box>
	);
}
