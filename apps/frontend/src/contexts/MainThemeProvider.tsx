import * as React from 'react';
import Theme from '@fuse/core/Theme';
import { useMainTheme } from '@fuse/core/Settings/hooks/themeHooks';

type MainThemeProviderProps = {
	children: React.ReactNode;
};

function MainThemeProvider({ children }: MainThemeProviderProps) {
	const mainTheme = useMainTheme();

	return (
		<Theme
			theme={mainTheme}
			root
		>
			{children}
		</Theme>
	);
}

export default MainThemeProvider;
