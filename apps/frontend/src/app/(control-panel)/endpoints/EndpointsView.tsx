'use client';

import { useState, useMemo } from 'react';
import { MRT_ColumnDef } from 'material-react-table';
import PageSimple from '@fuse/core/PageSimple';
import { styled } from '@mui/material/styles';
import {
	Typography,
	Button,
	Chip,
	IconButton,
	Tooltip,
	Box,
	Dialog,
	DialogTitle,
	DialogContent,
	DialogActions,
	TextField,
	Stack,
	FormControlLabel,
	Switch,
	MenuItem,
	Accordion,
	AccordionSummary,
	AccordionDetails
} from '@mui/material';
import LazyDataTable from '@/components/data-table/LazyDataTable';
import SvgIcon from '@fuse/core/SvgIcon';
import { useSnackbar } from 'notistack';

const Root = styled(PageSimple)(({ theme }) => ({
	'& .PageSimple-header': {
		backgroundColor: theme.vars.palette.background.paper,
		borderBottomWidth: 1,
		borderStyle: 'solid',
		borderColor: theme.vars.palette.divider
	},
	'& .PageSimple-content': {
		backgroundColor: theme.vars.palette.background.default
	}
}));

interface Endpoint {
	id: string;
	name: string;
	namespace: string;
	description?: string;
	enable_api_key_auth: boolean;
	enable_oauth: boolean;
	enable_public_access: boolean;
	rate_limit_requests: number;
	rate_limit_window: number;
	is_active: boolean;
	created_at: string;
	urls: {
		sse: string;
		http: string;
		websocket: string;
		openapi: string;
		documentation: string;
	};
}

