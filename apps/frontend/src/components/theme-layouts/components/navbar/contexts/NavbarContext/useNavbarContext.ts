import { useContext } from 'react';
import { NavbarContext } from './NavbarContext';

export function useNavbarContext() {
	const context = useContext(NavbarContext);

	if (context === undefined) {
		throw new Error('useNavbar must be used within a NavbarProvider');
	}

	return context;
}
