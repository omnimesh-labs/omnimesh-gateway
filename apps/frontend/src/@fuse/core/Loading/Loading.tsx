import { useTimeout } from '@fuse/hooks';
import { useState } from 'react';
import clsx from 'clsx';
import Box from '@mui/material/Box';

export type LoadingProps = {
	delay?: number;
	className?: string;
};

/**
 * Loading displays a loading state with an optional delay
 */
function Loading(props: LoadingProps) {
	const { delay = 0, className } = props;
	const [showLoading, setShowLoading] = useState(!delay);

	useTimeout(() => {
		setShowLoading(true);
	}, delay);

	return (
		<div
			className={clsx(
				className,
				'flex h-full min-h-full w-full flex-1 flex-col items-center justify-center self-center p-6',
				!showLoading ? 'hidden' : ''
			)}
		>
			<Box
				id="spinner"
				sx={{
					'& > div': {
						backgroundColor: 'palette.secondary.main'
					}
				}}
			>
				<div className="bounce1" />
				<div className="bounce2" />
				<div className="bounce3" />
			</Box>
		</div>
	);
}

export default Loading;
