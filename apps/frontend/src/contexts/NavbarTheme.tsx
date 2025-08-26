import { useNavbarTheme } from '@fuse/core/Settings/hooks/themeHooks';
import Theme from '@fuse/core/Theme';

type NavbarThemeProps = {
	children: React.ReactNode;
};

function NavbarTheme({ children }: NavbarThemeProps) {
	const navbarTheme = useNavbarTheme();

	return <Theme theme={navbarTheme}>{children}</Theme>;
}

export default NavbarTheme;
