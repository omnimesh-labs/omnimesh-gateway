import 'src/styles/splash-screen.css';
import 'src/styles/index.css';
import { SessionProvider } from 'next-auth/react';
import { auth } from '@auth/authJs';
import App from './App';
import StylesheetLoader from '../components/StylesheetLoader';
import type { Metadata, Viewport } from 'next';


export const metadata: Metadata = {
	title: 'Janex - Dashboard',
	description: 'Janex - Production-ready API gateway for Model Context Protocol servers',
	referrer: 'origin-when-cross-origin',
	keywords: ['MCP', 'Model Context Protocol', 'API Gateway', 'Janex'],
	authors: [{ name: 'Janex Team' }],
	creator: 'Janex Team',
	publisher: 'Janex Team',
	robots: 'follow, index',
	icons: { icon: '/favicon.ico' },
	manifest: '/manifest.json',
	metadataBase: new URL('https://janex.example.com'),
	openGraph: {
		url: 'https://janex.example.com',
		title: 'Janex - Dashboard',
		description: 'Janex - Production-ready API gateway for Model Context Protocol servers',
		images: ['/card.png'],
		type: 'website',
		siteName: 'Janex - Dashboard'
	},
	twitter: {
		card: 'summary_large_image',
		site: '@Janex',
		creator: '@Janex',
		title: 'Janex - Dashboard',
		description: 'Janex - Production-ready API gateway for Model Context Protocol servers',
		images: ['/card.png']
	},
	other: {
		'emotion-insertion-point': ''
	}
};


export const viewport: Viewport = {
	width: 'device-width',
	initialScale: 1,
	themeColor: '#000000'
};

export default async function RootLayout({
	children
}: Readonly<{
	children: React.ReactNode;
}>) {
	const session = await auth();

	return (
		<html lang="en">
			<body>
				<StylesheetLoader />
				<SessionProvider
					basePath="/auth"
					session={session}
				>
					<App>{children}</App>
				</SessionProvider>
			</body>
		</html>
	);
}
