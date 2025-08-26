import useThemeMediaQuery from '@fuse/hooks/useThemeMediaQuery';
import NavbarToggleFab from 'src/components/theme-layouts/components/navbar/NavbarToggleFab';
import { useNavbarContext } from '../../components/navbar/contexts/NavbarContext/useNavbarContext';

type NavbarToggleFabLayout2Props = {
	className?: string;
};

/**
 * The navbar toggle fab layout 2.
 */
function NavbarToggleFabLayout2(props: NavbarToggleFabLayout2Props) {
	const { className } = props;
	const isMobile = useThemeMediaQuery((theme) => theme.breakpoints.down('lg'));
	const { toggleMobileNavbar, toggleNavbar } = useNavbarContext();

	return (
		<NavbarToggleFab
			className={className}
			onClick={() => {
				if (isMobile) {
					toggleMobileNavbar();
				} else {
					toggleNavbar();
				}
			}}
		/>
	);
}

export default NavbarToggleFabLayout2;
