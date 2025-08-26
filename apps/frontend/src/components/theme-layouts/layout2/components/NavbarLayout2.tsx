import Scrollbars from '@fuse/core/Scrollbars';
import { styled } from '@mui/material/styles';
import clsx from 'clsx';
import { memo } from 'react';
import Navigation from 'src/components/theme-layouts/components/navigation/Navigation';
import Logo from '../../components/Logo';

const Root = styled('div')(({ theme }) => ({
	backgroundColor: theme.vars.palette.background.default,
	color: theme.vars.palette.text.primary
}));

type NavbarLayout2Props = {
	className?: string;
};

/**
 * The navbar layout 2.
 */
function NavbarLayout2(props: NavbarLayout2Props) {
	const { className = '' } = props;

	return (
		<Root className={clsx('h-16 max-h-16 min-h-16 w-full', className)}>
			<div className="z-20 container flex h-full w-full flex-auto items-center justify-between gap-2 p-0 lg:px-8">
				<div className="flex flex-auto">
					<Logo className="" />
				</div>

				<Scrollbars className="flex h-full w-full flex-auto items-center">
					<Navigation
						className="w-full justify-end"
						layout="horizontal"
					/>
				</Scrollbars>
			</div>
		</Root>
	);
}

export default memo(NavbarLayout2);
