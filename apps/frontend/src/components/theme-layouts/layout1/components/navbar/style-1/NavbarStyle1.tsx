import { styled, Theme } from '@mui/material/styles';
import SwipeableDrawer from '@mui/material/SwipeableDrawer';
import { useEffect } from 'react';
import useLayoutSettings from '@fuse/core/Layout/useLayoutSettings';
import useThemeMediaQuery from '@fuse/hooks/useThemeMediaQuery';
import NavbarStyle1Content from './NavbarStyle1Content';
import { Layout1ConfigDefaultsType } from '@/components/theme-layouts/layout1/Layout1Config';
import { useNavbarContext } from '@/components/theme-layouts/components/navbar/contexts/NavbarContext/useNavbarContext';

const navbarWidth = 280;

type StyledNavBarProps = {
	theme?: Theme;
	open: boolean;
	position: string;
};

const StyledNavBar = styled('div')<StyledNavBarProps>(({ theme }) => ({
	minWidth: navbarWidth,
	width: navbarWidth,
	maxWidth: navbarWidth,
	variants: [
		{
			props: ({ open }) => !open,
			style: {
				transition: theme.transitions.create('margin', {
					easing: theme.transitions.easing.easeOut,
					duration: theme.transitions.duration.leavingScreen
				})
			}
		},
		{
			props: ({ open, position }) => !open && position === 'left',
			style: {
				marginLeft: `-${navbarWidth}px`
			}
		},
		{
			props: ({ open, position }) => !open && position === 'right',
			style: {
				marginRight: `-${navbarWidth}px`
			}
		},
		{
			props: ({ open }) => open,
			style: {
				transition: theme.transitions.create('margin', {
					easing: theme.transitions.easing.easeOut,
					duration: theme.transitions.duration.enteringScreen
				})
			}
		},
		{
			props: ({ open, position }) => open && position === 'left',
			style: {
				borderRight: `1px solid ${theme.vars.palette.divider}`
			}
		},
		{
			props: ({ open, position }) => open && position === 'right',
			style: {
				borderLeft: `1px solid ${theme.vars.palette.divider}`
			}
		}
	]
}));

const StyledNavBarMobile = styled(SwipeableDrawer)(() => ({
	'& .MuiDrawer-paper': {
		minWidth: navbarWidth,
		width: navbarWidth,
		maxWidth: navbarWidth
	}
}));

/**
 * The navbar style 1.
 */
function NavbarStyle1() {
	const isMobile = useThemeMediaQuery((theme) => theme.breakpoints.down('lg'));

	const settings = useLayoutSettings();
	const config = settings.config as Layout1ConfigDefaultsType;

	const {
		isOpen: isNavbarOpen,
		mobileOpen: isNavbarMobileOpen,
		reset: resetNavbar,
		closeMobileNavbar
	} = useNavbarContext();

	useEffect(() => {
		return () => {
			resetNavbar();
		};
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, []);

	return (
		<>
			{!isMobile && (
				<StyledNavBar
					className="sticky top-0 z-20 h-screen flex-auto shrink-0 flex-col"
					open={isNavbarOpen}
					position={config.navbar?.position || 'left'}
				>
					<NavbarStyle1Content />
				</StyledNavBar>
			)}

			{isMobile && (
				<StyledNavBarMobile
					classes={{
						paper: 'flex-col flex-auto h-full'
					}}
					anchor={(config.navbar?.position || 'left') as 'left' | 'top' | 'right' | 'bottom'}
					variant="temporary"
					open={isNavbarMobileOpen}
					onClose={() => closeMobileNavbar()}
					onOpen={() => {}}
					disableSwipeToOpen
					ModalProps={{
						keepMounted: true // Better open performance on mobile.
					}}
				>
					<NavbarStyle1Content />
				</StyledNavBarMobile>
			)}
		</>
	);
}

export default NavbarStyle1;
