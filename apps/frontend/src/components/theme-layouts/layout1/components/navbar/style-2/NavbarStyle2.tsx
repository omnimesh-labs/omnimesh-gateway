import { styled } from '@mui/material/styles';
import SwipeableDrawer from '@mui/material/SwipeableDrawer';
import GlobalStyles from '@mui/material/GlobalStyles';
import { Theme } from '@mui/system';
import clsx from 'clsx';
import { useEffect } from 'react';
import useLayoutSettings from '@fuse/core/Layout/useLayoutSettings';
import useThemeMediaQuery from '@fuse/hooks/useThemeMediaQuery';
import NavbarStyle2Content from './NavbarStyle2Content';
import { Layout1ConfigDefaultsType } from '@/components/theme-layouts/layout1/Layout1Config';
import { useNavbarContext } from '@/components/theme-layouts/components/navbar/contexts/NavbarContext/useNavbarContext';

const navbarWidth = 48;
const panelWidth = 280;

type StyledNavBarProps = {
	theme?: Theme;
	open?: boolean;
	folded?: number;
	position?: string;
	className?: string;
	anchor?: string;
};

const StyledNavBar = styled('div')<StyledNavBarProps>(({ theme }) => ({
	minWidth: navbarWidth,
	width: navbarWidth,
	maxWidth: navbarWidth,
	variants: [
		{
			props: ({ open, position }) => !open && position === 'left',
			style: {
				marginLeft: -navbarWidth
			}
		},
		{
			props: ({ open, position }) => !open && position === 'right',
			style: {
				marginRight: -navbarWidth
			}
		},
		{
			props: ({ folded }) => !folded,
			style: {
				minWidth: navbarWidth + panelWidth,
				width: navbarWidth + panelWidth,
				maxWidth: navbarWidth + panelWidth,
				'& #fuse-navbar-panel': {
					opacity: '1!important',
					pointerEvents: 'initial!important'
				}
			}
		},
		{
			props: ({ folded, open, position }) => !folded && !open && position === 'left',
			style: {
				marginLeft: `${-(navbarWidth + panelWidth)}px!important`
			}
		},
		{
			props: ({ folded, open, position }) => !folded && !open && position === 'right',
			style: {
				marginRight: `${-(navbarWidth + panelWidth)}px!important`
			}
		},
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
			props: ({ open }) => open,
			style: {
				transition: theme.transitions.create('margin', {
					easing: theme.transitions.easing.easeOut,
					duration: theme.transitions.duration.enteringScreen
				})
			}
		}
	]
}));

const StyledNavBarMobile = styled(SwipeableDrawer)<StyledNavBarProps>(() => ({
	'& .MuiDrawer-paper': {
		'& #fuse-navbar-side-panel': {
			minWidth: 'auto',
			wdith: 'auto'
		},
		'& #fuse-navbar-panel': {
			opacity: '1!important',
			pointerEvents: 'initial!important'
		}
	},
	'& .user-menu': {
		minWidth: 56,
		width: 56,
		'& .title': {
			opacity: 0
		},
		'& .subtitle': {
			opacity: 0
		},
		'& .info-icon': {
			opacity: 0
		},
		'& .arrow': {
			opacity: 0
		}
	}
}));

type NavbarStyle2Props = {
	className?: string;
};

/**
 * The navbar style 3.
 */
function NavbarStyle2(props: NavbarStyle2Props) {
	const { className = '' } = props;
	const settings = useLayoutSettings();
	const config = settings.config as Layout1ConfigDefaultsType;
	const isMobile = useThemeMediaQuery((theme) => theme.breakpoints.down('lg'));

	const { folded } = config.navbar;

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
	}, [resetNavbar]);

	return (
		<>
			<GlobalStyles
				styles={(theme) => ({
					'& #fuse-navbar-side-panel': {
						width: navbarWidth,
						minWidth: navbarWidth,
						maxWidth: navbarWidth
					},
					'& #fuse-navbar-panel': {
						maxWidth: '100%',
						width: panelWidth,
						borderRight: `1px solid ${theme.vars.palette.divider}!important`,
						borderLeft: `1px solid ${theme.vars.palette.divider}!important`,
						[theme.breakpoints.up('lg')]: {
							minWidth: panelWidth,
							maxWidth: 'initial'
						}
					}
				})}
			/>

			{!isMobile && (
				<StyledNavBar
					open={isNavbarOpen}
					folded={folded ? 1 : 0}
					position={config.navbar.position}
					className={clsx('sticky top-0 z-20 h-screen flex-auto shrink-0 flex-col', className)}
				>
					<NavbarStyle2Content />
				</StyledNavBar>
			)}

			{isMobile && (
				<StyledNavBarMobile
					classes={{
						paper: clsx('h-screen w-auto max-w-full flex-auto flex-col overflow-hidden', className)
					}}
					anchor={config.navbar.position as 'left' | 'right'}
					variant="temporary"
					open={isNavbarMobileOpen}
					onClose={() => closeMobileNavbar()}
					onOpen={() => {}}
					disableSwipeToOpen
					ModalProps={{
						keepMounted: true // Better open performance on mobile.
					}}
				>
					<NavbarStyle2Content />
				</StyledNavBarMobile>
			)}
		</>
	);
}

export default NavbarStyle2;
