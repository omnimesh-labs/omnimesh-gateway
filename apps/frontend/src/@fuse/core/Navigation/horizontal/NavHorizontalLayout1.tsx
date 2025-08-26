import { styled } from '@mui/material/styles';
import clsx from 'clsx';
import NavItem from '../NavItem';
import { NavigationProps } from '../Navigation';
import { Box } from '@mui/material';

const Nav = styled(Box)(({ theme }) => ({
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
		},
		padding: '8px 12px 8px 12px',
		height: 36,
		minHeight: 36,
		'&.level-0': {
			minHeight: 36
		},
		'& .fuse-list-item-text': {
			padding: '0 0 0 8px'
		}
	},
	'&.active-square-list': {
		'& .fuse-list-item': {
			borderRadius: '0'
		}
	}
}));

/**
 * NavHorizontalLayout1 is a react component used for building and
 * rendering horizontal navigation menus, using the Material UI List component.
 */
function NavHorizontalLayout1(props: NavigationProps) {
	const { navigation, active, dense, className, checkPermission } = props;

	return (
		<Nav
			className={clsx(
				'navigation flex p-0 whitespace-nowrap',
				`active-${active}-list`,
				dense && 'dense',
				className
			)}
		>
			{navigation.map((_item) => (
				<NavItem
					key={_item.id}
					type={`horizontal-${_item.type}`}
					item={_item}
					nestedLevel={0}
					dense={dense}
					checkPermission={checkPermission}
				/>
			))}
		</Nav>
	);
}

export default NavHorizontalLayout1;
