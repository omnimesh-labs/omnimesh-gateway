'use client';
import NavLinkAdapter from '@fuse/core/NavLinkAdapter';
import { styled } from '@mui/material/styles';
import Tooltip from '@mui/material/Tooltip';
import clsx from 'clsx';
import { useMemo } from 'react';
import Box from '@mui/material/Box';
import { ListItemButton } from '@mui/material';
import Typography from '@mui/material/Typography';
import NavBadge from '../../NavBadge';
import SvgIcon from '../../../SvgIcon';
import { NavigationProps } from '../../Navigation';
import { NavItemComponentProps } from '../../NavItem';

const Root = styled(Box)(({ theme }) => ({
	'& > .fuse-list-item': {
		minHeight: 52,
		height: 52,
		width: 52,
		borderRadius: 12,
		margin: '0 0 4px 0',
		cursor: 'pointer',
		textDecoration: 'none!important',
		padding: 0,
		color: (theme) => `rgba(${theme.vars.palette.text.primaryChannel} / 0.7)`,
		'&.type-divider': {
			padding: 0,
			height: 2,
			minHeight: 2,
			margin: '12px 0',
			backgroundColor: theme.vars.palette.divider,
			pointerEvents: 'none'
		},
		'&:hover': {
			color: theme.vars.palette.text.primary
		},
		'&.active': {
			color: theme.vars.palette.text.primary,
			backgroundColor: 'rgba(255, 255, 255, .1)!important',
			transition: 'border-radius .15s cubic-bezier(0.4,0.0,0.2,1)',
			'& .fuse-list-item-text-primary': {
				color: 'inherit'
			},
			'& .fuse-list-item-icon': {
				color: 'inherit'
			},
			...theme.applyStyles('light', {
				backgroundColor: 'rgba(0, 0, 0, .05)!important'
			})
		},
		'& .fuse-list-item-icon': {
			color: 'inherit'
		},
		'& .fuse-list-item-text': {}
	}
}));

export type NavVerticalTabProps = Omit<NavigationProps, 'navigation'> & NavItemComponentProps;

/**
 *  The `NavVerticalTab` component renders vertical navigation item with an adaptable
 *  layout to be used within the `Navigation`. It only supports the `type`s of 'item',
 *  'selection' and 'divider'
 * */
function NavVerticalTab(props: NavVerticalTabProps) {
	const { item, onItemClick, firstLevel, selectedId, checkPermission } = props;
	const component = item.url ? NavLinkAdapter : 'li';

	const itemProps = useMemo(
		() => ({
			...(component !== 'li' && {
				disabled: item.disabled,
				to: item.url,
				end: item.end,
				role: 'button'
			})
		}),
		[item, component]
	);

	const memoizedContent = useMemo(
		() => (
			<Root sx={item.sx}>
				<ListItemButton
					component={component}
					className={clsx(
						`type-${item.type}`,
						selectedId === item.id && 'active',
						'fuse-list-item flex h-8 min-h-0 w-8 flex-col items-center justify-center rounded-lg'
					)}
					onClick={() => onItemClick && onItemClick(item)}
					{...itemProps}
				>
					<Tooltip
						title={item.title || ''}
						placement="right"
					>
						<div className="relative flex items-center justify-center">
							{item.icon ? (
								<SvgIcon
									className={clsx('fuse-list-item-icon', item.iconClass)}
									color="action"
								>
									{item.icon}
								</SvgIcon>
							) : (
								item.title && <Typography className="text-lg font-bold">{item.title[0]}</Typography>
							)}
							{item.badge && (
								<NavBadge
									badge={item.badge}
									className="absolute top-0 h-4 min-w-4 justify-center p-1 ltr:right-0 rtl:left-0"
								/>
							)}
						</div>
					</Tooltip>
				</ListItemButton>
				{!firstLevel &&
					item.children &&
					item.children.map((_item) => (
						<NavVerticalTab
							key={_item.id}
							type={`vertical-${_item.type}`}
							item={_item}
							nestedLevel={0}
							onItemClick={onItemClick}
							selectedId={selectedId}
							checkPermission={checkPermission}
						/>
					))}
			</Root>
		),
		[item, component, selectedId, itemProps, firstLevel, onItemClick, checkPermission]
	);

	if (checkPermission && !item?.hasPermission) {
		return null;
	}

	return memoizedContent;
}

export default NavVerticalTab;
