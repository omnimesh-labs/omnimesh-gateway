'use client';
import Collapse from '@mui/material/Collapse';
import clsx from 'clsx';
import { useMemo, useState } from 'react';
import isUrlInChildren from '@fuse/core/Navigation/isUrlInChildren';
import usePathname from '@fuse/hooks/usePathname';
import NavItem, { NavItemComponentProps } from '../../NavItem';
import NavVerticalItemBase from '../shared/NavVerticalItemBase';
import { NavItemType } from '../../types/NavItemType';

function needsToBeOpened(pathname: string, item: NavItemType) {
	return pathname && isUrlInChildren(item, pathname);
}

/**
 * NavVerticalCollapse component used for vertical navigation items with collapsible children.
 */
function NavVerticalCollapse(props: NavItemComponentProps) {
	const pathname = usePathname();
	const { item, nestedLevel = 0, onItemClick, checkPermission } = props;
	const [open, setOpen] = useState(() => needsToBeOpened(pathname, item));

	const memoizedContent = useMemo(
		() => (
			<>
				<NavVerticalItemBase
					item={item}
					nestedLevel={nestedLevel}
					onItemClick={() => setOpen(!open)}
					className={clsx('fuse-list-item', open && 'open')}
				/>

				{item.children && (
					<Collapse
						in={open}
						className="collapse-children"
					>
						{item.children.map((_item) => (
							<NavItem
								key={_item.id}
								type={`vertical-${_item.type}`}
								item={_item}
								nestedLevel={nestedLevel + 1}
								onItemClick={onItemClick}
								checkPermission={checkPermission}
							/>
						))}
					</Collapse>
				)}
			</>
		),
		[checkPermission, item, nestedLevel, onItemClick, open]
	);

	if (checkPermission && !item?.hasPermission) {
		return null;
	}

	return memoizedContent;
}

export default NavVerticalCollapse;