function EndpointsView() {
	const [createModalOpen, setCreateModalOpen] = useState(false);
	const [editingEndpoint, setEditingEndpoint] = useState<Endpoint | null>(null);
	const [viewingUrls, setViewingUrls] = useState<Endpoint | null>(null);
	const [urlsModalOpen, setUrlsModalOpen] = useState(false);
	const [formData, setFormData] = useState({
		name: '',
		namespace: '',
		description: '',
		enable_api_key_auth: true,
		enable_oauth: false,
		enable_public_access: false,
		rate_limit_requests: 100,
		rate_limit_window: 3600,
		is_active: true
	});
	const { enqueueSnackbar } = useSnackbar();

	// Mock data
	const mockEndpoints: Endpoint[] = [
		{
			id: '1',
			name: 'dev-api',
			namespace: 'development',
			description: 'Development API endpoint',
			enable_api_key_auth: true,
			enable_oauth: false,
			enable_public_access: false,
			rate_limit_requests: 100,
			rate_limit_window: 3600,
			is_active: true,
			created_at: '2024-01-15T10:00:00Z',
			urls: {
				sse: 'https://api.example.com/dev/sse',
				http: 'https://api.example.com/dev/http',
				websocket: 'wss://api.example.com/dev/ws',
				openapi: 'https://api.example.com/dev/openapi.json',
				documentation: 'https://docs.example.com/dev'
			}
		},
		{
			id: '2',
			name: 'prod-api',
			namespace: 'production',
			description: 'Production API endpoint',
			enable_api_key_auth: true,
			enable_oauth: true,
			enable_public_access: false,
			rate_limit_requests: 1000,
			rate_limit_window: 3600,
			is_active: true,
			created_at: '2024-01-10T09:00:00Z',
			urls: {
				sse: 'https://api.example.com/prod/sse',
				http: 'https://api.example.com/prod/http',
				websocket: 'wss://api.example.com/prod/ws',
				openapi: 'https://api.example.com/prod/openapi.json',
				documentation: 'https://docs.example.com/prod'
			}
		}
	];

	const handleCreateEndpoint = () => {
		setFormData({
			name: '',
			namespace: '',
			description: '',
			enable_api_key_auth: true,
			enable_oauth: false,
			enable_public_access: false,
			rate_limit_requests: 100,
			rate_limit_window: 3600,
			is_active: true
		});
		setEditingEndpoint(null);
		setCreateModalOpen(true);
	};

	const handleEditEndpoint = (endpoint: Endpoint) => {
		setFormData({
			name: endpoint.name,
			namespace: endpoint.namespace,
			description: endpoint.description || '',
			enable_api_key_auth: endpoint.enable_api_key_auth,
			enable_oauth: endpoint.enable_oauth,
			enable_public_access: endpoint.enable_public_access,
			rate_limit_requests: endpoint.rate_limit_requests,
			rate_limit_window: endpoint.rate_limit_window,
			is_active: endpoint.is_active
		});
		setEditingEndpoint(endpoint);
		setCreateModalOpen(true);
	};

	const handleViewUrls = (endpoint: Endpoint) => {
		setViewingUrls(endpoint);
		setUrlsModalOpen(true);
	};

	const handleSaveEndpoint = () => {
		const action = editingEndpoint ? 'updated' : 'created';
		enqueueSnackbar(`Endpoint ${action} successfully (demo)`, { variant: 'success' });
		setCreateModalOpen(false);
		setEditingEndpoint(null);
	};

	const handleDeleteEndpoint = (endpoint: Endpoint) => {
		enqueueSnackbar(`Delete functionality coming soon for ${endpoint.name}`, { variant: 'info' });
	};

	const handleCopyUrl = (url: string) => {
		navigator.clipboard.writeText(url);
		enqueueSnackbar('URL copied to clipboard', { variant: 'success' });
	};

	const columns = useMemo<MRT_ColumnDef<Endpoint>[]>(
		() => [
			{
				accessorKey: 'name',
				header: 'Name',
				size: 200,
				Cell: ({ row }) => (
					<Box className="flex items-center space-x-2">
						<SvgIcon size={20}>lucide:globe</SvgIcon>
						<Box>
							<Typography
								variant="body2"
								className="font-medium"
							>
								{row.original.name}
							</Typography>
							<Typography
								variant="caption"
								color="textSecondary"
							>
								{row.original.namespace}
							</Typography>
						</Box>
					</Box>
				)
			},
			{
				accessorKey: 'description',
				header: 'Description',
				size: 250,
				Cell: ({ cell }) => (
					<Typography variant="body2">{cell.getValue<string>() || 'No description'}</Typography>
				)
			},
			{
				accessorKey: 'authentication',
				header: 'Auth Methods',
				size: 200,
				Cell: ({ row }) => (
					<Box className="flex flex-wrap gap-1">
						{row.original.enable_api_key_auth && (
							<Chip
								size="small"
								label="API Key"
								color="primary"
							/>
						)}
						{row.original.enable_oauth && (
							<Chip
								size="small"
								label="OAuth"
								color="secondary"
							/>
						)}
						{row.original.enable_public_access && (
							<Chip
								size="small"
								label="Public"
								color="warning"
							/>
						)}
					</Box>
				)
			},
			{
				accessorKey: 'rate_limit_requests',
				header: 'Rate Limit',
				size: 120,
				Cell: ({ row }) => <Typography variant="body2">{row.original.rate_limit_requests}/hr</Typography>
			},
			{
				accessorKey: 'is_active',
				header: 'Status',
				size: 120,
				Cell: ({ cell }) => (
					<Chip
						size="small"
						label={cell.getValue<boolean>() ? 'Active' : 'Inactive'}
						color={cell.getValue<boolean>() ? 'success' : 'default'}
					/>
				)
			}
		],
		[]
	);

	return (
		<Root
			header={
				<div className="p-6">
					<div className="flex items-center justify-between">
						<div>
							<Typography variant="h4">Endpoints</Typography>
							<Typography
								variant="body1"
								color="textSecondary"
								className="mt-1"
							>
								Manage public-facing endpoints for your namespaces
							</Typography>
						</div>
						<Button
							variant="contained"
							color="primary"
							startIcon={<SvgIcon>lucide:plus</SvgIcon>}
							onClick={handleCreateEndpoint}
						>
							Create Endpoint
						</Button>
					</div>
				</div>
			}
			content={
				<div className="p-6">
					<LazyDataTable
						columns={columns}
						data={mockEndpoints}
						enableRowActions
						renderRowActions={({ row }) => (
							<Box className="flex items-center space-x-1">
								<Tooltip title="View URLs">
									<IconButton
										size="small"
										onClick={() => handleViewUrls(row.original)}
									>
										<SvgIcon size={18}>lucide:link</SvgIcon>
									</IconButton>
								</Tooltip>
								<Tooltip title="Edit Endpoint">
									<IconButton
										size="small"
										onClick={() => handleEditEndpoint(row.original)}
									>
										<SvgIcon size={18}>lucide:edit</SvgIcon>
									</IconButton>
								</Tooltip>
								<Tooltip title="Delete Endpoint">
									<IconButton
										size="small"
										color="error"
										onClick={() => handleDeleteEndpoint(row.original)}
									>
										<SvgIcon size={18}>lucide:trash-2</SvgIcon>
									</IconButton>
								</Tooltip>
							</Box>
						)}
						initialState={{
							pagination: {
								pageIndex: 0,
								pageSize: 10
							}
						}}
					/>

					{/* Create/Edit Endpoint Dialog */}
					<Dialog
						open={createModalOpen}
						onClose={() => setCreateModalOpen(false)}
						maxWidth="md"
						fullWidth
					>
						<DialogTitle>{editingEndpoint ? 'Edit Endpoint' : 'Create Endpoint'}</DialogTitle>
						<DialogContent>
							<Stack
								spacing={3}
								sx={{ mt: 1 }}
							>
								<TextField
									label="Name"
									value={formData.name}
									onChange={(e) => setFormData((prev) => ({ ...prev, name: e.target.value }))}
									fullWidth
									required
								/>
								<TextField
									label="Namespace"
									value={formData.namespace}
									onChange={(e) => setFormData((prev) => ({ ...prev, namespace: e.target.value }))}
									select
									fullWidth
									required
								>
									<MenuItem value="development">Development</MenuItem>
									<MenuItem value="production">Production</MenuItem>
									<MenuItem value="testing">Testing</MenuItem>
								</TextField>
								<TextField
									label="Description"
									value={formData.description}
									onChange={(e) => setFormData((prev) => ({ ...prev, description: e.target.value }))}
									fullWidth
									multiline
									rows={2}
								/>

								<Accordion>
									<AccordionSummary expandIcon={<SvgIcon>lucide:chevron-down</SvgIcon>}>
										<Typography>Authentication Settings</Typography>
									</AccordionSummary>
									<AccordionDetails>
										<Stack spacing={2}>
											<FormControlLabel
												control={
													<Switch
														checked={formData.enable_api_key_auth}
														onChange={(e) =>
															setFormData((prev) => ({
																...prev,
																enable_api_key_auth: e.target.checked
															}))
														}
													/>
												}
												label="Enable API Key Authentication"
											/>
											<FormControlLabel
												control={
													<Switch
														checked={formData.enable_oauth}
														onChange={(e) =>
															setFormData((prev) => ({
																...prev,
																enable_oauth: e.target.checked
															}))
														}
													/>
												}
												label="Enable OAuth Authentication"
											/>
											<FormControlLabel
												control={
													<Switch
														checked={formData.enable_public_access}
														onChange={(e) =>
															setFormData((prev) => ({
																...prev,
																enable_public_access: e.target.checked
															}))
														}
													/>
												}
												label="Enable Public Access"
											/>
										</Stack>
									</AccordionDetails>
								</Accordion>

								<Accordion>
									<AccordionSummary expandIcon={<SvgIcon>lucide:chevron-down</SvgIcon>}>
										<Typography>Rate Limiting</Typography>
									</AccordionSummary>
									<AccordionDetails>
										<Stack spacing={2}>
											<TextField
												label="Requests per Hour"
												type="number"
												value={formData.rate_limit_requests}
												onChange={(e) =>
													setFormData((prev) => ({
														...prev,
														rate_limit_requests: parseInt(e.target.value) || 0
													}))
												}
												fullWidth
											/>
											<TextField
												label="Window (seconds)"
												type="number"
												value={formData.rate_limit_window}
												onChange={(e) =>
													setFormData((prev) => ({
														...prev,
														rate_limit_window: parseInt(e.target.value) || 3600
													}))
												}
												fullWidth
											/>
										</Stack>
									</AccordionDetails>
								</Accordion>

								<FormControlLabel
									control={
										<Switch
											checked={formData.is_active}
											onChange={(e) =>
												setFormData((prev) => ({ ...prev, is_active: e.target.checked }))
											}
										/>
									}
									label="Active"
								/>
							</Stack>
						</DialogContent>
						<DialogActions>
							<Button onClick={() => setCreateModalOpen(false)}>Cancel</Button>
							<Button
								variant="contained"
								onClick={handleSaveEndpoint}
								disabled={!formData.name.trim() || !formData.namespace.trim()}
							>
								{editingEndpoint ? 'Update' : 'Create'}
							</Button>
						</DialogActions>
					</Dialog>

					{/* View URLs Dialog */}
					<Dialog
						open={urlsModalOpen}
						onClose={() => setUrlsModalOpen(false)}
						maxWidth="md"
						fullWidth
					>
						<DialogTitle>Endpoint URLs - {viewingUrls?.name}</DialogTitle>
						<DialogContent>
							{viewingUrls && (
								<Stack
									spacing={3}
									sx={{ mt: 1 }}
								>
									{Object.entries(viewingUrls.urls).map(([protocol, url]) => (
										<Box
											key={protocol}
											className="flex items-center justify-between"
										>
											<Box>
												<Typography
													variant="subtitle2"
													className="capitalize"
												>
													{protocol.replace('_', ' ')}
												</Typography>
												<Typography
													variant="body2"
													color="textSecondary"
													className="break-all"
												>
													{url}
												</Typography>
											</Box>
											<IconButton
												size="small"
												onClick={() => handleCopyUrl(url)}
												title="Copy URL"
											>
												<SvgIcon size={16}>lucide:copy</SvgIcon>
											</IconButton>
										</Box>
									))}
								</Stack>
							)}
						</DialogContent>
						<DialogActions>
							<Button onClick={() => setUrlsModalOpen(false)}>Close</Button>
						</DialogActions>
					</Dialog>
				</div>
			}
		/>
	);
}

export default EndpointsView;
