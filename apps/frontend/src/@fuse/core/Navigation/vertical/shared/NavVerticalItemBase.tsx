import NavLinkAdapter from '@fuse/core/NavLinkAdapter';
import clsx from 'clsx';
import { useMemo, ReactNode, useState, useEffect } from 'react';
import { Button, Collapse, IconButton, SxProps, Typography, TypographyProps } from '@mui/material';
import NavBadge from '../../NavBadge';
import SvgIcon from '../../../SvgIcon';
import { NavItemType } from '../../types/NavItemType';
import isUrlInChildren from '../../isUrlInChildren';
import usePathname from '@fuse/hooks/usePathname';
import Link from '@fuse/core/Link';

export type NavVerticalBaseProps = {
	item: NavItemType;
	nestedLevel?: number;
	onItemClick?: (item: NavItemType) => void;
	className?: string;
	component?: React.ElementType;
	children?: ReactNode;
	showIcon?: boolean;
	primaryTitleProps?: TypographyProps;
	subtitleProps?: TypographyProps;
	sx?: SxProps;
	checkPermission?: boolean;
	dense?: boolean;
};

function needsToBeOpened(pathname: string, item: NavItemType) {
	return pathname && isUrlInChildren(item, pathname);
}

/**
 * NavVerticalItemBase is a shared base component for all vertical navigation items
 */
function NavVerticalItemBase(props: NavVerticalBaseProps) {
	const {
		item,
		nestedLevel = 0,
		onItemClick,
		className,
		component,
		showIcon = true,
		primaryTitleProps = {},
		subtitleProps = {},
		sx,
		checkPermission,
		dense = false
	} = props;

	const pathname = usePathname();

	const isGroup = useMemo(() => item.type === 'group', [item.type]);
	const isCollapsable = useMemo(() => item.type === 'collapse', [item.type]);

	const buttonComponent = useMemo(() => {
		if (component) {
			return component;
		}

		if (item.url) {
			if (item.url.startsWith('http') || item.url.startsWith('mailto') || item.url.startsWith('https')) {
				return Link;
			}

			return NavLinkAdapter;
		}

		return 'div';
	}, [component, item.url]);
	const [open, setOpen] = useState(() => needsToBeOpened(pathname, item));

	useEffect(() => {
		if (isCollapsable && item.url) {
			setOpen((prev) => prev || needsToBeOpened(pathname, item));
		}
	}, [pathname, item, isCollapsable]);

	const itemProps = useMemo(
		() => ({
			...(buttonComponent === NavLinkAdapter && {
				disabled: item.disabled,
				to: item.url || '',
				end: item.end,
				role: 'button',
				exact: item?.exact
			}),
			...(buttonComponent === Link && {
				disabled: item.disabled,
				to: item.url,
				role: 'button',
				target: item.target ? item.target : '_blank',
				exact: item?.exact
			})
		}),
		[item, buttonComponent]
	);

	function handleClick(e: React.MouseEvent<HTMLButtonElement>) {
		e.stopPropagation();

		if (isCollapsable) {
			setOpen((prev) => prev || !prev);
		}

		if (onItemClick) {
			onItemClick(item);
		}
	}

	if (checkPermission && !item?.hasPermission) {
		return null;
	}

	return (
		<>
			<Button
				component={buttonComponent}
				className={clsx(
					'mb-1 flex w-full items-center justify-start gap-2 rounded-md p-2',
					item.active && 'active',
					item.subtitle && 'items-start',
					className,
					isGroup && 'my-4',
					dense && 'p-1'
				)}
				sx={{
					...sx,
					'&.active': {
						backgroundColor: (theme) =>
							theme.palette.mode === 'light'
								? 'rgba(0, 0, 0, .05)!important'
								: 'rgba(255, 255, 255, .1)!important'
					}
				}}
				color="inherit"
				onClick={handleClick}
				aria-label={item.title}
				aria-expanded={open}
				{...itemProps}
			>
				{!isGroup && showIcon && item.icon && (
					<SvgIcon className={clsx('fuse-list-item-icon', 'shrink-0', item.iconClass)}>
						{item.icon}
					</SvgIcon>
				)}
				<div className="flex min-w-0 flex-auto flex-col items-start gap-1">
					{item.title && (
						<Typography
							{...{
								...primaryTitleProps,
								className: clsx(
									'fuse-list-item-title',
									'leading-none truncate max-w-full',
									isGroup && 'font-medium',
									primaryTitleProps?.className
								),
								...(isGroup && {
									color: 'secondary'
								})
							}}
						>
							{item.title}
						</Typography>
					)}
					{item.subtitle && (
						<Typography
							{...{
								color: 'text.secondary',
								...subtitleProps,
								className: clsx(
									'fuse-list-item-subtitle',
									'text-sm leading-none truncate max-w-full',
									subtitleProps?.className
								)
							}}
						>
							{item.subtitle}
						</Typography>
					)}
				</div>
				{item.badge && <NavBadge badge={item.badge} />}

				{isCollapsable && (
					<IconButton
						disableRipple
						className="h-5 w-5 p-0 hover:bg-transparent focus:bg-transparent"
						onClick={(ev) => {
							ev.preventDefault();
							ev.stopPropagation();
							setOpen(!open);
						}}
					>
						<SvgIcon
							size={13}
							className="arrow-icon"
							color="inherit"
						>
							{open ? 'lucide:chevron-down' : 'lucide:chevron-right'}
						</SvgIcon>
					</IconButton>
				)}
			</Button>

			{isCollapsable && (
				<Collapse
					in={open}
					className="collapse-children"
					sx={{
						...(nestedLevel < 4 && {
							paddingLeft: '16px',
							'& > .MuiCollapse-wrapper': {
								borderLeft: '1px solid',
								borderColor: 'divider',
								'& > .MuiCollapse-wrapperInner': {
									paddingLeft: '8px'
								}
							}
						})
					}}
				>
					{item?.children?.map((_item) => (
						<NavVerticalItemBase
							key={_item.id}
							item={_item}
							nestedLevel={isGroup ? nestedLevel : nestedLevel + 1}
							onItemClick={onItemClick}
							checkPermission={checkPermission}
						/>
					))}
				</Collapse>
			)}

			{isGroup &&
				item?.children?.map((_item) => (
					<NavVerticalItemBase
						key={_item.id}
						item={_item}
						nestedLevel={isGroup ? nestedLevel : nestedLevel + 1}
						onItemClick={onItemClick}
						checkPermission={checkPermission}
					/>
				))}
		</>
	);
}

export default NavVerticalItemBase;
