import { useEffect } from 'react';
import useThemeMediaQuery from '@fuse/hooks/useThemeMediaQuery';
import usePathname from '@fuse/hooks/usePathname';
import useLayoutSettings from '@fuse/core/Layout/useLayoutSettings';
import NavbarToggleFabLayout1 from './NavbarToggleFabLayout1';
import NavbarStyle1 from './navbar/style-1/NavbarStyle1';
import NavbarStyle2 from './navbar/style-2/NavbarStyle2';
import { useNavbarContext } from '../../components/navbar/contexts/NavbarContext/useNavbarContext';
import NavbarTheme from '@/contexts/NavbarTheme';

/**
 * The navbar wrapper layout 1.
 */
function NavbarWrapperLayout1() {
	const { config } = useLayoutSettings();
	const { closeMobileNavbar, isOpen: isNavbarOpen } = useNavbarContext();

	const pathname = usePathname();

	const isMobile = useThemeMediaQuery((theme) => theme.breakpoints.down('lg'));

	useEffect(() => {
		if (isMobile) {
			closeMobileNavbar();
		}
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [pathname, isMobile]);

	// Ensure navbar config exists with default values
	const navbarStyle = config.navbar?.style || 'style-1';
	const navbarDisplay = config.navbar?.display !== false;
	const toolbarDisplay = config.toolbar?.display !== false;

	return (
		<>
			<NavbarTheme>
				<>
					{navbarStyle === 'style-1' && <NavbarStyle1 />}
					{navbarStyle === 'style-2' && <NavbarStyle2 />}
				</>
			</NavbarTheme>
			{navbarDisplay && !toolbarDisplay && !isNavbarOpen && <NavbarToggleFabLayout1 />}
		</>
	);
}

export default NavbarWrapperLayout1;
