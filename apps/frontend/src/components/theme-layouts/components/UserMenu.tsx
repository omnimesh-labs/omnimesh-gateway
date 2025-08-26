import Avatar from '@mui/material/Avatar';
import Button from '@mui/material/Button';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import MenuItem from '@mui/material/MenuItem';
import Typography from '@mui/material/Typography';
import { useState } from 'react';
import Link from '@fuse/core/Link';
import SvgIcon from '@fuse/core/SvgIcon';
import { darken } from '@mui/material/styles';
import Tooltip from '@mui/material/Tooltip';
import clsx from 'clsx';
import Popover, { PopoverProps } from '@mui/material/Popover';
import useUser from '@auth/useUser';

type UserMenuProps = {
	className?: string;
	popoverProps?: Partial<PopoverProps>;
	arrowIcon?: string;
	dense?: boolean;
	onlyAvatar?: boolean;
};

/**
 * The user menu.
 */
function UserMenu(props: UserMenuProps) {
	const { className, popoverProps, arrowIcon = 'lucide:chevron-up', dense = false, onlyAvatar = false } = props;
	const { data: user, signOut, isGuest } = useUser();
	const [userMenu, setUserMenu] = useState<HTMLElement | null>(null);
	const userMenuClick = (event: React.MouseEvent<HTMLElement>) => {
		setUserMenu(event.currentTarget);
	};

	const userMenuClose = () => {
		setUserMenu(null);
	};

	if (!user) {
		return null;
	}

	return (
		<>
			<Button
				className={clsx(
					'user-menu flex shrink-0 justify-start',
					onlyAvatar ? 'min-w-0 p-0' : dense ? 'h-9 min-h-9 gap-1.5 px-1' : 'h-14 min-h-14 gap-3',
					className
				)}
				onClick={userMenuClick}
				color="inherit"
			>
				{user?.photoURL ? (
					<Avatar
						sx={(theme) => ({
							background: theme.vars.palette.background.default,
							color: theme.vars.palette.text.secondary
						})}
						className={clsx('avatar rounded-lg', dense ? 'h-7 w-7' : 'h-10 w-10')}
						alt="user photo"
						src={user?.photoURL}
						variant="rounded"
					/>
				) : (
					<Avatar
						sx={(theme) => ({
							background: (theme) => darken(theme.palette.background.default, 0.05),
							color: theme.vars.palette.text.secondary
						})}
						className={clsx('avatar h-10 w-10', dense && 'h-8 w-8')}
					>
						{user?.displayName?.[0]}
					</Avatar>
				)}
				{!onlyAvatar && (
					<>
						<div className={clsx('flex flex-auto flex-col', dense ? '' : 'gap-2')}>
							<Typography
								component="span"
								className={clsx(
									'title flex truncate leading-none font-semibold tracking-tight capitalize',
									dense ? 'text-md' : 'text-base'
								)}
							>
								{user?.displayName}
							</Typography>
							<Typography
								className={clsx(
									'flex leading-none font-medium tracking-tighter',
									dense ? 'text-sm' : 'text-md'
								)}
								color="text.secondary"
							>
								{user?.email}
							</Typography>
						</div>
						<div className="flex shrink-0 items-center gap-2">
							<Tooltip
								title={
									<>
										{user.role?.toString()}
										{(!user.role || (Array.isArray(user.role) && user.role.length === 0)) &&
											'Guest'}
									</>
								}
							>
								<SvgIcon className="info-icon">lucide:info</SvgIcon>
							</Tooltip>
							<SvgIcon
								className="arrow"
								size={13}
							>
								{arrowIcon}
							</SvgIcon>
						</div>
					</>
				)}
			</Button>
			<Popover
				open={Boolean(userMenu)}
				anchorEl={userMenu}
				onClose={userMenuClose}
				anchorOrigin={{
					vertical: 'top',
					horizontal: 'right'
				}}
				transformOrigin={{
					vertical: 'top',
					horizontal: 'left'
				}}
				classes={{
					paper: 'min-w-32'
				}}
				{...popoverProps}
			>
				{isGuest ? (
					<>
						<MenuItem
							component={Link}
							to="/sign-in"
							role="button"
						>
							<ListItemIcon>
								<SvgIcon>lucide:lock</SvgIcon>
							</ListItemIcon>
							<ListItemText primary="Sign In" />
						</MenuItem>
						<MenuItem
							component={Link}
							to="/sign-up"
							role="button"
						>
							<ListItemIcon>
								<SvgIcon>lucide:user-plus</SvgIcon>
							</ListItemIcon>
							<ListItemText primary="Sign up" />
						</MenuItem>
					</>
				) : (
					<>
						<MenuItem
							component={Link}
							to="/apps/profile"
							onClick={userMenuClose}
							role="button"
						>
							<ListItemIcon>
								<SvgIcon>lucide:circle-user</SvgIcon>
							</ListItemIcon>
							<ListItemText primary="My Profile" />
						</MenuItem>
						<MenuItem
							component={Link}
							to="/apps/mailbox"
							onClick={userMenuClose}
							role="button"
						>
							<ListItemIcon>
								<SvgIcon>lucide:mail</SvgIcon>
							</ListItemIcon>
							<ListItemText primary="Inbox" />
						</MenuItem>
						<MenuItem
							onClick={() => {
								signOut();
							}}
						>
							<ListItemIcon>
								<SvgIcon>lucide:square-arrow-right</SvgIcon>
							</ListItemIcon>
							<ListItemText primary="Sign out" />
						</MenuItem>
					</>
				)}
			</Popover>
		</>
	);
}

export default UserMenu;
