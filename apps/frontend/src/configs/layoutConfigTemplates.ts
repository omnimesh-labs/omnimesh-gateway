import { DeepPartial } from 'react-hook-form';
import { SettingsConfigType } from '@fuse/core/Settings/Settings';

export const layoutConfigOnlyMain: DeepPartial<SettingsConfigType>['layout'] = {
	config: {
		navbar: {
			display: false
		},
		toolbar: {
			display: false
		},
		footer: {
			display: false
		},
		leftSidePanel: {
			display: false
		},
		rightSidePanel: {
			display: false
		}
	}
};

export const layoutConfigOnlyMainFullWidth: DeepPartial<SettingsConfigType>['layout'] = {
	config: {
		...layoutConfigOnlyMain.config,
		mode: 'fullwidth'
	}
};

export const layoutNoContainer: DeepPartial<SettingsConfigType>['layout'] = {
	config: {
		mode: 'fullwidth'
	}
};
