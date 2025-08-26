import { dark, skyBlue } from '@fuse/colors';
import { blueGrey } from '@mui/material/colors';
import { ThemesType } from '@fuse/core/Settings/Settings';

/**
 * The lightPaletteText object defines the text color palette for the light theme.
 */
export const lightPaletteText = {
	primary: 'rgb(17, 24, 39)',
	secondary: 'rgb(107, 114, 128)',
	disabled: 'rgb(149, 156, 169)'
};

/**
 * The darkPaletteText object defines the text color palette for the dark theme.
 */
export const darkPaletteText = {
	primary: 'rgb(255,255,255)',
	secondary: 'rgb(148, 163, 184)',
	disabled: 'rgb(156, 163, 175)'
};

/**
 * Shared neutral ramp & status colours
 * Light theme
 */
const neutralsLightTheme = {
	grey: {
		50: '#F9FAFB',
		100: '#F3F4F6',
		200: '#E5E7EB',
		300: '#D1D5DB',
		400: '#9CA3AF',
		500: '#6B7280',
		600: '#4B5563',
		700: '#374151',
		800: '#272F3C',
		900: '#1F232B',
		A100: '#F3F4F6',
		A200: '#E5E7EB',
		A400: '#6B7280',
		A700: '#374151'
	},
	success: { main: '#22C55E', light: '#4ADE80', dark: '#15803D', contrastText: '#1F232B' },
	info: { main: '#3B82F6', light: '#60A5FA', dark: '#1D4ED8', contrastText: '#FFFFFF' },
	warning: { main: '#F59E0B', light: '#FBBF24', dark: '#B45309', contrastText: '#1F232B' },
	error: { main: '#EF4444', light: '#F87171', dark: '#B91C1C', contrastText: '#FFFFFF' }
};

/**
 * Shared neutral ramp & status colours
 * Dark theme
 */
const neutralsDarkTheme = {
	grey: {
		50: '#101214',
		100: '#16181D',
		200: '#1E2026',
		300: '#25282F',
		400: '#2D3139',
		500: '#6B7280',
		600: '#8A8F99',
		700: '#A5ABB5',
		800: '#CDD1D9',
		900: '#E5E7EB',
		A100: '#16181D',
		A200: '#1E2026',
		A400: '#2D3139',
		A700: '#6B7280'
	},
	success: { main: '#22C55E', light: '#4ADE80', dark: '#15803D', contrastText: '#0F1115' },
	info: { main: '#3B82F6', light: '#60A5FA', dark: '#1D4ED8', contrastText: '#FFFFFF' },
	warning: { main: '#F59E0B', light: '#FBBF24', dark: '#B45309', contrastText: '#0F1115' },
	error: { main: '#EF4444', light: '#F87171', dark: '#B91C1C', contrastText: '#FFFFFF' }
};

/**
 * The themesConfig object is a configuration object for the color themes of the  application.
 */
