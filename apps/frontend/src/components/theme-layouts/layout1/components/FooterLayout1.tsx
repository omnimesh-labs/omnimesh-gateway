import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import { memo } from 'react';
import clsx from 'clsx';
import FooterTheme from '@/contexts/FooterTheme';

type FooterLayout1Props = { className?: string };

/**
 * The footer layout 1.
 */
function FooterLayout1(props: FooterLayout1Props) {
	const { className } = props;

	return (
		<FooterTheme>
			<AppBar
				id="fuse-footer"
				className={clsx('relative z-20 border-t', className)}
				color="default"
				sx={(theme) => ({
					backgroundColor: theme.vars.palette.background.default,
					color: theme.vars.palette.text.primary
				})}
				elevation={0}
			>
				<Toolbar className="flex min-h-12 items-center overflow-x-auto px-2 py-0 sm:px-3 md:min-h-16">
					Footer
				</Toolbar>
			</AppBar>
		</FooterTheme>
	);
}

export default memo(FooterLayout1);
