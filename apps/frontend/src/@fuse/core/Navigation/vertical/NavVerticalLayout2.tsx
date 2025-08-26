import List from '@mui/material/List';
import { styled } from '@mui/material/styles';
import clsx from 'clsx';
import NavVerticalTab from './types/NavVerticalTab';
import { NavigationProps } from '../Navigation';
import { NavItemType } from '../types/NavItemType';

const StyledList = styled(List)(({ theme }) => ({
	'& .fuse-list-item': {
		'&:hover': {
			backgroundColor: 'rgba(0,0,0,.04)',
			...theme.applyStyles('dark', {
				backgroundColor: 'rgba(255, 255, 255, 0.05)'
			})
		},
		'&:focus:not(.active)': {
			backgroundColor: 'rgba(0,0,0,.05)',
			...theme.applyStyles('dark', {
				backgroundColor: 'rgba(255, 255, 255, 0.06)'
			})
		}
	},
	'& .fuse-list-item-text-primary': {
		lineHeight: '1'
	},
	'&.active-square-list': {
		'& .fuse-list-item, & .active.fuse-list-item': {
			width: '100%',
			borderRadius: '0'
		}
	},
	'&.dense': {}
}));

/**
 * NavVerticalLayout2 component represents a vertical navigation layout with material UI elements.
 * It displays the navigation object in the structured vertical menu and allows to handle onClick events for each navigation item.
 */
function NavVerticalLayout2(props: NavigationProps) {
	const { navigation, active, dense, className, onItemClick, firstLevel, selectedId, checkPermission } = props;

	function handleItemClick(item: NavItemType) {
		onItemClick?.(item);
	}

	return (
		<StyledList
			className={clsx(
				'navigation flex flex-col items-center gap-1 whitespace-nowrap',
				`active-${active}-list`,
				dense && 'dense',
				className
			)}
		>
			{navigation.map((_item) => (
				<NavVerticalTab
					key={_item.id}
					type={`vertical-${_item.type}`}
					item={_item}
					nestedLevel={0}
					onItemClick={handleItemClick}
					firstLevel={firstLevel}
					dense={dense}
					selectedId={selectedId}
					checkPermission={checkPermission}
				/>
			))}
		</StyledList>
	);
}

export default NavVerticalLayout2;
