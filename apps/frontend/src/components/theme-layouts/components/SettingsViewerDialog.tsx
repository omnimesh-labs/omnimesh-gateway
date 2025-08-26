import { useState } from 'react';
import clsx from 'clsx';
import Button from '@mui/material/Button';
import SvgIcon from '@fuse/core/SvgIcon';
import Dialog from '@mui/material/Dialog';
import Highlight from '@fuse/core/Highlight';
import DialogTitle from '@mui/material/DialogTitle';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import qs from 'qs';
import Typography from '@mui/material/Typography';
import useSettings from '@fuse/core/Settings/hooks/useSettings';

type SettingsViewerDialogProps = {
	className?: string;
};

/**
 * The settings viewer dialog.
 */
function SettingsViewerDialog(props: SettingsViewerDialogProps) {
	const { className = '' } = props;

	const [openDialog, setOpenDialog] = useState(false);
	const { data: settings } = useSettings();

	const jsonStringifiedSettings = JSON.stringify(settings);
	const queryString = qs.stringify({
		defaultSettings: jsonStringifiedSettings,
		strictNullHandling: true
	});

	function handleOpenDialog() {
		setOpenDialog(true);
	}

	function handleCloseDialog() {
		setOpenDialog(false);
	}

	return (
		<div className={clsx('', className)}>
			<Button
				variant="contained"
				color="secondary"
				className="w-full"
				onClick={handleOpenDialog}
				startIcon={<SvgIcon>lucide:code-xml</SvgIcon>}
			>
				View settings as json/query params
			</Button>

			<Dialog
				open={openDialog}
				onClose={handleCloseDialog}
				aria-labelledby="form-dialog-title"
			>
				<DialogTitle> Settings Viewer</DialogTitle>
				<DialogContent>
					<Typography className="mb-4 mt-6 text-lg font-bold">JSON</Typography>

					<Highlight
						component="pre"
						className="language-json"
					>
						{JSON.stringify(settings, null, 2)}
					</Highlight>

					<Typography className="mb-4 mt-6 text-lg font-bold">Query Params</Typography>

					{queryString}
				</DialogContent>
				<DialogActions>
					<Button
						color="secondary"
						variant="contained"
						onClick={handleCloseDialog}
					>
						Close
					</Button>
				</DialogActions>
			</Dialog>
		</div>
	);
}

export default SettingsViewerDialog;
