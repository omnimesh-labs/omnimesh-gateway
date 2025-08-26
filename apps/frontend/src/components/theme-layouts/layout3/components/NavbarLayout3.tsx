import Scrollbars from '@fuse/core/Scrollbars';
import { styled } from '@mui/material/styles';
import clsx from 'clsx';
import { memo } from 'react';
import Navigation from '../../components/navigation/Navigation';
import NavigationSearch from '../../components/navigation/NavigationSearch';
import useThemeMediaQuery from '@fuse/hooks/useThemeMediaQuery';

const Root = styled('div')(({ theme }) => ({
	backgroundColor: theme.vars.palette.background.default,
	color: theme.vars.palette.text.primary
}));

type NavbarLayout3Props = {
	className?: string;
};

/**
 * The navbar layout 3.
 */
function NavbarLayout3(props: NavbarLayout3Props) {
	const { className = '' } = props;

	const isMobile = useThemeMediaQuery((theme) => theme.breakpoints.down('lg'));

	return (
		<Root className={clsx('h-16 max-h-16 min-h-16 w-full', className)}>
			<div className="container flex h-full w-full flex-auto items-center px-4 lg:px-8">
				<Scrollbars className="flex h-full w-full flex-auto items-center">
					<Navigation
						className="w-full"
						layout="horizontal"
						dense
					/>
				</Scrollbars>
				<div className="flex">{!isMobile && <NavigationSearch variant="basic" />}</div>
			</div>
		</Root>
	);
}

export default memo(NavbarLayout3);