export const themesConfig: ThemesType = {
	default: {
		palette: {
			mode: 'light',
			primary: { main: '#1F232B', light: '#363B44', dark: '#0F1115', contrastText: '#FFFFFF' },
			secondary: { main: '#1565C0', light: '#3C83D6', dark: '#0E4B90', contrastText: '#FFFFFF' },
			...neutralsLightTheme,
			text: { primary: '#1F232B', secondary: '#4B5563', disabled: '#9CA3AF' },
			background: { default: '#F6F7F8', paper: '#FFFFFF' },
			divider: '#E5E7EB',
			action: {
				active: '#4B5563',
				hover: '#E5E7EB',
				selected: '#D1D5DB',
				disabled: '#BFC4CC',
				disabledBackground: '#F6F7F8',
				focus: '#D1D5DB'
			}
		}
	},
	defaultDark: {
		palette: {
			mode: 'dark',
			primary: {
				main: '#2B2F38',
				light: '#414652',
				dark: '#16191E',
				contrastText: '#FFFFFF'
			},
			secondary: {
				main: '#3C83D6',
				light: '#5A9AF0',
				dark: '#1D5EB0',
				contrastText: '#FFFFFF'
			},
			...neutralsDarkTheme,
			text: {
				primary: '#E5E7EB',
				secondary: '#A5ABB5',
				disabled: '#6B7280'
			},
			background: {
				default: '#0F1115',
				paper: '#1A1D22'
			},
			divider: '#2D3139',
			action: {
				active: '#CDD1D9',
				hover: 'rgba(255,255,255,0.08)',
				selected: 'rgba(255,255,255,0.16)',
				disabled: 'rgba(255,255,255,0.3)',
				disabledBackground: 'rgba(255,255,255,0.12)',
				focus: 'rgba(255,255,255,0.12)'
			}
		}
	},
	defaultNavbar: {
		palette: {
			mode: 'light',
			...neutralsLightTheme,
			primary: { main: '#1F232B', light: '#363B44', dark: '#0F1115', contrastText: '#FFFFFF' },
			secondary: { main: '#1565C0', light: '#3C83D6', dark: '#0E4B90', contrastText: '#FFFFFF' },
			text: { primary: '#1F232B', secondary: '#4B5563', disabled: '#9CA3AF' },
			background: { default: '#E5E7EB', paper: '#FFFFFF' },
			divider: '#E5E7EB',
			action: {
				active: '#4B5563',
				hover: '#E5E7EB',
				selected: '#D1D5DB',
				disabled: '#BFC4CC',
				disabledBackground: '#F6F7F8',
				focus: '#D1D5DB'
			}
		}
	},
	legacyDefault: {
		palette: {
			mode: 'light',
			divider: 'rgba(0, 0, 0, 0.12)',
			text: {
				primary: '#212121',
				secondary: '#5F6368'
			},
			common: {
				black: '#000000',
				white: '#FFFFFF'
			},
			primary: {
				light: '#536D89',
				main: '#0A74DA',
				dark: '#00418A',
				contrastText: '#FFFFFF'
			},
			secondary: {
				light: '#6BC9F7',
				main: '#00A4EF',
				dark: '#0078D7',
				contrastText: '#FFFFFF'
			},
			background: {
				paper: '#F4F4F4',
				default: '#E8E8E8'
			},
			error: {
				light: '#FFCDD2',
				main: '#D32F2F',
				dark: '#B71C1C',
				contrastText: '#FFFFFF'
			}
		}
	},
	legacyDark: {
		palette: {
			mode: 'dark',
			divider: 'rgba(255, 255, 255, 0.12)',
			text: {
				primary: '#E0E0E0',
				secondary: '#B0BEC5'
			},
			common: {
				black: '#000000',
				white: '#FFFFFF'
			},
			primary: {
				light: '#536D89',
				main: '#0A74DA',
				dark: '#00418A',
				contrastText: '#FFFFFF'
			},
			secondary: {
				light: '#6BC9F7',
				main: '#00A4EF',
				dark: '#0078D7',
				contrastText: '#FFFFFF'
			},
			background: {
				paper: '#1E1E1E',
				default: '#121212'
			},
			error: {
				light: '#FFCDD2',
				main: '#D32F2F',
				dark: '#B71C1C',
				contrastText: '#FFFFFF'
			}
		}
	},
	navyTealPalette: {
		palette: {
			mode: 'light',
			primary: { main: '#0F3C6E', light: '#365E92', dark: '#08254A', contrastText: '#FFFFFF' },
			secondary: { main: '#00B4A6', light: '#33C7BB', dark: '#00867B', contrastText: '#FFFFFF' },
			...neutralsLightTheme,
			text: { primary: '#0F3C6E', secondary: '#4B5563', disabled: '#9CA3AF' },
			background: { default: '#F6F8FA', paper: '#FFFFFF' },
			divider: '#E5E7EB',
			action: {
				active: '#4B5563',
				hover: '#E5E7EB',
				selected: '#D1D5DB',
				disabled: '#BFC4CC',
				disabledBackground: '#F6F7F8',
				focus: '#D1D5DB'
			}
		}
	},
	indigoAmberPalette: {
		palette: {
			mode: 'light',
			primary: { main: '#21255F', light: '#40448A', dark: '#15183C', contrastText: '#FFFFFF' },
			secondary: { main: '#FFB547', light: '#FFC66D', dark: '#D89330', contrastText: '#FFFFFF' },
			...neutralsLightTheme,
			text: { primary: '#21255F', secondary: '#4B5563', disabled: '#9CA3AF' },
			background: { default: '#F7F8FC', paper: '#FFFFFF' },
			divider: '#E5E7EB',
			action: {
				active: '#4B5563',
				hover: '#E5E7EB',
				selected: '#D1D5DB',
				disabled: '#BFC4CC',
				disabledBackground: '#F6F7F8',
				focus: '#D1D5DB'
			}
		}
	},
	darkBlueSilver: {
		palette: {
			mode: 'light',
			primary: {
				main: '#0D47A1',
				light: '#5472D3',
				dark: '#002171',
				contrastText: darkPaletteText.primary
			},
			secondary: {
				main: '#B0BEC5',
				light: '#E2F1F8',
				dark: '#808E95',
				contrastText: lightPaletteText.primary
			},
			background: {
				paper: '#FFFFFF',
				default: '#f1f5f9'
			},
			text: lightPaletteText,
			divider: '#d8d9da'
		}
	},
	darkBlueSilverDark: {
		palette: {
			mode: 'dark',
			primary: {
				main: '#0D47A1',
				light: '#5472D3',
				dark: '#002171',
				contrastText: darkPaletteText.primary
			},
			secondary: {
				main: '#B0BEC5',
				light: '#E2F1F8',
				dark: '#808E95',
				contrastText: lightPaletteText.primary
			},
			background: {
				default: '#263238',
				paper: '#2d3940'
			},
			text: darkPaletteText,
			divider: '#42474d'
		}
	},
	slateCrimson: {
		palette: {
			mode: 'light',
			primary: {
				main: '#37474F',
				light: '#62727B',
				dark: '#102027',
				contrastText: darkPaletteText.primary
			},
			secondary: {
				main: '#D32F2F',
				light: '#FF6659',
				dark: '#9A0007',
				contrastText: darkPaletteText.primary
			},
			background: {
				default: '#e6e6e6',
				paper: '#f2f2f2'
			},
			text: lightPaletteText,
			divider: '#d9d9d9'
		}
	},
	slateCrimsonDark: {
		palette: {
			mode: 'dark',
			primary: {
				main: '#37474F',
				light: '#62727B',
				dark: '#102027',
				contrastText: darkPaletteText.primary
			},
			secondary: {
				main: '#D32F2F',
				light: '#FF6659',
				dark: '#9A0007',
				contrastText: darkPaletteText.primary
			},
			background: {
				default: '#212121',
				paper: '#2e2e2e'
			},
			text: darkPaletteText,
			divider: '#3a3d40'
		}
	},
	emeraldGold: {
		palette: {
			mode: 'light',
			primary: {
				main: '#00695C',
				light: '#439889',
				dark: '#003D33',
				contrastText: darkPaletteText.primary
			},
			secondary: {
				main: '#FFD740',
				light: '#FFFF74',
				dark: '#C8A600',
				contrastText: lightPaletteText.primary
			},
			background: {
				default: '#dcf2f2',
				paper: '#f2fdfa'
			},
			text: lightPaletteText,
			divider: '#b3c4c3'
		}
	},
	emeraldGoldDark: {
		palette: {
			mode: 'dark',
			primary: {
				main: '#00695C',
				light: '#439889',
				dark: '#003D33',
				contrastText: darkPaletteText.primary
			},
			secondary: {
				main: '#FFD740',
				light: '#FFFF74',
				dark: '#C8A600',
				contrastText: lightPaletteText.primary
			},
			background: {
				default: '#004D40',
				paper: '#00544a'
			},
			text: darkPaletteText,
			divider: '#2d6360'
		}
	},
	indigoCoral: {
		palette: {
			mode: 'light',
			primary: {
				main: '#283593',
				light: '#5F5FC4',
				dark: '#001064',
				contrastText: darkPaletteText.primary
			},
			secondary: {
				main: '#FF7043',
				light: '#FFA270',
				dark: '#C63F17',
				contrastText: lightPaletteText.primary
			},
			background: {
				default: '#eaebfb',
				paper: '#f6f7fd'
			},
			text: lightPaletteText,
			divider: '#dcdcf2'
		}
	},
	indigoCoralDark: {
		palette: {
			mode: 'dark',
			primary: {
				main: '#283593',
				light: '#5F5FC4',
				dark: '#001064',
				contrastText: darkPaletteText.primary
			},
			secondary: {
				main: '#FF7043',
				light: '#FFA270',
				dark: '#C63F17',
				contrastText: lightPaletteText.primary
			},
			background: {
				default: '#1A237E',
				paper: '#283593'
			},
			text: darkPaletteText,
			divider: '#4d557e'
		}
	},
	charcoalTeal: {
		palette: {
			mode: 'light',
			primary: {
				main: '#094a43',
				light: '#28635a',
				dark: '#004a41',
				contrastText: darkPaletteText.primary
			},
			secondary: {
				main: '#009688',
				light: '#52C7B8',
				dark: '#00675B',
				contrastText: darkPaletteText.primary
			},
			background: {
				default: '#edf6fa',
				paper: '#f7fcfc'
			},
			text: lightPaletteText,
			divider: '#cee5f0'
		}
	},
	charcoalTealDark: {
		palette: {
			mode: 'dark',
			primary: {
				main: '#455A64',
				light: '#718792',
				dark: '#1C313A',
				contrastText: darkPaletteText.primary
			},
			secondary: {
				main: '#009688',
				light: '#52C7B8',
				dark: '#00675B',
				contrastText: darkPaletteText.primary
			},
			background: {
				default: '#000000',
				paper: '#102027'
			},
			text: darkPaletteText,
			divider: '#2d383d'
		}
	},
	skyBlueOrange: {
		palette: {
			mode: 'light',
			primary: {
				main: '#64B5F6',
				light: '#9BE7FF',
				dark: '#2286C3',
				contrastText: lightPaletteText.primary
			},
			secondary: {
				main: '#faa528',
				light: '#f6ad3f',
				dark: '#cb8721',
				contrastText: lightPaletteText.primary
			},
			background: {
				default: '#F5F5F5',
				paper: '#FFFFFF'
			},
			text: lightPaletteText,
			divider: '#e9e6e0'
		}
	},
	skyBlueOrangeDark: {
		palette: {
			mode: 'dark',
			primary: {
				main: '#64B5F6',
				light: '#9BE7FF',
				dark: '#2286C3',
				contrastText: lightPaletteText.primary
			},
			secondary: {
				main: '#faa528',
				light: '#f6ad3f',
				dark: '#cb8721',
				contrastText: lightPaletteText.primary
			},
			background: {
				default: '#1a1a1a',
				paper: '#333333'
			},
			text: darkPaletteText,
			divider: '#544949'
		}
	},
	softGreenMaroon: {
		palette: {
			mode: 'light',
			primary: {
				main: '#81C784',
				light: '#B2F2B6',
				dark: '#519657',
				contrastText: lightPaletteText.primary
			},
			secondary: {
				main: '#D81B60',
				light: '#FF5C8D',
				dark: '#A00037',
				contrastText: darkPaletteText.primary
			},
			background: {
				default: '#f5f5f5',
				paper: '#fafcfa'
			},
			text: lightPaletteText,
			divider: '#dadeda'
		}
	},
	softGreenMaroonDark: {
		palette: {
			mode: 'dark',
			primary: {
				main: '#81C784',
				light: '#B2F2B6',
				dark: '#519657',
				contrastText: lightPaletteText.primary
			},
			secondary: {
				main: '#D81B60',
				light: '#FF5C8D',
				dark: '#A00037',
				contrastText: darkPaletteText.primary
			},
			background: {
				default: '#1a1a1a',
				paper: '#323332'
			},
			text: darkPaletteText,
			divider: '#505250'
		}
	},
	coolGreyPink: {
		palette: {
			mode: 'light',
			primary: {
				main: '#dde6eb',
				light: '#FFFFFF',
				dark: '#9EA7AA',
				contrastText: lightPaletteText.primary
			},
			secondary: {
				main: '#F06292',
				light: '#FF94C2',
				dark: '#BA2D65',
				contrastText: darkPaletteText.primary
			},
			background: {
				default: '#F5F5F5',
				paper: '#FFFFFF'
			},
			text: lightPaletteText,
			divider: '#e1e1e1'
		}
	},
	coolGreyPinkDark: {
		palette: {
			mode: 'dark',
			primary: {
				main: '#dde6eb',
				light: '#FFFFFF',
				dark: '#9EA7AA',
				contrastText: lightPaletteText.primary
			},
			secondary: {
				main: '#F06292',
				light: '#FF94C2',
				dark: '#BA2D65',
				contrastText: darkPaletteText.primary
			},
			background: {
				default: '#1a1a1a',
				paper: '#292929'
			},
			text: darkPaletteText,
			divider: '#424242'
		}
	},
	legacy: {
		palette: {
			mode: 'light',
			divider: '#e2e8f0',
			text: lightPaletteText,
			common: {
				black: 'rgb(17, 24, 39)',
				white: 'rgb(255, 255, 255)'
			},
			primary: {
				light: dark[200],
				main: dark[500],
				dark: dark[800],
				contrastText: darkPaletteText.primary
			},
			secondary: {
				light: skyBlue[100],
				main: skyBlue[500],
				dark: skyBlue[900],
				contrastText: lightPaletteText.primary
			},
			background: {
				paper: '#FFFFFF',
				default: '#f6f7f9'
			},
			error: {
				light: '#ffcdd2',
				main: '#f44336',
				dark: '#b71c1c'
			}
		}
	},
	light1: {
		palette: {
			mode: 'light',
			divider: '#e2e8f0',
			text: lightPaletteText,
			primary: {
				light: '#b3d1d1',
				main: '#006565',
				dark: '#003737',
				contrastText: darkPaletteText.primary
			},
			secondary: {
				light: '#ffecc0',
				main: '#FFBE2C',
				dark: '#ff9910',
				contrastText: lightPaletteText.primary
			},
			background: {
				paper: '#FFFFFF',
				default: '#F0F7F7'
			},
			error: {
				light: '#ffcdd2',
				main: '#f44336',
				dark: '#b71c1c'
			}
		}
	},
	light2: {
		palette: {
			mode: 'light',
			divider: '#e2e8f0',
			text: lightPaletteText,
			primary: {
				light: '#BBE2DA',
				main: '#1B9E85',
				dark: '#087055',
				contrastText: darkPaletteText.primary
			},
			secondary: {
				light: '#FFD0C1',
				main: '#FF6231',
				dark: '#FF3413',
				contrastText: darkPaletteText.primary
			},
			background: {
				paper: '#FFFFFF',
				default: '#F2F8F1'
			},
			error: {
				light: '#ffcdd2',
				main: '#f44336',
				dark: '#b71c1c'
			}
		}
	},
	light3: {
		palette: {
			mode: 'light',
			divider: '#e2e8f0',
			text: lightPaletteText,
			primary: {
				light: '#D3C0CD',
				main: '#6B2C57',
				dark: '#3C102C',
				contrastText: darkPaletteText.primary
			},
			secondary: {
				light: '#C3C2D2',
				main: '#36336A',
				dark: '#16143C',
				contrastText: darkPaletteText.primary
			},
			background: {
				paper: '#FFFFFF',
				default: '#FAFAFE'
			},
			error: {
				light: '#ffcdd2',
				main: '#f44336',
				dark: '#b71c1c'
			}
		}
	},
	light4: {
		palette: {
			mode: 'light',
			divider: '#e2e8f0',
			text: lightPaletteText,
			primary: {
				light: '#C6C9CD',
				main: '#404B57',
				dark: '#1C232C',
				contrastText: darkPaletteText.primary
			},
			secondary: {
				light: '#C2C8D2',
				main: '#354968',
				dark: '#16213A',
				contrastText: darkPaletteText.primary
			},
			background: {
				paper: '#FFFFFF',
				default: '#F5F4F6'
			},
			error: {
				light: '#ffcdd2',
				main: '#f44336',
				dark: '#b71c1c'
			}
		}
	},
	light5: {
		palette: {
			mode: 'light',
			divider: '#e2e8f0',
			text: lightPaletteText,
			primary: {
				light: '#C4C4C4',
				main: '#3A3A3A',
				dark: '#181818',
				contrastText: darkPaletteText.primary
			},
			secondary: {
				light: '#EFEFED',
				main: '#CBCAC3',
				dark: '#ACABA1',
				contrastText: lightPaletteText.primary
			},
			background: {
				paper: '#EFEEE7',
				default: '#FAF8F2'
			},
			error: {
				light: '#F7EAEA',
				main: '#EBCECE',
				dark: '#E3B9B9'
			}
		}
	},
	dark1: {
		palette: {
			mode: 'dark',
			divider: 'rgba(241,245,249,.12)',
			text: darkPaletteText,
			primary: {
				light: '#C2C2C3',
				main: '#323338',
				dark: '#131417',
				contrastText: darkPaletteText.primary
			},
			secondary: {
				light: '#B8E1D9',
				main: '#129B7F',
				dark: '#056D4F',
				contrastText: darkPaletteText.primary
			},
			background: {
				paper: '#262526',
				default: '#1E1D1E'
			},
			error: {
				light: '#ffcdd2',
				main: '#f44336',
				dark: '#b71c1c'
			}
		}
	},
	dark2: {
		palette: {
			mode: 'dark',
			divider: 'rgba(241,245,249,.12)',
			text: darkPaletteText,
			primary: {
				light: '#C9CACE',
				main: '#4B4F5A',
				dark: '#23262E',
				contrastText: darkPaletteText.primary
			},
			secondary: {
				light: '#F8F5F2',
				main: '#E6DED5',
				dark: '#D5C8BA',
				contrastText: lightPaletteText.primary
			},
			background: {
				paper: '#31343E',
				default: '#2A2D35'
			},
			error: {
				light: '#F7EAEA',
				main: '#EBCECE',
				dark: '#E3B9B9'
			}
		}
	},
	dark3: {
		palette: {
			mode: 'dark',
			divider: 'rgba(241,245,249,.12)',
			text: darkPaletteText,
			primary: {
				light: '#C2C8D2',
				main: '#354968',
				dark: '#16213A',
				contrastText: darkPaletteText.primary
			},
			secondary: {
				light: '#F4CFCA',
				main: '#D55847',
				dark: '#C03325',
				contrastText: darkPaletteText.primary
			},
			background: {
				paper: '#23354E',
				default: '#1B2A3F'
			},
			error: {
				light: '#ffcdd2',
				main: '#f44336',
				dark: '#b71c1c'
			}
		}
	},
	dark4: {
		palette: {
			mode: 'dark',
			divider: 'rgba(241,245,249,.12)',
			text: darkPaletteText,
			primary: {
				light: '#CECADF',
				main: '#5A4E93',
				dark: '#2E2564',
				contrastText: darkPaletteText.primary
			},
			secondary: {
				light: '#B3EBD6',
				main: '#00BC77',
				dark: '#009747',
				contrastText: darkPaletteText.primary
			},
			background: {
				paper: '#22184B',
				default: '#180F3D'
			},
			error: {
				light: '#ffcdd2',
				main: '#f44336',
				dark: '#b71c1c'
			}
		}
	},
	dark5: {
		palette: {
			mode: 'dark',
			divider: 'rgba(241,245,249,.12)',
			text: darkPaletteText,
			primary: {
				light: '#CCD7E2',
				main: '#56789D',
				dark: '#2B486F',
				contrastText: darkPaletteText.primary
			},
			secondary: {
				light: '#D7D3ED',
				main: '#796CC4',
				dark: '#493DA2',
				contrastText: darkPaletteText.primary
			},
			background: {
				paper: '#465261',
				default: '#232931'
			},
			error: {
				light: '#ffcdd2',
				main: '#f44336',
				dark: '#b71c1c'
			}
		}
	},
	dark6: {
		palette: {
			mode: 'dark',
			divider: 'rgba(241,245,249,.12)',
			text: darkPaletteText,
			primary: {
				light: '#BEBFC8',
				main: '#252949',
				dark: '#0D0F21',
				contrastText: darkPaletteText.primary
			},
			secondary: {
				light: '#CBD7FE',
				main: '#5079FC',
				dark: '#2749FA',
				contrastText: darkPaletteText.primary
			},
			background: {
				paper: '#2D3159',
				default: '#202441'
			},
			error: {
				light: '#ffcdd2',
				main: '#f44336',
				dark: '#b71c1c'
			}
		}
	},
	dark7: {
		palette: {
			mode: 'dark',
			divider: 'rgba(241,245,249,.12)',
			text: darkPaletteText,
			primary: {
				light: '#BCC8CD',
				main: '#204657',
				dark: '#0B202C',
				contrastText: darkPaletteText.primary
			},
			secondary: {
				light: '#B3EBC5',
				main: '#00BD3E',
				dark: '#00981B',
				contrastText: darkPaletteText.primary
			},
			background: {
				paper: '#1C1E27',
				default: '#15171E'
			},
			error: {
				light: '#ffcdd2',
				main: '#f44336',
				dark: '#b71c1c'
			}
		}
	},
	greyDark: {
		palette: {
			mode: 'dark',
			divider: 'rgba(241,245,249,.12)',
			text: darkPaletteText,
			primary: {
				light: dark[200],
				main: dark[700],
				dark: dark[800],
				contrastText: darkPaletteText.primary
			},
			secondary: {
				light: skyBlue[100],
				main: skyBlue[500],
				dark: skyBlue[900],
				contrastText: lightPaletteText.primary
			},
			background: {
				paper: blueGrey[700],
				default: blueGrey[900]
			},
			error: {
				light: '#ffcdd2',
				main: '#f44336',
				dark: '#b71c1c'
			}
		}
	}
};

export default themesConfig;
