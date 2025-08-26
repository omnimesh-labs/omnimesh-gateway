import { ReactNode } from 'react';
import Dialog from '@mui/material/Dialog';

export interface DialogContentProps {
	handleClose: () => void;
	data?: unknown;
}

export interface DialogProps {
	id: string;
	open?: boolean;
	onClose?: (T: string) => void;
	content: (T: DialogContentProps) => ReactNode;
	data?: unknown;
	classes?: { paper?: string };
}

export default function AppDialog(props: DialogProps) {
	const { id, open = false, onClose, content, data } = props;

	function handleClose() {
		onClose?.(id);
	}

	return (
		<Dialog
			open={open}
			onClose={() => handleClose()}
		>
			{content?.({ handleClose, data })}
		</Dialog>
	);
}
