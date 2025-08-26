import React from 'react';
import AppDialog, { DialogProps } from '../../Dialog';
import { DialogContext, DialogDefaultContext } from './DialogContext';

interface DialogContextProviderProps {
	children: React.ReactNode;
}

export function DialogContextProvider(props: DialogContextProviderProps) {
	const { children } = props;
	const [dialogs, setDialogs] = React.useState(DialogDefaultContext.dialogs);

	function openDialog(dialogProps: DialogProps) {
		setDialogs((prev) => ({
			...prev,
			[dialogProps.id]: { ...dialogProps, open: true }
		}));
	}

	function closeDialog(id: DialogProps['id']) {
		setDialogs((prev) => {
			const newDialogs = { ...prev };
			delete newDialogs[id];

			return newDialogs;
		});
	}

	return (
		<DialogContext.Provider value={{ dialogs, openDialog, closeDialog }}>
			{children}
			{Object.entries(dialogs).map(([id, dialog]) => (
				<AppDialog
					{...dialog}
					key={id}
					onClose={closeDialog}
				/>
			))}
		</DialogContext.Provider>
	);
}
