'use client';
import { useMemo } from 'react';
import NavItem, { NavItemComponentProps } from '../../NavItem';
import NavVerticalItemBase from '../shared/NavVerticalItemBase';

/**
 * NavVerticalGroup is a component used to render a group of navigation items in a vertical layout.
 */
function NavVerticalGroup(props: NavItemComponentProps) {
	const { item, nestedLevel = 0, onItemClick, checkPermission } = props;

	const memoizedContent = useMemo(
		() => (
			<>
				<NavVerticalItemBase
					className={`fuse-list-subheader mt-4 ${!item.url ? 'cursor-default' : ''}`}
					item={item}
					nestedLevel={nestedLevel}
					onItemClick={onItemClick}
					showIcon={false}
					primaryTitleProps={{
						className: 'fuse-list-subheader-text font-medium',
						color: 'secondary'
					}}
					subtitleProps={{
						className: 'fuse-list-subheader-text-secondary font-medium'
					}}
				/>

				{item?.children?.map((_item) => (
					<NavItem
						key={_item.id}
						type={`vertical-${_item.type}`}
						item={_item}
						nestedLevel={nestedLevel}
						onItemClick={onItemClick}
						checkPermission={checkPermission}
					/>
				))}
			</>
		),
		[checkPermission, item, nestedLevel, onItemClick]
	);

	if (checkPermission && !item?.hasPermission) {
		return null;
	}

	return memoizedContent;
}


export default NavVerticalGroup;
