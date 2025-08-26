import IconButton from '@mui/material/IconButton';
import _ from 'lodash';
import useThemeMediaQuery from '@fuse/hooks/useThemeMediaQuery';
import SvgIcon from '@fuse/core/SvgIcon';
import { IconButtonProps } from '@mui/material/IconButton';
import useLayoutSettings from '@fuse/core/Layout/useLayoutSettings';
import useSettings from '@fuse/core/Settings/hooks/useSettings';
import { useNavbarContext } from './contexts/NavbarContext/useNavbarContext';

export type NavbarToggleButtonProps = IconButtonProps;

/**
 * The navbar toggle button.
 */
function NavbarToggleButton(props: NavbarToggleButtonProps) {
	const {
		className = 'h-7 w-7 border border-divider',
		children = <SvgIcon>lucide:panel-left</SvgIcon>,
		...rest
	} = props;

	const { toggleMobileNavbar, toggleNavbar } = useNavbarContext();
	const isMobile = useThemeMediaQuery((theme) => theme.breakpoints.down('lg'));
	const { config } = useLayoutSettings();
	const { setSettings } = useSettings();

	return (
		<IconButton
			size="small"
			onClick={() => {
				if (isMobile) {
					toggleMobileNavbar();
				} else if (config?.navbar?.style === 'style-2') {
					setSettings(_.set({}, 'layout.config.navbar.folded', !config?.navbar?.folded));
				} else {
					toggleNavbar();
				}
			}}
			{...rest}
			className={className}
		>
			{children}
		</IconButton>
	);
}

export default NavbarToggleButton;
