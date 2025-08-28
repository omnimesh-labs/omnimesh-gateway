'use client';

import { Dialog, DialogTitle, DialogContent, DialogActions, Button, Typography, Box, Alert } from '@mui/material';
import SvgIcon from '@fuse/core/SvgIcon';
import { MCPServer } from '@/lib/client-api'; import type { MCPServer } from '@/lib/types';

interface UnregisterConfirmDialogProps {
	server: MCPServer | null;
	open: boolean;
	onClose: () => void;
	onConfirm: (serverId: string) => void;
	loading?: boolean;
}

export default function UnregisterConfirmDialog({
	server,
	open,
	onClose,
	onConfirm,
	loading = false
}: UnregisterConfirmDialogProps) {
	if (!server) return null;

	const handleConfirm = () => {
		onConfirm(server.id);
	};

	return (
		<Dialog
			open={open}
			onClose={onClose}
			maxWidth="sm"
			fullWidth
		>
			<DialogTitle className="flex items-center space-x-2">
				<SvgIcon
					size={24}
					className="text-red-500"
				>
					lucide:alert-triangle
				</SvgIcon>
				<Typography variant="h6">Unregister Server</Typography>
			</DialogTitle>

			<DialogContent>
				<Box className="space-y-4">
					<Typography variant="body1">
						Are you sure you want to unregister the server <strong>"{server.name}"</strong>?
					</Typography>

					<Alert
						severity="warning"
						className="mt-4"
					>
						<Typography variant="body2">
							<strong>This action cannot be undone.</strong> The server will be removed from your
							organization and all associated sessions will be terminated.
						</Typography>
					</Alert>

					<Box className="rounded bg-gray-50 p-3">
						<Typography
							variant="body2"
							color="textSecondary"
							className="mb-1"
						>
							Server Details:
						</Typography>
						<Typography variant="body2">
							<strong>Name:</strong> {server.name}
						</Typography>
						<Typography variant="body2">
							<strong>Protocol:</strong> {server.protocol.toUpperCase()}
						</Typography>
						<Typography variant="body2">
							<strong>Status:</strong> {server.status}
						</Typography>
						{server.description && (
							<Typography variant="body2">
								<strong>Description:</strong> {server.description}
							</Typography>
						)}
					</Box>
				</Box>
			</DialogContent>

			<DialogActions className="p-6 pt-0">
				<Button
					onClick={onClose}
					disabled={loading}
					color="inherit"
				>
					Cancel
				</Button>
				<Button
					onClick={handleConfirm}
					disabled={loading}
					color="error"
					variant="contained"
					startIcon={
						loading ? (
							<SvgIcon size={16}>lucide:loader-2</SvgIcon>
						) : (
							<SvgIcon size={16}>lucide:trash-2</SvgIcon>
						)
					}
				>
					{loading ? 'Unregistering...' : 'Unregister Server'}
				</Button>
			</DialogActions>
		</Dialog>
	);
}
