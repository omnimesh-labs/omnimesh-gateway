import { ThemeType } from '@fuse/core/Settings/Settings';
import { createTheme, getContrastRatio, Theme, ThemeOptions } from '@mui/material/styles';
import _ from 'lodash';
import { defaultThemeOptions, extendThemeWithMixins, mustHaveThemeOptions } from '@fuse/default-settings';
import { darkPaletteText, lightPaletteText } from '@/configs/themesConfig';
import useSettings from './useSettings';

type Direction = 'ltr' | 'rtl';

// Function to generate the MUI theme
const generateMuiTheme = (theme: ThemeType, direction: Direction, prefix?: string): Theme => {
	const mergedTheme = _.merge({}, defaultThemeOptions, theme, mustHaveThemeOptions, {
		cssVariables: { cssVarPrefix: prefix }
	}) as ThemeOptions;
	const themeOptions = {
		...mergedTheme,
		mixins: extendThemeWithMixins(mergedTheme),
		direction
	} as ThemeOptions;
	return createTheme(themeOptions);
};

// Custom hooks for selecting themes
export const useMainTheme = (): Theme => {
	const { data: current } = useSettings();
	return generateMuiTheme(current.theme.main, current.direction);
};

export const useNavbarTheme = (): Theme => {
	const { data: current } = useSettings();
	return generateMuiTheme(current.theme.navbar, current.direction, 'navbar');
};

export const useToolbarTheme = (): Theme => {
	const { data: current } = useSettings();
	return generateMuiTheme(current.theme.toolbar, current.direction, 'toolbar');
};

export const useFooterTheme = (): Theme => {
	const { data: current } = useSettings();
	return generateMuiTheme(current.theme.footer, current.direction, 'footer');
};

// Helper functions for theme mode changes
export const changeThemeMode = (theme: ThemeType, mode: 'dark' | 'light'): ThemeType => {
	const modes = {
		dark: {
			palette: {
				mode: 'dark',
				divider: 'rgba(241,245,249,.12)',
				background: {
					paper: '#1E2125',
					default: '#121212'
				},
				text: darkPaletteText
			}
		},
		light: {
			palette: {
				mode: 'light',
				divider: '#e2e8f0',
				background: {
					paper: '#FFFFFF',
					default: '#F7F7F7'
				},
				text: lightPaletteText
			}
		}
	};
	return _.merge({}, theme, modes[mode]);
};

// Custom hook for contrast theme
export const useContrastMainTheme = (bgColor: string): Theme => {
	const isDark = (color: string): boolean => getContrastRatio(color, '#ffffff') >= 3;
	const darkTheme = useMainThemeDark();
	const lightTheme = useMainThemeLight();

	return isDark(bgColor) ? darkTheme : lightTheme;
};

export const useMainThemeDark = (): Theme => {
	const { data: current } = useSettings();
	return generateMuiTheme(changeThemeMode(current.theme.main, 'dark'), current.direction, 'main-dark');
};

export const useMainThemeLight = (): Theme => {
	const { data: current } = useSettings();
	return generateMuiTheme(changeThemeMode(current.theme.main, 'light'), current.direction, 'main-light');
};
