import React, { useState } from 'react';
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import IconButton from '@mui/material/IconButton';
import SvgIcon from '@fuse/core/SvgIcon';
import { ThemeOption } from '@fuse/core/ThemeSelector/ThemePreview';
import { useMainTheme } from '@fuse/core/Settings/hooks/themeHooks';
import useSettings from '@fuse/core/Settings/hooks/useSettings';
import { SettingsConfigType } from '@fuse/core/Settings/Settings';
// import { useSnackbar } from 'notistack';

type LightDarkModeToggleProps = {
	className?: string;
	lightTheme: ThemeOption;
	darkTheme: ThemeOption;
};

function LightDarkModeToggle(props: LightDarkModeToggleProps) {
	const { className = '', lightTheme, darkTheme } = props;
	const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
	const { setSettings } = useSettings();
	// const { isGuest, updateUserSettings } = useUser();
	// const { enqueueSnackbar } = useSnackbar();
	const mainTheme = useMainTheme();

	const handleClick = (event: React.MouseEvent<HTMLButtonElement>) => {
		setAnchorEl(event.currentTarget);
	};

	const handleClose = () => {
		setAnchorEl(null);
	};

	const handleSelectionChange = (selection: 'light' | 'dark') => {
		if (selection === 'light') {
			handleThemeSelect(lightTheme);
		} else {
			handleThemeSelect(darkTheme);
		}

		handleClose();
	};

	async function handleThemeSelect(_theme: ThemeOption) {
		const _newSettings = setSettings({ theme: { ..._theme?.section } } as Partial<SettingsConfigType>);

		/**
		 * Updating user settings disabled for demonstration purposes
		 * The request is made to the mock API and will not persist the changes
		 * You can enable it by removing the comment block below when using a real API
		 * */
		/* if (!isGuest) {
			const updatedUserData = await updateUserSettings(_newSettings);

			if (updatedUserData) {
				enqueueSnackbar('User settings saved.', {
					variant: 'success'
				});
			}
		} */
	}

	return (
		<>
			<IconButton
				aria-controls="light-dark-toggle-menu"
				aria-haspopup="true"
				onClick={handleClick}
				className={className}
			>
				{mainTheme.palette.mode === 'light' && <SvgIcon>lucide:sun</SvgIcon>}
				{mainTheme.palette.mode === 'dark' && <SvgIcon>lucide:moon</SvgIcon>}
			</IconButton>
			<Menu
				id="light-dark-toggle-menu"
				anchorEl={anchorEl}
				keepMounted
				open={Boolean(anchorEl)}
				onClose={handleClose}
			>
				<MenuItem
					selected={mainTheme.palette.mode === 'light'}
					onClick={() => handleSelectionChange('light')}
				>
					Light
				</MenuItem>
				<MenuItem
					selected={mainTheme.palette.mode === 'dark'}
					onClick={() => handleSelectionChange('dark')}
				>
					Dark
				</MenuItem>
			</Menu>
		</>
	);
}

export default LightDarkModeToggle;
