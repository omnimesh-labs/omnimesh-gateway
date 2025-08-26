import { useState } from 'react';

import { ReactNode } from 'react';
import { QuickPanelContext, QuickPanelData } from './QuickPanelContext';

interface QuickPanelProviderProps {
	children: ReactNode;
}

export const QuickPanelProvider: React.FC<QuickPanelProviderProps> = ({ children }) => {
	const [data, setData] = useState<QuickPanelData>();
	const [open, setOpen] = useState(false);

	const removeEvents = () => {
		setData((prevData) => ({
			...prevData,
			events: []
		}));
	};

	const toggleQuickPanel = () => {
		setOpen((prevOpen) => !prevOpen);
	};

	const openQuickPanel = () => {
		setOpen(true);
	};

	const closeQuickPanel = () => {
		setOpen(false);
	};

	const value = {
		data,
		open,
		removeEvents,
		toggleQuickPanel,
		openQuickPanel,
		closeQuickPanel
	};

	return <QuickPanelContext.Provider value={value}>{children}</QuickPanelContext.Provider>;
};
