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
	const { setSettings } = useSettings();
	// const { isGuest, updateUserSettings } = useUser();
	// const { enqueueSnackbar } = useSnackbar();
	const mainTheme = useMainTheme();

	const handleToggle = () => {
		const isCurrentlyLight = mainTheme.palette.mode === 'light';
		const targetTheme = isCurrentlyLight ? darkTheme : lightTheme;
		handleThemeSelect(targetTheme);
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
		<IconButton
			aria-label="Toggle light/dark theme"
			onClick={handleToggle}
			className={className}
		>
			{mainTheme.palette.mode === 'light' && <SvgIcon>lucide:sun</SvgIcon>}
			{mainTheme.palette.mode === 'dark' && <SvgIcon>lucide:moon</SvgIcon>}
		</IconButton>
	);
}

export default LightDarkModeToggle;
