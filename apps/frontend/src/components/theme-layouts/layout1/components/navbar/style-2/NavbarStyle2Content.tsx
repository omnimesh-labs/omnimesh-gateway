import Scrollbars from '@fuse/core/Scrollbars';
import { styled } from '@mui/material/styles';
import ClickAwayListener from '@mui/material/ClickAwayListener';
import clsx from 'clsx';
import { memo, useEffect, useState } from 'react';
import Navigation from '@fuse/core/Navigation';
import useThemeMediaQuery from '@fuse/hooks/useThemeMediaQuery';
import isUrlInChildren from '@fuse/core/Navigation/isUrlInChildren';
import { Theme } from '@mui/system';
import { NavItemType } from '@fuse/core/Navigation/types/NavItemType';
import UserMenu from '@/@auth/components/UserMenu';
import usePathname from '@fuse/hooks/usePathname';
import useNavigationItems from '@/components/theme-layouts/components/navigation/hooks/useNavigationItems';
import { useNavbarContext } from '@/components/theme-layouts/components/navbar/contexts/NavbarContext/useNavbarContext';

const Root = styled('div')(({ theme }) => ({
	backgroundColor: theme.vars.palette.background.default,
	color: theme.vars.palette.text.primary
}));

type StyledPanelProps = {
	theme?: Theme;
	opened?: boolean;
};

const StyledPanel = styled(Scrollbars)<StyledPanelProps>(({ theme }) => ({
	backgroundColor: theme.vars.palette.background.default,
	color: theme.vars.palette.text.primary,
	transition: theme.transitions.create(['opacity'], {
		easing: theme.transitions.easing.sharp,
		duration: theme.transitions.duration.shortest
	}),
	opacity: 0,
	pointerEvents: 'none',
	minHeight: 0,
	variants: [
		{
			props: ({ opened }) => opened,
			style: {
				opacity: 1,
				pointerEvents: 'initial'
			}
		}
	]
}));

/**
 * Check if the item needs to be opened.
 */
function needsToBeOpened(pathname: string, item: NavItemType) {
	return pathname && isUrlInChildren(item, pathname);
}

type NavbarStyle2ContentProps = { className?: string };

/**
 * The navbar style 3 content.
 */
function NavbarStyle2Content(props: NavbarStyle2ContentProps) {
	const { className = '' } = props;
	const isMobile = useThemeMediaQuery((theme) => theme.breakpoints.down('lg'));
	const { data: navigation } = useNavigationItems();
	const { closeMobileNavbar } = useNavbarContext();
	const [selectedNavigation, setSelectedNavigation] = useState<NavItemType[]>([]);
	const [panelOpen, setPanelOpen] = useState(false);
	const pathname = usePathname();

	useEffect(() => {
		navigation?.forEach((item) => {
			if (needsToBeOpened(pathname, item)) {
				setSelectedNavigation([item]);
			}
		});
	}, [navigation, pathname]);

	function handleParentItemClick(selected: NavItemType) {
		/** if there is no child item do not set/open panel
		 */
		if (!selected.children) {
			setSelectedNavigation([]);
			setPanelOpen(false);
			return;
		}

		/**
		 * If navigation already selected toggle panel visibility
		 */
		if (selectedNavigation[0]?.id === selected.id) {
			setPanelOpen(!panelOpen);
		} else {
			/**
			 * Set navigation and open panel
			 */
			setSelectedNavigation([selected]);
			setPanelOpen(true);
		}
	}

	function handleChildItemClick() {
		setPanelOpen(false);

		if (isMobile) {
			closeMobileNavbar();
		}
	}

	return (
		<ClickAwayListener onClickAway={() => setPanelOpen(false)}>
			<Root className={clsx('flex h-full flex-auto', className)}>
				<div
					id="fuse-navbar-side-panel"
					className="flex h-full shrink-0 flex-col items-center"
				>
					<img
						className="my-4 w-6"
						src="/assets/images/logo/logo.svg"
						alt="logo"
					/>

					<Scrollbars
						className="flex min-h-0 w-full flex-1 flex-col justify-start overflow-y-auto overflow-x-hidden"
						option={{
							suppressScrollX: true,
							wheelPropagation: false
						}}
					>
						<Navigation
							className={clsx('navigation min-h-full shrink-0')}
							navigation={navigation}
							layout="vertical-2"
							onItemClick={handleParentItemClick}
							firstLevel
							selectedId={selectedNavigation[0]?.id}
						/>
					</Scrollbars>

					<div className="flex w-full shrink-0 items-center justify-center py-2">
						<UserMenu onlyAvatar />
					</div>
				</div>

				{selectedNavigation.length > 0 && (
					<StyledPanel
						id="fuse-navbar-panel"
						opened={panelOpen}
						className={clsx('overflow-y-auto overflow-x-hidden shadow-sm')}
						option={{ suppressScrollX: true, wheelPropagation: false }}
					>
						<Navigation
							className={clsx('navigation')}
							navigation={selectedNavigation}
							layout="vertical"
							onItemClick={handleChildItemClick}
						/>
					</StyledPanel>
				)}
			</Root>
		</ClickAwayListener>
	);
}

export default memo(NavbarStyle2Content);
