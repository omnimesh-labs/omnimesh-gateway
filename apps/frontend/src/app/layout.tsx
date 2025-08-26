import clsx from 'clsx';
import 'src/styles/splash-screen.css';
import 'src/styles/index.css';
import { SessionProvider } from 'next-auth/react';
import { auth } from '@auth/authJs';
import generateMetadata from '../utils/generateMetadata';
import App from './App';

// eslint-disable-next-line react-refresh/only-export-components
export const metadata = await generateMetadata({
	title: 'MCP Gateway - Dashboard',
	description: 'MCP Gateway - Production-ready API gateway for Model Context Protocol servers',
	cardImage: '/card.png',
	robots: 'follow, index',
	favicon: '/favicon.ico',
	url: 'https://mcp-gateway.example.com'
});

export default async function RootLayout({
	children
}: Readonly<{
	children: React.ReactNode;
}>) {
	const session = await auth();

	return (
		<html lang="en">
			<head>
				<meta charSet="utf-8" />
				<meta
					name="viewport"
					content="width=device-width, initial-scale=1, shrink-to-fit=no"
				/>
				<meta
					name="theme-color"
					content="#000000"
				/>
				<base href="/" />
				{/*
					manifest.json provides metadata used when your web app is added to the
					homescreen on Android. See https://developers.google.com/web/fundamentals/engage-and-retain/web-app-manifest/
				*/}
				<link
					rel="manifest"
					href="/manifest.json"
				/>
				<link
					rel="shortcut icon"
					href="/favicon.ico"
				/>
				{/* Font and style imports - optimized for performance */}
				<link
					rel="preload"
					as="style"
					href="/assets/fonts/Geist/geist.css"
				/>
				<link
					rel="stylesheet"
					href="/assets/fonts/Geist/geist.css"
				/>
				<link
					rel="preload"
					as="style"
					href="/assets/fonts/material-design-icons/MaterialIconsOutlined.css"
				/>
				<link
					rel="stylesheet"
					href="/assets/fonts/material-design-icons/MaterialIconsOutlined.css"
				/>
				<link
					rel="stylesheet"
					href="/assets/fonts/meteocons/style.css"
				/>
				<link
					rel="stylesheet"
					href="/assets/styles/prism.css"
				/>
				<noscript id="emotion-insertion-point" />
			</head>
			<body
				id="root"
				className={clsx('loading')}
			>
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
