import Drawer from '@mui/material/Drawer';
import SwipeableDrawer from '@mui/material/SwipeableDrawer';
import clsx from 'clsx';
import { useCallback, useEffect, useImperativeHandle, useState, ReactNode } from 'react';
import { SwipeableDrawerProps } from '@mui/material/SwipeableDrawer';
import PageCardedSidebarContent from './PageCardedSidebarContent';
import useThemeMediaQuery from '../../hooks/useThemeMediaQuery';

/**
 * Props for the PageCardedSidebar component.
 */
export type PageCardedSidebarProps = {
	open?: boolean;
	position?: SwipeableDrawerProps['anchor'];
	variant?: SwipeableDrawerProps['variant'];
	onClose?: () => void;
	children?: ReactNode;
	content?: ReactNode;
	ref?: React.RefObject<{ toggleSidebar: (T: boolean) => void }>;
	width?: number;
};

/**
 * The PageCardedSidebar component is a sidebar for the PageCarded component.
 */
function PageCardedSidebar(props: PageCardedSidebarProps) {
	const { open = true, position, variant, onClose = () => {}, ref, width = 240, children, content } = props;

	const isMobile = useThemeMediaQuery((theme) => theme.breakpoints.down('lg'));

	const [isOpen, setIsOpen] = useState(open);

	const handleToggleDrawer = useCallback((val: boolean) => {
		setIsOpen(val);
	}, []);

	useImperativeHandle(ref, () => ({
		toggleSidebar: handleToggleDrawer
	}));

	useEffect(() => {
		handleToggleDrawer(open);
	}, [handleToggleDrawer, open]);

	if (!content && !children) {
		return null;
	}

	return (
		<>
			{((variant === 'permanent' && isMobile) || variant !== 'permanent') && (
				<SwipeableDrawer
					variant="temporary"
					anchor={position}
					open={isOpen}
					onOpen={() => {}}
					onClose={() => onClose()}
					disableSwipeToOpen
					classes={{
						root: clsx('PageCarded-sidebarWrapper', variant),
						paper: clsx(
							'PageCarded-sidebar',
							variant,
							position === 'left' ? 'PageCarded-leftSidebar' : 'PageCarded-rightSidebar',
							'max-w-full min-w-80'
						)
					}}
					ModalProps={{
						keepMounted: true // Better open performance on mobile.
					}}
					slotProps={{
						backdrop: {
							classes: {
								root: 'PageCarded-backdrop'
							}
						}
					}}
					sx={{ position: 'absolute', '& .MuiPaper-root': { width: isMobile ? 'auto' : `${width}px` } }}
				>
					<PageCardedSidebarContent {...props} />
				</SwipeableDrawer>
			)}
			{variant === 'permanent' && !isMobile && (
				<Drawer
					variant="permanent"
					anchor={position}
					className={clsx(
						'PageCarded-sidebarWrapper',
						variant,
						isOpen ? 'opened' : 'closed',
						position === 'left' ? 'PageCarded-leftSidebar' : 'PageCarded-rightSidebar'
					)}
					open={isOpen}
					onClose={onClose}
					classes={{
						paper: clsx('PageCarded-sidebar', variant)
					}}
					sx={{ '& .MuiPaper-root': { width: `${width}px` } }}
				>
					<PageCardedSidebarContent {...props} />
				</Drawer>
			)}
		</>
	);
}

export default PageCardedSidebar;
