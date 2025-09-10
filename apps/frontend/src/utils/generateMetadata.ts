import type { Metadata } from 'next';

async function generateMetadata(meta: {
	title: string;
	description: string;
	cardImage: string;
	robots: string;
	favicon: string;
	url: string;
}): Promise<Metadata> {
	return {
		title: meta.title,
		description: meta.description,
		referrer: 'origin-when-cross-origin',
		keywords: ['MCP', 'Model Context Protocol', 'API Gateway', 'Omnimesh AI Gateway'],
		authors: [{ name: 'Omnimesh AI Gateway Team' }],
		creator: 'Omnimesh AI Gateway Team',
		publisher: 'Omnimesh AI Gateway Team',
		robots: meta.robots,
		icons: { icon: meta.favicon },
		manifest: '/manifest.json',
		metadataBase: new URL(meta.url),
		openGraph: {
			url: meta.url,
			title: meta.title,
			description: meta.description,
			images: [meta.cardImage],
			type: 'website',
			siteName: meta.title
		},
		twitter: {
			card: 'summary_large_image',
			site: '@OmnimeshAI',
			creator: '@OmnimeshAI',
			title: meta.title,
			description: meta.description,
			images: [meta.cardImage]
		},
		other: {
			'emotion-insertion-point': ''
		}
	};
}

export default generateMetadata;
