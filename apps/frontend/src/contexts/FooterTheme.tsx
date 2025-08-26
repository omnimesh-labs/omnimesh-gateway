import { ThemeProvider } from '@mui/material/styles';
import { useFooterTheme } from '@fuse/core/Settings/hooks/themeHooks';

type FooterThemeProps = {
	children: React.ReactNode;
};

function FooterTheme({ children }: FooterThemeProps) {
	const footerTheme = useFooterTheme();

	return <ThemeProvider theme={footerTheme}>{children}</ThemeProvider>;
}

export default FooterTheme;
