import { Control } from 'react-hook-form';
import { debounce } from 'lodash';
import { Controller } from 'react-hook-form';
import { SettingsConfigType } from '../Settings';
import { useEffect, useState } from 'react';
import { useRef } from 'react';
import { FormControl, FormLabel, TextField } from '@mui/material';

type NumberFormControllerProps = {
	name: keyof SettingsConfigType;
	control: Control<SettingsConfigType>;
	item: {
		title: string;
		min?: number;
		max?: number;
	};
};

function NumberFormController(props: NumberFormControllerProps) {
	const { name, control, item } = props;

	return (
		<div
			key={name}
			className="Settings-formControl"
		>
			<Controller
				name={name}
				control={control}
				render={({ field: { onChange, value } }) => (
					<NumberTextField
						name={name}
						value={+value}
						onChange={onChange}
						item={item}
					/>
				)}
			/>
		</div>
	);
}

type NumberTextFieldProps = {
	value: number;
	onChange: (value: number) => void;
	item: NumberFormControllerProps['item'];
	name: string;
};

function NumberTextField(props: NumberTextFieldProps) {
	const { value, onChange, item, name } = props;
	const [localValue, setLocalValue] = useState(value);
	const [error, setError] = useState('');

	const debouncedOnChange = useRef(
		debounce((newValue: number) => {
			onChange(newValue);
		}, 500)
	).current;

	const handleChange = (ev: React.ChangeEvent<HTMLInputElement>) => {
		const newValue = +ev.target.value;

		if (item?.min && newValue < item?.min) {
			setError(`Value is too low (min: ${item?.min})`);
			setLocalValue(newValue);
			return;
		}

		if (item?.max && newValue > item?.max) {
			setError(`Value is too high (max: ${item?.max})`);
			setLocalValue(newValue);
			return;
		}

		setError('');
		setLocalValue(newValue);
		debouncedOnChange(newValue);
	};

	useEffect(() => {
		setLocalValue(value);
	}, [value]);

	return (
		<FormControl>
			<FormLabel htmlFor={`${name}-input`}>{item.title}</FormLabel>
			<TextField
				id={`${name}-input`}
				value={localValue}
				onChange={handleChange}
				type="number"
				slotProps={{
					input: {
						inputProps: { min: item.min, max: item.max }
					}
				}}
				variant="outlined"
				error={!!error}
				helperText={error}
			/>
		</FormControl>
	);
}

export default NumberFormController;
