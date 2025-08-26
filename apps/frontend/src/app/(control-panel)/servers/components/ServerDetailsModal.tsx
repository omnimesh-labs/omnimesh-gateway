'use client';

import {
	Dialog,
	DialogTitle,
	DialogContent,
	DialogActions,
	Button,
	Typography,
	Box,
	Chip,
	Divider,
	Paper,
	IconButton,
	Tooltip,
	List,
	ListItem,
	ListItemText
} from '@mui/material';
import SvgIcon from '@fuse/core/SvgIcon';
import { MCPServer } from '@/lib/api';

interface ServerDetailsModalProps {
	server: MCPServer | null;
	open: boolean;
	onClose: () => void;
}

const getStatusColor = (status: string): 'success' | 'warning' | 'error' | 'default' => {
	switch (status) {
		case 'active':
			return 'success';
		case 'maintenance':
			return 'warning';
		case 'unhealthy':
			return 'error';
		default:
			return 'default';
	}
};

export default function ServerDetailsModal({ server, open, onClose }: ServerDetailsModalProps) {
	if (!server) return null;

	const formatDate = (dateString: string) => {
		const date = new Date(dateString);
		return date.toLocaleString('en-US', {
			year: 'numeric',
			month: 'long',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	};

	return (
		<Dialog
			open={open}
			onClose={onClose}
			maxWidth="md"
			fullWidth
		>
			<DialogTitle className="s-center flex justify-between">
				<Box className="s-center flex space-x-2">
					<SvgIcon size={24}>lucide:server</SvgIcon>
					<Typography variant="h6">{server.name}</Typography>
				</Box>
				<Chip
					size="small"
					label={server.status}
					color={getStatusColor(server.status)}
					sx={{ textTransform: 'capitalize' }}
				/>
			</DialogTitle>

			<DialogContent>
				<Box className="space-y-6">
					{/* Basic Information */}
					<Paper className="p-4">
						<Typography
							variant="h6"
							className="s-center mb-3 flex"
						>
							<SvgIcon
								size={20}
								className="mr-2"
							>
								lucide:info
							</SvgIcon>
							Basic Information
						</Typography>
						<Grid
							container
							spacing={2}
						>
							<Grid
								xs={12}
								sm={6}
							>
								<Typography
									variant="body2"
									color="textSecondary"
								>
									Name
								</Typography>
								<Typography
									variant="body1"
									className="font-medium"
								>
									{server.name}
								</Typography>
							</Grid>
							<Grid
								xs={12}
								sm={6}
							>
								<Typography
									variant="body2"
									color="textSecondary"
								>
									Protocol
								</Typography>
								<Chip
									size="small"
									label={server.protocol.toUpperCase()}
									variant="outlined"
								/>
							</Grid>
							<Grid xs={12}>
								<Typography
									variant="body2"
									color="textSecondary"
								>
									Description
								</Typography>
								<Typography variant="body1">
									{server.description || 'No description provided'}
								</Typography>
							</Grid>
							<Grid
								xs={12}
								sm={6}
							>
								<Typography
									variant="body2"
									color="textSecondary"
								>
									Active
								</Typography>
								<Chip
									size="small"
									label={server.is_active ? 'Yes' : 'No'}
									color={server.is_active ? 'success' : 'error'}
									variant="outlined"
								/>
							</Grid>
							<Grid
								xs={12}
								sm={6}
							>
								<Typography
									variant="body2"
									color="textSecondary"
								>
									Version
								</Typography>
								<Typography variant="body1">{server.version}</Typography>
							</Grid>
							<Grid xs={12}>
								<Divider className="my-3" />
							</Grid>
							<Grid
								xs={12}
								sm={6}
							>
								<Typography
									variant="body2"
									color="textSecondary"
								>
									Created
								</Typography>
								<Typography variant="body1">{formatDate(server.created_at)}</Typography>
							</Grid>
							<Grid
								xs={12}
								sm={6}
							>
								<Typography
									variant="body2"
									color="textSecondary"
								>
									Last Updated
								</Typography>
								<Typography variant="body1">{formatDate(server.updated_at)}</Typography>
							</Grid>
						</Grid>
					</Paper>

					{/* Connection Details */}
					<Paper className="p-4">
						<Typography
							variant="h6"
							className="s-center mb-3 flex"
						>
							<SvgIcon
								size={20}
								className="mr-2"
							>
								lucide:link
							</SvgIcon>
							Connection Details
						</Typography>
						<Grid
							container
							spacing={2}
						>
							{server.url && (
								<Grid xs={12}>
									<Typography
										variant="body2"
										color="textSecondary"
									>
										URL
									</Typography>
									<Box className="s-center flex space-x-2">
										<Typography
											variant="body1"
											className="break-all font-mono"
										>
											{server.url}
										</Typography>
										<Tooltip title="Copy URL">
											<IconButton
												size="small"
												onClick={() => navigator.clipboard.writeText(server.url)}
											>
												<SvgIcon size={16}>lucide:copy</SvgIcon>
											</IconButton>
										</Tooltip>
									</Box>
								</Grid>
							)}
							{server.command && (
								<Grid xs={12}>
									<Typography
										variant="body2"
										color="textSecondary"
									>
										Command
									</Typography>
									<Typography
										variant="body1"
										className="font-mono"
									>
										{server.command}
									</Typography>
								</Grid>
							)}
							{server.args && server.args.length > 0 && (
								<Grid xs={12}>
									<Typography
										variant="body2"
										color="textSecondary"
									>
										Arguments
									</Typography>
									<List
										dense
										className="rounded bg-gray-50"
									>
										{server.args.map((arg, index) => (
											<ListItem
												key={index}
												className="py-1"
											>
												<ListItemText
													primary={
														<Typography
															variant="body2"
															className="font-mono"
														>
															{arg}
														</Typography>
													}
												/>
											</ListItem>
										))}
									</List>
								</Grid>
							)}
							{server.working_dir && (
								<Grid xs={12}>
									<Typography
										variant="body2"
										color="textSecondary"
									>
										Working Directory
									</Typography>
									<Typography
										variant="body1"
										className="font-mono"
									>
										{server.working_dir}
									</Typography>
								</Grid>
							)}
						</Grid>
					</Paper>

					{/* Environment Variables */}
					{server.environment && server.environment.length > 0 && (
						<Paper className="p-4">
							<Typography
								variant="h6"
								className="s-center mb-3 flex"
							>
								<SvgIcon
									size={20}
									className="mr-2"
								>
									lucide:settings
								</SvgIcon>
								Environment Variables
							</Typography>
							<List
								dense
								className="rounded bg-gray-50"
							>
								{server.environment.map((env, index) => (
									<ListItem
										key={index}
										className="py-1"
									>
										<ListItemText
											primary={
												<Typography
													variant="body2"
													className="font-mono"
												>
													{env}
												</Typography>
											}
										/>
									</ListItem>
								))}
							</List>
						</Paper>
					)}

					{/* Health & Configuration */}
					<Paper className="p-4">
						<Typography
							variant="h6"
							className="s-center mb-3 flex"
						>
							<SvgIcon
								size={20}
								className="mr-2"
							>
								lucide:activity
							</SvgIcon>
							Health & Configuration
						</Typography>
						<Grid
							container
							spacing={2}
						>
							{server.health_check_url && (
								<Grid xs={12}>
									<Typography
										variant="body2"
										color="textSecondary"
									>
										Health Check URL
									</Typography>
									<Typography
										variant="body1"
										className="break-all font-mono"
									>
										{server.health_check_url}
									</Typography>
								</Grid>
							)}
							<Grid
								xs={12}
								sm={6}
							>
								<Typography
									variant="body2"
									color="textSecondary"
								>
									Timeout
								</Typography>
								<Typography variant="body1">{server.timeout}s</Typography>
							</Grid>
							<Grid
								xs={12}
								sm={6}
							>
								<Typography
									variant="body2"
									color="textSecondary"
								>
									Max Retries
								</Typography>
								<Typography variant="body1">{server.max_retries}</Typography>
							</Grid>
						</Grid>
					</Paper>

					{/* Metadata */}
					{server.metadata && Object.keys(server.metadata).length > 0 && (
						<Paper className="p-4">
							<Typography
								variant="h6"
								className="s-center mb-3 flex"
							>
								<SvgIcon
									size={20}
									className="mr-2"
								>
									lucide:tag
								</SvgIcon>
								Metadata
							</Typography>
							<Grid
								container
								spacing={2}
							>
								{Object.entries(server.metadata).map(([key, value]) => (
									<Grid
										xs={12}
										sm={6}
										key={key}
									>
										<Typography
											variant="body2"
											color="textSecondary"
										>
											{key}
										</Typography>
										<Typography
											variant="body1"
											className="break-all"
										>
											{value}
										</Typography>
									</Grid>
								))}
							</Grid>
						</Paper>
					)}
				</Box>
			</DialogContent>

			<DialogActions>
				<Button
					onClick={onClose}
					color="primary"
				>
					Close
				</Button>
			</DialogActions>
		</Dialog>
	);
}
