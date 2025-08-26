'use client';

import _ from 'lodash';
import React, { useEffect, useMemo } from 'react';
import { SettingsConfigType } from '@fuse/core/Settings/Settings';
import { themeLayoutsType } from 'src/components/theme-layouts/themeLayouts';
import usePathname from '@fuse/hooks/usePathname';
import useSettings from '@fuse/core/Settings/hooks/useSettings';
import LayoutSettingsContext from './LayoutSettingsContext';

export type RouteObjectType = {
	settings?: SettingsConfigType;
	auth?: string[] | [] | null | undefined;
};

export type LayoutProps = {
	layouts: themeLayoutsType;
	children?: React.ReactNode;
	settings?: SettingsConfigType['layout'];
};

/**
 * Layout
 * React frontend component in a React project that is used for layouting the user interface. The component
 * handles generating user interface settings related to current routes, merged with default settings, and uses
 * the new settings to generate layouts.
 */
function Layout(props: LayoutProps) {
	const { layouts, children, settings: forcedSettings } = props;

	const { data: current } = useSettings();
	const currentLayoutSetting = useMemo(() => current.layout, [current]);
	const pathname = usePathname();

	const layoutSetting = useMemo(
		() => _.merge({}, currentLayoutSetting, forcedSettings),
		[currentLayoutSetting, forcedSettings]
	);

	const layoutStyle = useMemo(() => layoutSetting.style, [layoutSetting]);

	useEffect(() => {
		window.scrollTo(0, 0);
	}, [pathname]);

	return (
		<LayoutSettingsContext value={layoutSetting}>
			{useMemo(() => {
				return Object.entries(layouts).map(([key, Layout]) => {
					if (key === layoutStyle) {
						return (
							<React.Fragment key={key}>
								<Layout>{children}</Layout>
							</React.Fragment>
						);
					}

					return null;
				});
			}, [layoutStyle, layouts, children])}
		</LayoutSettingsContext>
	);
}

export default Layout;
