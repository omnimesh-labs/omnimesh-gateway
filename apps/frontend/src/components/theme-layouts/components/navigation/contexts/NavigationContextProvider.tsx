// Create the provider component
import { ReactNode, useCallback, useState } from 'react';
import { FlatNavItemType, NavItemType } from '@fuse/core/Navigation/types/NavItemType';
import NavigationHelper from '@fuse/utils/NavigationHelper';
import navigationConfig from '@/configs/navigationConfig';
import NavItemModel from '@fuse/core/Navigation/models/NavItemModel';
import { PartialDeep } from 'type-fest';
import { NavigationContext } from '@/components/theme-layouts/components/navigation/contexts/NavigationContext';

export function NavigationContextProvider({ children }: { children: ReactNode }) {
	const [navigationItems, setNavigationItems] = useState<FlatNavItemType[]>(
		NavigationHelper.flattenNavigation(navigationConfig)
	);

	const setNavigation = useCallback((items: NavItemType[]) => {
		setNavigationItems(NavigationHelper.flattenNavigation(items));
	}, []);

	const appendNavigationItem = useCallback(
		(item: NavItemType, parentId?: string | null) => {
			const navigation = NavigationHelper.unflattenNavigation(navigationItems);
			setNavigation(NavigationHelper.appendNavItem(navigation, NavItemModel(item), parentId));
		},
		[navigationItems, setNavigation]
	);

	const prependNavigationItem = useCallback(
		(item: NavItemType, parentId?: string | null) => {
			const navigation = NavigationHelper.unflattenNavigation(navigationItems);
			setNavigation(NavigationHelper.prependNavItem(navigation, NavItemModel(item), parentId));
		},
		[navigationItems, setNavigation]
	);

	const updateNavigationItem = useCallback(
		(id: string, item: PartialDeep<NavItemType>) => {
			const navigation = NavigationHelper.unflattenNavigation(navigationItems);
			setNavigation(NavigationHelper.updateNavItem(navigation, id, item));
		},
		[navigationItems, setNavigation]
	);

	const removeNavigationItem = useCallback(
		(id: string) => {
			const navigation = NavigationHelper.unflattenNavigation(navigationItems);
			setNavigation(NavigationHelper.removeNavItem(navigation, id));
		},
		[navigationItems, setNavigation]
	);

	const resetNavigation = useCallback(() => {
		setNavigationItems(NavigationHelper.flattenNavigation(navigationConfig));
	}, []);

	const getNavigationItemById = useCallback(
		(id: string) => navigationItems.find((item) => item.id === id),
		[navigationItems]
	);

	const value = {
		setNavigation,
		navigationItems,
		appendNavigationItem,
		prependNavigationItem,
		updateNavigationItem,
		removeNavigationItem,
		resetNavigation,
		getNavigationItemById
	};

	return <NavigationContext.Provider value={value}>{children}</NavigationContext.Provider>;
}
