'use client';
import { useMemo } from 'react';
import { NavItemComponentProps } from '../../NavItem';
import NavVerticalItemBase from '../shared/NavVerticalItemBase';

/**
 * NavVerticalItem is a React component used to render NavItem as part of the  navigational component.
 */
function NavVerticalItem(props: NavItemComponentProps) {
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

export default NavVerticalItem;
