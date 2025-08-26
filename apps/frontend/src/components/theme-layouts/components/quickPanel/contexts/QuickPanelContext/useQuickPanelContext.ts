import { useContext } from 'react';
import { QuickPanelContext, QuickPanelContextType } from './QuickPanelContext';

export const useQuickPanelContext = (): QuickPanelContextType => {
	const context = useContext(QuickPanelContext);

	if (context === undefined) {
		throw new Error('useQuickPanel must be used within a QuickPanelProvider');
	}

	return context;
};
