import { styled } from '@mui/material/styles';
import SwipeableDrawer from '@mui/material/SwipeableDrawer';
import { useEffect } from 'react';
import NavbarToggleFabLayout2 from 'src/components/theme-layouts/layout2/components/NavbarToggleFabLayout2';
import useThemeMediaQuery from '@fuse/hooks/useThemeMediaQuery';
import usePathname from '@fuse/hooks/usePathname';
import useLayoutSettings from '@fuse/core/Layout/useLayoutSettings';
import NavbarLayout2 from './NavbarLayout2';
import NavbarMobileLayout2 from './NavbarMobileLayout2';
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

type NavbarWrapperLayout2Props = {
	className?: string;
};

/**
 * The navbar wrapper layout 2.
 */
function NavbarWrapperLayout2(props: NavbarWrapperLayout2Props) {
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
				{!isMobile && <NavbarLayout2 />}

				{isMobile && (
					<StyledSwipeableDrawer
						anchor="left"
						variant="temporary"
						open={isNavbarMobileOpen}
						onClose={() => closeMobileNavbar()}
						onOpen={() => {}}
						disableSwipeToOpen
						className={className}
						ModalProps={{
							keepMounted: true // Better open performance on mobile.
						}}
					>
						<NavbarMobileLayout2 />
					</StyledSwipeableDrawer>
				)}
			</NavbarTheme>
			{config.navbar.display && !config.toolbar.display && isMobile && <NavbarToggleFabLayout2 />}
		</>
	);
}

export default NavbarWrapperLayout2;
