'use client';

import NavLinkAdapter from '@fuse/core/NavLinkAdapter';
import { styled } from '@mui/material/styles';
import ListItemText from '@mui/material/ListItemText';
import clsx from 'clsx';
import { memo, useMemo } from 'react';
import { ListItemButton, ListItemButtonProps } from '@mui/material';
import NavBadge from '../../NavBadge';
import SvgIcon from '../../../SvgIcon';
import { NavItemComponentProps } from '../../NavItem';

const Root = styled(ListItemButton)<ListItemButtonProps>(({ theme }) => ({
	color: theme.vars.palette.text.primary,
	textDecoration: 'none!important',
	flex: '0 1 auto',
	minHeight: 48,
	'&.active': {
		backgroundColor: `${theme.vars.palette.secondary.main}!important`,
		color: `${theme.vars.palette.secondary.contrastText}!important`,
		'& .fuse-list-item-text-primary': {
			color: 'inherit'
		},
		'& .fuse-list-item-icon': {
			color: 'inherit'
		}
	},
	'& .fuse-list-item-icon': {},
	'& .fuse-list-item-text': {
		padding: '0 0 0 16px'
	}
}));

type NavHorizontalItemProps = NavItemComponentProps;

/**
 * NavHorizontalItem is a component responsible for rendering the navigation element in the horizontal menu in the  theme.
 */
function NavHorizontalItem(props: NavHorizontalItemProps) {
	const { item, checkPermission } = props;
	const component = item.url ? NavLinkAdapter : 'li';

	const itemProps = useMemo(
		() => ({
			...(component !== 'li' && {
				disabled: item.disabled,
				to: item.url || '',
				end: item.end,
				role: 'button',
				exact: item?.exact
			})
		}),
		[item, component]
	);

	const memoizedContent = useMemo(
		() => (
			<Root
				component={component}
				className={clsx('fuse-list-item', item.active && 'active')}
				sx={item.sx}
				{...itemProps}
			>
				{item.icon && (
					<SvgIcon
						className={clsx('fuse-list-item-icon shrink-0', item.iconClass)}
						color="action"
					>
						{item.icon}
					</SvgIcon>
				)}

				<ListItemText
					className="fuse-list-item-text"
					primary={item.title}
					classes={{ primary: 'text-md fuse-list-item-text-primary truncate' }}
				/>

				{item.badge && (
					<NavBadge
						className="ltr:ml-2 rtl:mr-2"
						badge={item.badge}
					/>
				)}
			</Root>
		),
		[component, item.active, item.badge, item.icon, item.iconClass, item.sx, item.title, itemProps]
	);

	if (checkPermission && !item?.hasPermission) {
		return null;
	}

	return memoizedContent;
}

const NavHorizontalItemWithMemo = memo(NavHorizontalItem);

export default NavHorizontalItemWithMemo;
