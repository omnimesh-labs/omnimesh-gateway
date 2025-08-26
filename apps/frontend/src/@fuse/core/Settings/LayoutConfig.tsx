import { Control } from 'react-hook-form';
import { Typography } from '@mui/material';
import { AnyFormFieldType } from '@fuse/core/Settings/ThemeFormConfigTypes';
import { SettingsConfigType } from '@fuse/core/Settings/Settings';
import LayoutConfigs from './LayoutConfigs';
import RadioFormController from './form-controllers/RadioFormController';
import SwitchFormController from './form-controllers/SwitchFormController';
import NumberFormController from './form-controllers/NumberFormController';

type SettingsControllerProps = {
	key?: string;
	name: keyof SettingsConfigType;
	control: Control<SettingsConfigType>;
	title?: string;
	item: AnyFormFieldType;
};

function LayoutConfig(props: SettingsControllerProps) {
	const { item, name, control } = props;

	switch (item.type) {
		case 'radio':
			return (
				<RadioFormController
					name={name}
					control={control}
					item={item}
				/>
			);
		case 'switch':
			return (
				<SwitchFormController
					name={name}
					control={control}
					item={item}
				/>
			);
		case 'number':
			return (
				<NumberFormController
					name={name}
					control={control}
					item={item}
				/>
			);
		case 'group':
			return (
				<div
					key={name}
					className="Settings-formGroup"
				>
					<Typography
						className="Settings-formGroupTitle"
						color="text.secondary"
					>
						{item.title}
					</Typography>
					<LayoutConfigs
						value={item.children}
						prefix={name}
						control={control}
					/>
				</div>
			);
		default:
			return '';
	}
}

export default LayoutConfig;
