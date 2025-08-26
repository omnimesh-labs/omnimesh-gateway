import { Control } from 'react-hook-form';
import LayoutConfig from './LayoutConfig';
import ThemeFormConfigTypes from './ThemeFormConfigTypes';
import { SettingsConfigType } from './Settings';

type SettingsControllersProps = {
	value: ThemeFormConfigTypes;
	prefix: string;
	control: Control<SettingsConfigType>;
};

function LayoutConfigs(props: SettingsControllersProps) {
	const { value, prefix, control } = props;

	return Object?.entries?.(value)?.map?.(([key, item]) => {
		const name = prefix ? `${prefix}.${key}` : key;
		return (
			<LayoutConfig
				key={key}
				name={name as keyof SettingsConfigType}
				control={control}
				item={item}
			/>
		);
	});
}

export default LayoutConfigs;
