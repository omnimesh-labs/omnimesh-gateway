'use client';

import { useEffect, useState } from 'react';

const stylesheets = [
	'/assets/fonts/Geist/geist.css',
	'/assets/fonts/material-design-icons/MaterialIconsOutlined.css',
	'/assets/styles/prism.css'
];

export default function StylesheetLoader() {
	const [isClient, setIsClient] = useState(false);

	useEffect(() => {
		setIsClient(true);
	}, []);

	useEffect(() => {
		if (!isClient) return;

		// Load external stylesheets
		stylesheets.forEach((href) => {
			// Check if stylesheet is already loaded
			if (document.querySelector(`link[href="${href}"]`)) return;
			
			const link = document.createElement('link');
			link.rel = 'stylesheet';
			link.href = href;
			document.head.appendChild(link);
		});

		// Create emotion insertion point only if it doesn't exist
		if (!document.getElementById('emotion-insertion-point')) {
			const emotionInsertionPoint = document.createElement('noscript');
			emotionInsertionPoint.id = 'emotion-insertion-point';
			document.head.appendChild(emotionInsertionPoint);
		}
	}, [isClient]);

	return null;
}