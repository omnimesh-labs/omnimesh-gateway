'use client';

import { useEffect } from 'react';

/**
 * Component to optimize font loading by progressively loading fonts
 * after the initial page load
 */
export default function FontOptimization() {
	useEffect(() => {
		// Activate non-critical fonts after page load
		const links = document.querySelectorAll<HTMLLinkElement>('link[media="print"]');
		links.forEach((link) => {
			if (link.onload) {
				link.media = 'all';
			}
		});
	}, []);

	return null;
}