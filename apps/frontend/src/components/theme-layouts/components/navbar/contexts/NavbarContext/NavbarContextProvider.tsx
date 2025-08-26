import { ReactNode, useState } from 'react';
import { NavbarContext } from './NavbarContext';

const initialState = {
	isOpen: true,
	mobileOpen: false,
	foldedOpen: false
};

type NavbarContextProviderProps = {
	children: ReactNode;
};

export function NavbarContextProvider({ children }: NavbarContextProviderProps) {
	const [state, setState] = useState(initialState);

	const toggleFolded = () => setState((prev) => ({ ...prev, foldedOpen: !prev.foldedOpen }));
	const openFolded = () => setState((prev) => ({ ...prev, foldedOpen: true }));
	const closeFolded = () => setState((prev) => ({ ...prev, foldedOpen: false }));

	const toggleMobileNavbar = () => setState((prev) => ({ ...prev, mobileOpen: !prev.mobileOpen }));
	const openMobileNavbar = () => setState((prev) => ({ ...prev, mobileOpen: true }));
	const closeMobileNavbar = () => setState((prev) => ({ ...prev, mobileOpen: false }));

	const closeNavbar = () => setState((prev) => ({ ...prev, isOpen: false }));
	const openNavbar = () => setState((prev) => ({ ...prev, isOpen: true }));
	const toggleNavbar = () => setState((prev) => ({ ...prev, isOpen: !prev.isOpen }));

	const reset = () => setState(initialState);

	const value = {
		...state,
		toggleFolded,
		openFolded,
		closeFolded,
		toggleMobileNavbar,
		openMobileNavbar,
		closeMobileNavbar,
		closeNavbar,
		openNavbar,
		toggleNavbar,
		reset
	};

	return <NavbarContext.Provider value={value}>{children}</NavbarContext.Provider>;
}
