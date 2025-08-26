import { SettingsConfigType, ThemesType } from '@fuse/core/Settings/Settings';
import { createContext } from 'react';

// SettingsContext type
export type SettingsContextType = {
	data: SettingsConfigType;
	setSettings: (newSettings: Partial<SettingsConfigType>) => SettingsConfigType;
	changeTheme: (newTheme: ThemesType) => void;
};

// Context with a default value of undefined
const SettingsContext = createContext<SettingsContextType | undefined>(undefined);

export default SettingsContext;
