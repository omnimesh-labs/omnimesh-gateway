'use client';

import NavLinkAdapter from '@fuse/core/NavLinkAdapter';
import { styled, useTheme } from '@mui/material/styles';
import { useDebounce } from '@fuse/hooks';
import Grow from '@mui/material/Grow';
import IconButton from '@mui/material/IconButton';
import ListItemText from '@mui/material/ListItemText';
import Paper from '@mui/material/Paper';
import clsx from 'clsx';
import { memo, useMemo, useState, useEffect } from 'react';
import * as ReactDOM from 'react-dom';
import { Manager, Popper, Reference } from 'react-popper';
import { ListItemButton, ListItemButtonProps } from '@mui/material';
import { Location } from 'history';
import isUrlInChildren from '@fuse/core/Navigation/isUrlInChildren';
import usePathname from '@fuse/hooks/usePathname';
import NavBadge from '../../NavBadge';
import NavItem, { NavItemComponentProps } from '../../NavItem';
import SvgIcon from '../../../SvgIcon';

const Root = styled(ListItemButton)<ListItemButtonProps>(({ theme }) => ({
	color: theme.vars.palette.text.primary,
	minHeight: 48,
	cursor: 'pointer',
	'&.active, &.active:hover, &.active:focus': {
		backgroundColor: `${theme.vars.palette.secondary.main}!important`,
		color: `${theme.vars.palette.secondary.contrastText}!important`,
		'&.open': {
			backgroundColor: 'rgba(0,0,0,.08)'
		},
		'& > .fuse-list-item-text': {
			padding: '0 0 0 16px'
		},
		'& .fuse-list-item-icon': {
			color: 'inherit'
		}
	}
}));

type NavHorizontalCollapseProps = NavItemComponentProps & {
	location: Location;
};

/**
 * NavHorizontalCollapse component helps rendering Horizontal  Navigation Item with children
 * Used in NavVerticalItems and NavHorizontalItems
 */
function NavHorizontalCollapse(props: NavHorizontalCollapseProps) {
	const { item, nestedLevel, dense, checkPermission } = props;
	const [opened, setOpened] = useState(false);
	const [isMounted, setIsMounted] = useState(false);
	const pathname = usePathname();
	const theme = useTheme();
	const component = item.url ? NavLinkAdapter : 'li';

	const itemProps = useMemo(
		() => ({
			...(component !== 'li' && {
				disabled: item.disabled,
				to: item.url,
				end: item.end,
				role: 'button',
				exact: item?.exact
			})
		}),
		[item, component]
	);

	const handleToggle = useDebounce((open: boolean) => {
		setOpened(open);
	}, 150);

	useEffect(() => {
		setIsMounted(true);
	}, []);

	const memoizedContent = useMemo(
		() => (
			<ul className="relative px-0">
				<Manager>
					<Reference>
						{({ ref }) => (
							<div ref={ref}>
								<Root
									component={component}
									className={clsx(
										'fuse-list-item',
										opened && 'open',
										isUrlInChildren(item, pathname) && 'active'
									)}
									onMouseEnter={() => handleToggle(true)}
									onMouseLeave={() => handleToggle(false)}
									aria-owns={opened ? 'menu-fuse-list-grow' : null}
									aria-haspopup="true"
									sx={item.sx}
									{...itemProps}
								>
									{item.icon && (
										<SvgIcon
											color="action"
											className={clsx('fuse-list-item-icon shrink-0', item.iconClass)}
										>
											{item.icon}
										</SvgIcon>
									)}

									<ListItemText
										className="fuse-list-item-text"
										primary={item.title}
										classes={{ primary: 'text-md truncate' }}
									/>

									{item.badge && (
										<NavBadge
											className="mx-1"
											badge={item.badge}
										/>
									)}
									<IconButton
										disableRipple
										className="h-3 w-3 p-0 ltr:ml-1 rtl:mr-1"
										color="inherit"
									>
										<SvgIcon
											size={12}
											className="arrow-icon"
										>
											{theme.direction === 'ltr' ? 'lucide:chevron-right' : 'lucide:chevron-left'}
										</SvgIcon>
									</IconButton>
								</Root>
							</div>
						)}
					</Reference>
					{isMounted &&
						ReactDOM.createPortal(
							<Popper placement={theme.direction === 'ltr' ? 'right' : 'left'}>
								{({ ref, style, placement }) =>
									opened && (
										<div
											ref={ref}
											style={{
												...style,
												zIndex: 999 + nestedLevel + 1
											}}
											data-placement={placement}
											className={clsx('z-999', !opened && 'pointer-events-none')}
										>
											<Grow
												in={opened}
												id="menu-fuse-list-grow"
												style={{ transformOrigin: '0 0 0' }}
											>
												<Paper
													className="rounded-sm"
													onMouseEnter={() => handleToggle(true)}
													onMouseLeave={() => handleToggle(false)}
												>
													{item.children && (
														<ul
															className={clsx(
																'popper-navigation-list',
																dense && 'dense',
																'px-0'
															)}
														>
															{item.children.map((_item) => (
																<NavItem
																	key={_item.id}
																	type={`horizontal-${_item.type}`}
																	item={_item}
																	nestedLevel={nestedLevel + 1}
																	dense={dense}
																/>
															))}
														</ul>
													)}
												</Paper>
											</Grow>
										</div>
									)
								}
							</Popper>,
							document.querySelector('#root')
						)}
				</Manager>
			</ul>
		),
		[component, dense, handleToggle, isMounted, item, itemProps, nestedLevel, opened, pathname, theme.direction]
	);

	if (checkPermission && !item?.hasPermission) {
		return null;
	}

	return memoizedContent;
}

const NavHorizontalCollapseWithMemo = memo(NavHorizontalCollapse);

export default NavHorizontalCollapseWithMemo;
