import clsx from 'clsx';
import { NavigationProps } from '../Navigation';
import { NavItemType } from '../types/NavItemType';
import NavVerticalItemBase from './shared/NavVerticalItemBase';

/**
 * NavVerticalLayout1
 * This component is used to render vertical navigations using
 * the Material-UI List component. It accepts the NavigationProps props
 * and renders the NavItem components accordingly
 */
function NavVerticalLayout1(props: NavigationProps) {
	const { navigation, active, dense, className, onItemClick, checkPermission } = props;

	function handleItemClick(item: NavItemType) {
		onItemClick?.(item);
	}

	return (
		<div className={clsx('navigation px-3', `active-${active}-list`, dense && 'dense', className)}>
			{navigation.map((_item) => (
				<NavVerticalItemBase
					key={_item.id}
					item={_item}
					nestedLevel={0}
					onItemClick={handleItemClick}
					checkPermission={checkPermission}
					dense={dense}
				/>
			))}
		</div>
	);
}

export default NavVerticalLayout1;
