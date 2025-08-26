'use client';

import { useEffect } from 'react';

interface PerformanceMonitorProps {
	children: React.ReactNode;
}

export function PerformanceMonitor({ children }: PerformanceMonitorProps) {
	useEffect(() => {
		if (process.env.NODE_ENV === 'development') {
			// Monitor Core Web Vitals
			// const measurePerformance = () => {
			// 	if ('performance' in window) {
			// 		const navigationEntry = performance.getEntriesByType(
			// 			'navigation'
			// 		)[0] as PerformanceNavigationTiming;
			// 		if (navigationEntry) {
			// 			console.group('🚀 Performance Metrics');
			// 			console.log('📊 Navigation Timing:');
			// 			console.log(
			// 				`  DNS Lookup: ${Math.round(navigationEntry.domainLookupEnd - navigationEntry.domainLookupStart)}ms`
			// 			);
			// 			console.log(
			// 				`  Connection: ${Math.round(navigationEntry.connectEnd - navigationEntry.connectStart)}ms`
			// 			);
			// 			console.log(
			// 				`  Time to First Byte: ${Math.round(navigationEntry.responseStart - navigationEntry.requestStart)}ms`
			// 			);
			// 			console.log(
			// 				`  DOM Interactive: ${Math.round(navigationEntry.domInteractive - navigationEntry.navigationStart)}ms`
			// 			);
			// 			console.log(
			// 				`  DOM Complete: ${Math.round(navigationEntry.domComplete - navigationEntry.navigationStart)}ms`
			// 			);
			// 			console.log(
			// 				`  Load Complete: ${Math.round(navigationEntry.loadEventEnd - navigationEntry.navigationStart)}ms`
			// 			);
			// 			console.groupEnd();
			// 		}
			// 		// Monitor bundle size
			// 		const resources = performance.getEntriesByType('resource') as PerformanceResourceTiming[];
			// 		const jsResources = resources.filter((r) => r.name.includes('.js') && r.name.includes('/_next/'));
			// 		const totalJSSize = jsResources.reduce(
			// 			(total, resource) => total + (resource.transferSize || 0),
			// 			0
			// 		);
			// 		if (totalJSSize > 0) {
			// 			console.group('📦 Bundle Analysis');
			// 			console.log(`  JavaScript Bundle Size: ${Math.round(totalJSSize / 1024)}KB`);
			// 			console.log(`  Number of JS chunks: ${jsResources.length}`);
			// 			console.groupEnd();
			// 		}
			// 	}
			// };
			// Wait for page to fully load
			// if (document.readyState === 'complete') {
			// 	measurePerformance();
			// } else {
			// 	window.addEventListener('load', measurePerformance);
			// 	return () => window.removeEventListener('load', measurePerformance);
			// }
		}
	}, []);

	return <>{children}</>;
}

export default PerformanceMonitor;
