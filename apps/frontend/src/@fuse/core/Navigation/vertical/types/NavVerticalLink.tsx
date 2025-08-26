'use client';
import { useMemo } from 'react';
import { NavItemComponentProps } from '../../NavItem';
import NavVerticalItemBase from '../shared/NavVerticalItemBase';

/**
 * NavVerticalLink
 * Create a vertical Link to use inside the navigation component.
 */
function NavVerticalLink(props: NavItemComponentProps) {
	const { item, nestedLevel = 0, onItemClick, checkPermission } = props;

	const memoizedContent = useMemo(
		() => (
			<NavVerticalItemBase
				item={item}
				nestedLevel={nestedLevel}
				onItemClick={onItemClick}
			/>
		),
		[item, nestedLevel, onItemClick]
	);

	if (checkPermission && !item?.hasPermission) {
		return null;
	}

	return memoizedContent;
}

export default NavVerticalLink;
