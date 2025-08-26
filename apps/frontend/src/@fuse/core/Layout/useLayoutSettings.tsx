import { useContext } from 'react';
import LayoutSettingsContext from './LayoutSettingsContext';

const useLayoutSettings = () => {
	const context = useContext(LayoutSettingsContext);

	if (context === undefined) {
		throw new Error('useLayoutSettings must be used within a SettingsProvider');
	}

	return context;
};

export default useLayoutSettings;
