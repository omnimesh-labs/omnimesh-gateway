'use client';

import { useState, useMemo, useEffect, useCallback } from 'react';
import { MRT_ColumnDef } from 'material-react-table';
import PageSimple from '@fuse/core/PageSimple';
import { styled } from '@mui/material/styles';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import IconButton from '@mui/material/IconButton';
import Tooltip from '@mui/material/Tooltip';
import Box from '@mui/material/Box';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import TextField from '@mui/material/TextField';
import Stack from '@mui/material/Stack';
import FormControlLabel from '@mui/material/FormControlLabel';
import Switch from '@mui/material/Switch';
import MenuItem from '@mui/material/MenuItem';
import Accordion from '@mui/material/Accordion';
import AccordionSummary from '@mui/material/AccordionSummary';
import AccordionDetails from '@mui/material/AccordionDetails';
import LazyDataTable from '@/components/data-table/LazyDataTable';
import SvgIcon from '@fuse/core/SvgIcon';
import { useSnackbar } from 'notistack';
import { endpointApi } from '@/lib/api';

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

import { Endpoint } from '@/lib/api';

function EndpointsView() {
	const [createModalOpen, setCreateModalOpen] = useState(false);
	const [editingEndpoint, setEditingEndpoint] = useState<Endpoint | null>(null);
	const [viewingUrls, setViewingUrls] = useState<Endpoint | null>(null);
	const [urlsModalOpen, setUrlsModalOpen] = useState(false);
	const [formData, setFormData] = useState({
		name: '',
		namespace_id: '',
		description: '',
		enable_api_key_auth: true,
		enable_oauth: false,
		enable_public_access: false,
		rate_limit_requests: 100,
		rate_limit_window: 3600,
		is_active: true
	});
	const [endpoints, setEndpoints] = useState<Endpoint[]>([]);
	const [isLoading, setIsLoading] = useState(false);
	const [togglingEndpointId, setTogglingEndpointId] = useState<string | null>(null);
	const { enqueueSnackbar } = useSnackbar();

	const fetchEndpoints = useCallback(async () => {
		setIsLoading(true);
		try {
			const data = await endpointApi.listEndpoints();
			setEndpoints(data);
		} catch (error) {
			enqueueSnackbar('Failed to fetch endpoints', { variant: 'error' });
			console.error('Error fetching endpoints:', error);
			setEndpoints([]); // Set empty array on error
		} finally {
			setIsLoading(false);
		}
	}, [enqueueSnackbar]);

	// Fetch endpoints on mount
	useEffect(() => {
		fetchEndpoints();
	}, [fetchEndpoints]);

	const handleCreateEndpoint = () => {
		setFormData({
			name: '',
			namespace_id: '',
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
			namespace_id: endpoint.namespace?.id || '',
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

	const handleSaveEndpoint = async () => {
		try {
			if (editingEndpoint) {
				await endpointApi.updateEndpoint(editingEndpoint.id, formData);
				enqueueSnackbar('Endpoint updated successfully', { variant: 'success' });
			} else {
				await endpointApi.createEndpoint(formData);
				enqueueSnackbar('Endpoint created successfully', { variant: 'success' });
			}

			setCreateModalOpen(false);
			setEditingEndpoint(null);
			fetchEndpoints(); // Refresh the list
		} catch (error) {
			const action = editingEndpoint ? 'update' : 'create';
			enqueueSnackbar(`Failed to ${action} endpoint`, { variant: 'error' });
			console.error(`Error ${action}ing endpoint:`, error);
		}
	};

	const handleDeleteEndpoint = async (endpoint: Endpoint) => {
		try {
			await endpointApi.deleteEndpoint(endpoint.id);
			enqueueSnackbar(`Endpoint ${endpoint.name} deleted successfully`, { variant: 'success' });
			fetchEndpoints(); // Refresh the list
		} catch (error) {
			enqueueSnackbar(`Failed to delete endpoint ${endpoint.name}`, { variant: 'error' });
			console.error('Error deleting endpoint:', error);
		}
	};

	const handleCopyUrl = (url: string) => {
		navigator.clipboard.writeText(url);
		enqueueSnackbar('URL copied to clipboard', { variant: 'success' });
	};

	const handleToggleEndpointStatus = async (endpoint: Endpoint) => {
		setTogglingEndpointId(endpoint.id);
		try {
			const updatedEndpoint = await endpointApi.updateEndpoint(endpoint.id, {
				is_active: !endpoint.is_active
			});

			// Update the local state with the updated endpoint
			setEndpoints(prev => prev.map(ep =>
				ep.id === endpoint.id ? updatedEndpoint : ep
			));

			// Show specific toast based on activation/deactivation
			if (!endpoint.is_active) {
				enqueueSnackbar('Endpoint activated successfully', { variant: 'success' });
			} else {
				enqueueSnackbar('Endpoint deactivated successfully', { variant: 'success' });
			}
		} catch (error) {
			const message = error instanceof Error ? error.message : 'Failed to update endpoint status';
			enqueueSnackbar(message, { variant: 'error' });
			console.error('Error updating endpoint status:', error);
			// Refresh data on error to ensure UI is in sync
			fetchEndpoints();
		} finally {
			setTogglingEndpointId(null);
		}
	};

	const columns = useMemo<MRT_ColumnDef<Endpoint>[]>(
		() => [
			{
				accessorKey: 'is_active',
				header: 'Active',
				size: 80,
				Cell: ({ row }) => (
					<Tooltip title={row.original.is_active ? 'Deactivate endpoint' : 'Activate endpoint'}>
						<Switch
							checked={!!row.original.is_active}
							onChange={() => handleToggleEndpointStatus(row.original)}
							disabled={togglingEndpointId === row.original.id}
							size="small"
							color="success"
						/>
					</Tooltip>
				)
			},
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
								{row.original.namespace?.name || 'No namespace'}
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
		],
		[togglingEndpointId]
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
						data={endpoints}
						state={{ isLoading }}
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
									value={formData.namespace_id}
									onChange={(e) => setFormData((prev) => ({ ...prev, namespace_id: e.target.value }))}
									select
									fullWidth
									required
								>
									<MenuItem value="ns-dev-001">Development</MenuItem>
									<MenuItem value="ns-prod-001">Production</MenuItem>
									<MenuItem value="ns-test-001">Testing</MenuItem>
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
								disabled={!formData.name.trim() || !formData.namespace_id.trim()}
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
									{Object.entries(viewingUrls.urls || {}).map(([protocol, url]) => (
										<Box
											key={protocol}
											className="flex items-center justify-between"
										>
											<Box>
												<Typography
													variant="subtitle1"
													className="capitalize"
												>
													{protocol.replace('_', ' ').toUpperCase()}
												</Typography>
												<Typography
													variant="body2"
													color="textSecondary"
													className="break-all"
												>
													{String(url)}
												</Typography>
											</Box>
											<IconButton
												size="small"
												onClick={() => handleCopyUrl(String(url))}
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
