import { useState, ReactNode, useMemo, useEffect, useCallback } from 'react';
import { merge, isEqual } from '../../../utils/lodashReplacements';
import { defaultSettings, getParsedQuerySettings } from '@fuse/default-settings';
import settingsConfig from 'src/configs/settingsConfig';
import themeLayoutConfigs from 'src/components/theme-layouts/themeLayoutConfigs';
import { SettingsConfigType, ThemesType } from '@fuse/core/Settings/Settings';
import useUser from '@auth/useUser';
import { PartialDeep } from 'type-fest';
import SettingsContext from './SettingsContext';

// Get initial settings
const getInitialSettings = (): SettingsConfigType => {
	const defaultLayoutStyle = settingsConfig.layout?.style || 'layout1';
	const layout = {
		style: defaultLayoutStyle,
		config: themeLayoutConfigs[defaultLayoutStyle]?.defaults
	};
	return merge({}, defaultSettings, { layout }, settingsConfig, getParsedQuerySettings());
};

const initialSettings = getInitialSettings();

const generateSettings = (_defaultSettings: SettingsConfigType, _newSettings: PartialDeep<SettingsConfigType>) => {
	return merge(
		{},
		_defaultSettings,
		{ layout: { config: themeLayoutConfigs[_newSettings?.layout?.style]?.defaults } },
		_newSettings
	);
};

// SettingsProvider component
export function SettingsProvider({ children }: { children: ReactNode }) {
	const { data: user, isGuest } = useUser();

	const userSettings = useMemo(() => user?.settings || {}, [user]);

	const calculateSettings = useCallback(() => {
		const defaultSettings = merge({}, initialSettings);
		return isGuest ? defaultSettings : merge({}, defaultSettings, userSettings);
	}, [isGuest, userSettings]);

	const [data, setData] = useState<SettingsConfigType>(calculateSettings());

	// Sync data with userSettings when isGuest or userSettings change
	useEffect(() => {
		const newSettings = calculateSettings();

		// Only update if settings are different
		if (!isEqual(data, newSettings)) {
			setData(newSettings);
		}
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [calculateSettings]);

	const setSettings = useCallback(
		(newSettings: Partial<SettingsConfigType>) => {
			const _settings = generateSettings(data, newSettings);

			if (!isEqual(_settings, data)) {
				setData(merge({}, _settings));
			}

			return _settings;
		},
		[data]
	);

	const changeTheme = useCallback(
		(newTheme: ThemesType) => {
			const { navbar, footer, toolbar, main } = newTheme;

			const newSettings: SettingsConfigType = {
				...data,
				theme: {
					main,
					navbar,
					toolbar,
					footer
				}
			};

			setSettings(newSettings);
		},
		[data, setSettings]
	);

	return (
		<SettingsContext
			value={useMemo(
				() => ({
					data,
					setSettings,
					changeTheme
				}),
				[data, setSettings, changeTheme]
			)}
		>
			{children}
		</SettingsContext>
	);
}

export default SettingsProvider;
