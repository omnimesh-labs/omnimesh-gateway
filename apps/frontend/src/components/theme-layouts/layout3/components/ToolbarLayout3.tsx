import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import clsx from 'clsx';
import { memo } from 'react';
import NavbarToggleButton from 'src/components/theme-layouts/components/navbar/NavbarToggleButton';
import useLayoutSettings from '@fuse/core/Layout/useLayoutSettings';
import AdjustFontSize from '../../components/AdjustFontSize';
import FullScreenToggle from '../../components/FullScreenToggle';
import LanguageSwitcher from '../../components/LanguageSwitcher';
import NavigationSearch from '../../components/navigation/NavigationSearch';
import UserMenu from '../../components/UserMenu';
import QuickPanelToggleButton from '../../components/quickPanel/QuickPanelToggleButton';
import Logo from '../../components/Logo';
import useThemeMediaQuery from '../../../../@fuse/hooks/useThemeMediaQuery';
import ToolbarTheme from '@/contexts/ToolbarTheme';

type ToolbarLayout3Props = {
	className?: string;
};

/**
 * The toolbar layout 3.
 */
function ToolbarLayout3(props: ToolbarLayout3Props) {
	const { className = '' } = props;
	const { config } = useLayoutSettings();
	const isMobile = useThemeMediaQuery((theme) => theme.breakpoints.down('lg'));

	return (
		<ToolbarTheme>
			<AppBar
				id="fuse-toolbar"
				className={clsx('relative z-20 flex border-b', className)}
				color="default"
				sx={{ backgroundColor: (theme) => theme.vars.palette.background.paper }}
				elevation={0}
			>
				<Toolbar className="container min-h-12 gap-2 p-0 md:min-h-16 lg:px-8">
					<div className={clsx('flex gap-2')}>
						{config.navbar.display && isMobile && <NavbarToggleButton className="h-9 w-9 p-0" />}

						{!isMobile && <Logo />}
					</div>

					<div className="flex flex-auto items-center justify-end overflow-x-auto">
						{isMobile && <NavigationSearch />}
						<LanguageSwitcher />
						<AdjustFontSize />
						<FullScreenToggle />
						<QuickPanelToggleButton />
					</div>
					{!isMobile && (
						<UserMenu
							className="border-divider border border-solid"
							dense
							arrowIcon="lucide:chevron-down"
							popoverProps={{
								anchorOrigin: {
									vertical: 'bottom',
									horizontal: 'center'
								},
								transformOrigin: {
									vertical: 'top',
									horizontal: 'center'
								}
							}}
						/>
					)}
				</Toolbar>
			</AppBar>
		</ToolbarTheme>
	);
}

export default memo(ToolbarLayout3);
