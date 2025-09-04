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
	title: 'MCP Gateway - Dashboard',
	description: 'MCP Gateway - Production-ready API gateway for Model Context Protocol servers',
	referrer: 'origin-when-cross-origin',
	keywords: ['MCP', 'Model Context Protocol', 'API Gateway', 'MCP Gateway'],
	authors: [{ name: 'MCP Gateway Team' }],
	creator: 'MCP Gateway Team',
	publisher: 'MCP Gateway Team',
	robots: 'follow, index',
	icons: { icon: '/favicon.ico' },
	manifest: '/manifest.json',
	metadataBase: new URL('https://github.com/theognis1002/mcp-gateway'),
	openGraph: {
		url: 'https://github.com/theognis1002/mcp-gateway',
		title: 'MCP Gateway - Dashboard',
		description: 'MCP Gateway - Production-ready API gateway for Model Context Protocol servers',
		images: ['/card.png'],
		type: 'website',
		siteName: 'MCP Gateway - Dashboard'
	},
	twitter: {
		card: 'summary_large_image',
		site: '@MCPGateway',
		creator: '@MCPGateway',
		title: 'MCP Gateway - Dashboard',
		description: 'MCP Gateway - Production-ready API gateway for Model Context Protocol servers',
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
