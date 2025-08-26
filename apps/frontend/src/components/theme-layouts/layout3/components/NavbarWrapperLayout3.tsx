import { styled } from '@mui/material/styles';
import SwipeableDrawer from '@mui/material/SwipeableDrawer';
import { useEffect } from 'react';
import useThemeMediaQuery from '@fuse/hooks/useThemeMediaQuery';
import NavbarToggleFab from 'src/components/theme-layouts/components/navbar/NavbarToggleFab';
import usePathname from '@fuse/hooks/usePathname';
import useLayoutSettings from '@fuse/core/Layout/useLayoutSettings';
import NavbarLayout3 from './NavbarLayout3';
import NavbarMobileLayout3 from './NavbarMobileLayout3';
import { useNavbarContext } from '../../components/navbar/contexts/NavbarContext/useNavbarContext';
import NavbarTheme from '@/contexts/NavbarTheme';

const StyledSwipeableDrawer = styled(SwipeableDrawer)(({ theme }) => ({
	'& > .MuiDrawer-paper': {
		height: '100%',
		flexDirection: 'column',
		flex: '1 1 auto',
		width: 280,
		minWidth: 280,
		transition: theme.transitions.create(['width', 'min-width'], {
			easing: theme.transitions.easing.sharp,
			duration: theme.transitions.duration.shorter
		})
	}
}));

type NavbarWrapperLayout3Props = {
	className?: string;
};

/**
 * The navbar wrapper layout 3.
 */
function NavbarWrapperLayout3(props: NavbarWrapperLayout3Props) {
	const { className = '' } = props;

	const { config } = useLayoutSettings();
	const { mobileOpen: isNavbarMobileOpen, closeMobileNavbar } = useNavbarContext();
	const pathname = usePathname();
	const isMobile = useThemeMediaQuery((theme) => theme.breakpoints.down('lg'));

	useEffect(() => {
		if (isMobile) {
			closeMobileNavbar();
		}
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [pathname, isMobile]);

	return (
		<>
			<NavbarTheme>
				{!isMobile && <NavbarLayout3 className={className} />}

				{isMobile && (
					<StyledSwipeableDrawer
						anchor="left"
						variant="temporary"
						open={isNavbarMobileOpen}
						onClose={() => closeMobileNavbar()}
						onOpen={() => {}}
						disableSwipeToOpen
						ModalProps={{
							keepMounted: true // Better open performance on mobile.
						}}
					>
						<NavbarMobileLayout3 />
					</StyledSwipeableDrawer>
				)}
			</NavbarTheme>

			{config.navbar.display && !config.toolbar.display && isMobile && <NavbarToggleFab />}
		</>
	);
}

export default NavbarWrapperLayout3;
