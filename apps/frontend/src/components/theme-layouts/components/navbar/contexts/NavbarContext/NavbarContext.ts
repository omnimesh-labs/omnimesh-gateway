import { createContext } from 'react';

type NavbarContextType = {
	isOpen: boolean;
	mobileOpen: boolean;
	foldedOpen: boolean;
	toggleFolded: () => void;
	openFolded: () => void;
	closeFolded: () => void;
	toggleMobileNavbar: () => void;
	openMobileNavbar: () => void;
	closeMobileNavbar: () => void;
	closeNavbar: () => void;
	openNavbar: () => void;
	toggleNavbar: () => void;
	reset: () => void;
};

export const NavbarContext = createContext<NavbarContextType | undefined>(undefined);
