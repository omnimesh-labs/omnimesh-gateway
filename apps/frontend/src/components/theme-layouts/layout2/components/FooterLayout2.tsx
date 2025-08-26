import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import clsx from 'clsx';
import { memo } from 'react';
import FooterTheme from '@/contexts/FooterTheme';

type FooterLayout2Props = {
	className?: string;
};

/**
 * The footer layout 2.
 */
function FooterLayout2(props: FooterLayout2Props) {
	const { className = '' } = props;

	return (
		<FooterTheme>
			<AppBar
				id="fuse-footer"
				className={clsx('relative z-20 border-t', className)}
				color="default"
				sx={{ backgroundColor: (theme) => theme.vars.palette.background.paper }}
			>
				<Toolbar className="container flex min-h-12 items-center overflow-x-auto px-2 py-0 sm:px-8 md:min-h-16">
					Footer
				</Toolbar>
			</AppBar>
		</FooterTheme>
	);
}

export default memo(FooterLayout2);
