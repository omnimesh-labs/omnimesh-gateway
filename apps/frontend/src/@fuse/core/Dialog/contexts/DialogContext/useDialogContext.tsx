import { useContext } from 'react';
import { DialogContext } from './DialogContext';

// Dialog hook to access the context
export function useDialogContext() {
	const context = useContext(DialogContext);

	if (context === null) {
		throw new Error('useDialogContext must be used within a AppProvider');
	}

	return context;
}
