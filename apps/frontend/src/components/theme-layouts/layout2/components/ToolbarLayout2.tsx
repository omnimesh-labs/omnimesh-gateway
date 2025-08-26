import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import clsx from 'clsx';
import { memo } from 'react';
import NavbarToggleButton from 'src/components/theme-layouts/components/navbar/NavbarToggleButton';
import useLayoutSettings from '@fuse/core/Layout/useLayoutSettings';
import AdjustFontSize from '../../components/AdjustFontSize';
import FullScreenToggle from '../../components/FullScreenToggle';
import LanguageSwitcher from '../../components/LanguageSwitcher';
import NavigationShortcuts from '../../components/navigation/NavigationShortcuts';
import NavigationSearch from '../../components/navigation/NavigationSearch';
import UserMenu from '../../components/UserMenu';
import QuickPanelToggleButton from '../../components/quickPanel/QuickPanelToggleButton';
import useThemeMediaQuery from '../../../../@fuse/hooks/useThemeMediaQuery';
import ToolbarTheme from '@/contexts/ToolbarTheme';

type ToolbarLayout2Props = {
	className?: string;
};

/**
 * The toolbar layout 2.
 */
function ToolbarLayout2(props: ToolbarLayout2Props) {
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
				<Toolbar className="container min-h-12 p-0 md:min-h-16 lg:px-8">
					<div className="flex flex-1 gap-2">
						{config.navbar.display && isMobile && <NavbarToggleButton className="h-9 w-9 p-0" />}

						{!isMobile && <NavigationShortcuts />}
					</div>

					<div className="flex items-center overflow-x-auto px-2 py-2 md:px-4">
						<LanguageSwitcher />
						<AdjustFontSize />
						<FullScreenToggle />
						<NavigationSearch />
						<QuickPanelToggleButton />
					</div>
					{!isMobile && (
						<UserMenu
							className="border-divider border border-solid"
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
							dense
						/>
					)}
				</Toolbar>
			</AppBar>
		</ToolbarTheme>
	);
}

export default memo(ToolbarLayout2);
