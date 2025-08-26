'use client';

import { LayoutProps } from '@fuse/core/Layout/Layout';
import Layout from '@fuse/core/Layout';
import { useMemo } from 'react';
import themeLayouts from './theme-layouts/themeLayouts';

type MainLayoutProps = Omit<LayoutProps, 'layouts'> & {
	navbar?: boolean;
	toolbar?: boolean;
	footer?: boolean;
	leftSidePanel?: boolean;
	rightSidePanel?: boolean;
};

function MainLayout(props: MainLayoutProps) {
	const {
		children,
		navbar,
		toolbar,
		footer,
		leftSidePanel,
		rightSidePanel,
		settings = {}, // Default to an empty object if settings is undefined
		...rest
	} = props;

	const mergedSettings = useMemo(() => {
		const shorthandSettings = {
			config: {
				...(navbar !== undefined && { navbar: { display: navbar } }),
				...(toolbar !== undefined && { toolbar: { display: toolbar } }),
				...(footer !== undefined && { footer: { display: footer } }),
				...(leftSidePanel !== undefined && { leftSidePanel: { display: leftSidePanel } }),
				...(rightSidePanel !== undefined && { rightSidePanel: { display: rightSidePanel } })
			}
		};
		return { ...settings, ...shorthandSettings };
	}, [settings, navbar, toolbar, footer, leftSidePanel, rightSidePanel]);

	return (
		<Layout
			{...rest}
			layouts={themeLayouts}
			settings={mergedSettings}
		>
			{children}
		</Layout>
	);
}

export default MainLayout;
