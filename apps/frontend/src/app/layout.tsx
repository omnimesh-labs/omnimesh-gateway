import 'src/styles/splash-screen.css';
import 'src/styles/index.css';
import { SessionProvider } from 'next-auth/react';
import { auth } from '@auth/authJs';
import App from './App';
import StylesheetLoader from '../components/StylesheetLoader';
import type { Metadata, Viewport } from 'next';

// Force dynamic rendering for all pages
export const dynamic = 'force-dynamic';

export const metadata: Metadata = {
	title: 'Omnimesh AI Gateway - Dashboard',
	description: 'Omnimesh AI Gateway - Production-ready API gateway for Model Context Protocol servers',
	referrer: 'origin-when-cross-origin',
	keywords: ['MCP', 'Model Context Protocol', 'API Gateway', 'Omnimesh AI Gateway'],
	authors: [{ name: 'Omnimesh AI Gateway Team' }],
	creator: 'Omnimesh AI Gateway Team',
	publisher: 'Omnimesh AI Gateway Team',
	robots: 'follow, index',
	icons: { icon: '/favicon.ico' },
	manifest: '/manifest.json',
	metadataBase: new URL('https://github.com/omnimesh-labs/omnimesh-gateway'),
	openGraph: {
		url: 'https://github.com/omnimesh-labs/omnimesh-gateway',
		title: 'Omnimesh AI Gateway - Dashboard',
		description: 'Omnimesh AI Gateway - Production-ready API gateway for Model Context Protocol servers',
		images: ['/card.png'],
		type: 'website',
		siteName: 'Omnimesh AI Gateway - Dashboard'
	},
	twitter: {
		card: 'summary_large_image',
		site: '@OmnimeshAI',
		creator: '@OmnimeshAI',
		title: 'Omnimesh AI Gateway - Dashboard',
		description: 'Omnimesh AI Gateway - Production-ready API gateway for Model Context Protocol servers',
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
