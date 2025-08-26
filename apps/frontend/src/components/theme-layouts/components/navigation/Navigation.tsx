'use client';
import Navigation from '@fuse/core/Navigation';
import clsx from 'clsx';
import { useMemo } from 'react';
import useThemeMediaQuery from '@fuse/hooks/useThemeMediaQuery';
import { NavigationProps } from '@fuse/core/Navigation/Navigation';
import useNavigationItems from './hooks/useNavigationItems';
import { useNavbarContext } from '../navbar/contexts/NavbarContext/useNavbarContext';
/**
 * NavigationWrapper
 */

type NavigationWrapperProps = Partial<NavigationProps>;

function NavigationWrapper(props: NavigationWrapperProps) {
	const { className = '', layout = 'vertical', dense, active } = props;
	const { data: navigation } = useNavigationItems();
	const { closeMobileNavbar } = useNavbarContext();

	const isMobile = useThemeMediaQuery((theme) => theme.breakpoints.down('lg'));

	return useMemo(() => {
		function handleItemClick(item) {
			if (item?.url && isMobile) {
				closeMobileNavbar();
			}
		}

		return (
			<Navigation
				className={clsx('navigation flex-1', className)}
				navigation={navigation}
				layout={layout}
				dense={dense}
				active={active}
				onItemClick={handleItemClick}
				checkPermission
			/>
		);
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [isMobile, navigation, active, className, dense, layout]);
}

export default NavigationWrapper;
