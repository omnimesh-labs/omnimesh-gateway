import { createContext } from 'react';
import { SettingsConfigType } from '@fuse/core/Settings/Settings';

type LayoutSettingsContextType = SettingsConfigType['layout'];

const LayoutSettingsContext = createContext<LayoutSettingsContextType | undefined>(undefined);

export default LayoutSettingsContext;
