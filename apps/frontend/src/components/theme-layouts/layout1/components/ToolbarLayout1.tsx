import Toolbar from '@mui/material/Toolbar';
import clsx from 'clsx';
import { memo } from 'react';
import NavbarToggleButton from 'src/components/theme-layouts/components/navbar/NavbarToggleButton';
import themeOptions from 'src/configs/themeOptions';
import find from 'lodash/find';
import LightDarkModeToggle from 'src/components/LightDarkModeToggle';
import useLayoutSettings from '@fuse/core/Layout/useLayoutSettings';
import FullScreenToggle from '../../components/FullScreenToggle';
// import NavigationShortcuts from '../../components/navigation/NavigationShortcuts';
// import NavigationSearch from '../../components/navigation/NavigationSearch';
import UserMenu from '../../components/UserMenu';
import { Layout1ConfigDefaultsType } from '@/components/theme-layouts/layout1/Layout1Config';
import useThemeMediaQuery from '../../../../@fuse/hooks/useThemeMediaQuery';
import AppBar from '@mui/material/AppBar';
import Divider from '@mui/material/Divider';
import ToolbarTheme from 'src/contexts/ToolbarTheme';

type ToolbarLayout1Props = {
	className?: string;
};

/**
 * The toolbar layout 1.
 */
function ToolbarLayout1(props: ToolbarLayout1Props) {
	const { className } = props;

	const settings = useLayoutSettings();
	const config = settings.config as Layout1ConfigDefaultsType;
	const isMobile = useThemeMediaQuery((theme) => theme.breakpoints.down('lg'));

	return (
		<ToolbarTheme>
			<AppBar
				id="fuse-toolbar"
				className={clsx('relative z-20 flex', className)}
				sx={(theme) => ({
					backgroundColor: theme.vars.palette.background.default,
					color: theme.vars.palette.text.primary
				})}
			>
				<Toolbar className="min-h-12 p-0 md:min-h-16">
					<div className="flex flex-1 items-center gap-3 px-2 md:px-4">
						{config.navbar.display && config.navbar.position === 'left' && (
							<>
								<NavbarToggleButton />

								<Divider
									orientation="vertical"
									flexItem
									variant="middle"
								/>
							</>
						)}

						{/* {!isMobile && <NavigationShortcuts />} */}
					</div>

					<div className="flex items-center overflow-x-auto px-2 py-2 md:px-4">
						<FullScreenToggle />
						<LightDarkModeToggle
							lightTheme={find(themeOptions, { id: 'Default' })}
							darkTheme={find(themeOptions, { id: 'Default Dark' })}
						/>
						{/* <NavigationSearch /> */}
						<UserMenu
							className="ml-2"
							dense={true}
							onlyAvatar={isMobile}
							popoverProps={{
								anchorOrigin: {
									vertical: 'bottom',
									horizontal: 'right'
								},
								transformOrigin: {
									vertical: 'top',
									horizontal: 'right'
								}
							}}
						/>
					</div>

					{config.navbar.display && config.navbar.position === 'right' && (
						<>
							{!isMobile && (
								<>
									<Divider
										orientation="vertical"
										flexItem
										variant="middle"
									/>
									<NavbarToggleButton />
								</>
							)}

							{isMobile && <NavbarToggleButton className="h-10 w-10 p-0 sm:mx-2" />}
						</>
					)}
				</Toolbar>
			</AppBar>
		</ToolbarTheme>
	);
}

export default memo(ToolbarLayout1);
