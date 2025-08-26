import { createContext } from 'react';
import { PartialDeep } from 'type-fest';
import { FlatNavItemType, NavItemType } from '@fuse/core/Navigation/types/NavItemType';

// Define the context type
export type NavigationContextType = {
	navigationItems: FlatNavItemType[];
	appendNavigationItem: (item: NavItemType, parentId?: string | null) => void;
	prependNavigationItem: (item: NavItemType, parentId?: string | null) => void;
	updateNavigationItem: (id: string, item: PartialDeep<NavItemType>) => void;
	removeNavigationItem: (id: string) => void;
	resetNavigation: () => void;
	getNavigationItemById: (id: string) => FlatNavItemType | undefined;
	setNavigation: (items: NavItemType[]) => void;
};

// Create the context
export const NavigationContext = createContext<NavigationContextType | undefined>(undefined);
