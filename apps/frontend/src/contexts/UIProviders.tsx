'use client';

import { memo, ReactNode } from 'react';
import { DialogContextProvider } from '@fuse/core/Dialog/contexts/DialogContext/DialogContextProvider';
import { NavbarContextProvider } from '@/components/theme-layouts/components/navbar/contexts/NavbarContext/NavbarContextProvider';
import { QuickPanelProvider } from '@/components/theme-layouts/components/quickPanel/contexts/QuickPanelContext/QuickPanelContextProvider';
import { NavigationContextProvider } from '@/components/theme-layouts/components/navigation/contexts/NavigationContextProvider';

type UIProvidersProps = {
	children: ReactNode;
};

/**
 * Combined UI context providers to reduce component tree depth
 */
const UIProviders = memo(({ children }: UIProvidersProps) => {
	return (
		<NavbarContextProvider>
			<NavigationContextProvider>
				<DialogContextProvider>
					<QuickPanelProvider>{children}</QuickPanelProvider>
				</DialogContextProvider>
			</NavigationContextProvider>
		</NavbarContextProvider>
	);
});

UIProviders.displayName = 'UIProviders';

export default UIProviders;