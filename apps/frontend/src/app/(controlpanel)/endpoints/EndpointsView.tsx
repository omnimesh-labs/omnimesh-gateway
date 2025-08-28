'use client';

import { useState, useMemo, useEffect, useCallback } from 'react';
import { useSearchParams } from 'next/navigation';
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
import CircularProgress from '@mui/material/CircularProgress';
import LazyDataTable from '@/components/data-table/LazyDataTable';
import SvgIcon from '@fuse/core/SvgIcon';
import { useSnackbar } from 'notistack';
// Import types and client API
import type { Endpoint, Namespace } from '@/lib/types';
import { endpointApi, namespaceApi } from '@/lib/client-api';

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
	const [namespaces, setNamespaces] = useState<Namespace[]>([]);
	const [loading, setLoading] = useState(true);
	const [togglingEndpointId, setTogglingEndpointId] = useState<string | null>(null);
	const [highlightedEndpointId, setHighlightedEndpointId] = useState<string | null>(null);
	const { enqueueSnackbar } = useSnackbar();
	const searchParams = useSearchParams();

	// Fetch data on client side
	useEffect(() => {
		const fetchData = async () => {
			try {
				setLoading(true);
				const [endpointsData, namespacesData] = await Promise.all([
					endpointApi.listEndpoints(),
					namespaceApi.listNamespaces()
				]);
				setEndpoints(endpointsData);
				setNamespaces(namespacesData);
			} catch (error) {
				console.error('Error fetching data:', error);
				enqueueSnackbar('Failed to load data', { variant: 'error' });
			} finally {
				setLoading(false);
			}
		};

		fetchData();
	}, [enqueueSnackbar]);

	// Handle query parameters for auto-opening create modal and highlighting
	useEffect(() => {
		const action = searchParams.get('action');
		const namespaceId = searchParams.get('namespace_id');
		const highlightId = searchParams.get('highlight');

		if (action === 'create') {
			setFormData(prev => ({
				...prev,
				namespace_id: namespaceId || ''
			}));
			handleCreateEndpoint();
		}

		if (highlightId) {
			setHighlightedEndpointId(highlightId);
			// Clear highlight after 3 seconds
			setTimeout(() => setHighlightedEndpointId(null), 3000);
		}
	}, [searchParams]);

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
		// Validate that the selected namespace exists and is active
		const selectedNamespace = namespaces.find(ns => ns.id === formData.namespace_id);
		if (!selectedNamespace) {
			enqueueSnackbar('Please select a valid namespace.', { variant: 'error' });
			return;
		}
		if (!selectedNamespace.is_active) {
			enqueueSnackbar('Selected namespace is not active. Please select an active namespace.', { variant: 'error' });
			return;
		}

		try {
			let result;
			if (editingEndpoint) {
				result = await endpointApi.updateEndpoint(editingEndpoint.id, formData);
			} else {
				result = await endpointApi.createEndpoint(formData);
			}

			const action = editingEndpoint ? 'updated' : 'created';
			enqueueSnackbar(`Endpoint ${action} successfully`, { variant: 'success' });
			setCreateModalOpen(false);
			setEditingEndpoint(null);
			// Update local state with the new/updated endpoint
			if (editingEndpoint) {
				setEndpoints(prev => prev.map(ep => ep.id === result.id ? result : ep));
			} else {
				setEndpoints(prev => [...prev, result]);
			}
		} catch (error) {
			console.error('Error saving endpoint:', error);
			enqueueSnackbar('Failed to save endpoint', { variant: 'error' });
		}
	};

	const handleDeleteEndpoint = async (endpoint: Endpoint) => {
		if (!confirm(`Are you sure you want to delete "${endpoint.name}"? This action cannot be undone.`)) {
			return;
		}

		try {
			await endpointApi.deleteEndpoint(endpoint.id);
			enqueueSnackbar(`Endpoint ${endpoint.name} deleted successfully`, { variant: 'success' });
			setEndpoints(prev => prev.filter(ep => ep.id !== endpoint.id));
		} catch (error) {
			console.error('Error deleting endpoint:', error);
			enqueueSnackbar('Failed to delete endpoint', { variant: 'error' });
		}
	};

	const handleCopyUrl = (url: string) => {
		navigator.clipboard.writeText(url);
		enqueueSnackbar('URL copied to clipboard', { variant: 'success' });
	};

	const handleToggleEndpointStatus = useCallback(async (endpoint: Endpoint) => {
		setTogglingEndpointId(endpoint.id);
		try {
			const result = await endpointApi.updateEndpoint(endpoint.id, {
				is_active: !endpoint.is_active
			});

			// Update the local state with the updated endpoint
			setEndpoints(prev => prev.map(ep =>
				ep.id === endpoint.id ? result : ep
			));

			// Show specific toast based on activation/deactivation
			const action = !endpoint.is_active ? 'activated' : 'deactivated';
			enqueueSnackbar(`Endpoint ${action} successfully`, { variant: 'success' });
		} catch (error) {
			console.error('Error updating endpoint status:', error);
			enqueueSnackbar('Failed to update endpoint status', { variant: 'error' });
		} finally {
			setTogglingEndpointId(null);
		}
	}, [enqueueSnackbar]);

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
		[togglingEndpointId, handleToggleEndpointStatus]
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
					{loading && endpoints.length === 0 ? (
						<Box display="flex" justifyContent="center" alignItems="center" minHeight="300px">
							<CircularProgress />
						</Box>
					) : (
						<LazyDataTable
						columns={columns}
						data={endpoints}
						state={{ isLoading: loading }}
						enableRowActions
						muiTableBodyRowProps={({ row }) => ({
							sx: {
								backgroundColor: highlightedEndpointId === row.original.id
									? 'primary.50'
									: 'inherit',
								transition: 'background-color 0.3s ease',
								'&:hover': {
									backgroundColor: highlightedEndpointId === row.original.id
										? 'primary.100'
										: 'action.hover'
								}
							}
						})}
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
										<SvgIcon size={18}>lucide:pencil</SvgIcon>
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
					)}

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
									helperText={
										!namespaces || namespaces.length === 0 ? "No namespaces available. Please create a namespace first." :
										namespaces.filter(ns => ns.is_active).length === 0 ? "No active namespaces available." :
										"Select the namespace for this endpoint"
									}
									error={!namespaces || namespaces.length === 0}
								>
									{!namespaces || namespaces.length === 0 ? (
										<MenuItem disabled>No namespaces available</MenuItem>
									) : (
										namespaces
											.filter(namespace => namespace.is_active) // Only show active namespaces
											.map((namespace) => (
												<MenuItem key={namespace.id} value={namespace.id}>
													<Box>
														<Typography variant="body2">{namespace.name}</Typography>
														{namespace.description && (
															<Typography variant="caption" color="textSecondary">
																{namespace.description}
															</Typography>
														)}
													</Box>
												</MenuItem>
											))
									)}
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
								disabled={
									!formData.name.trim() ||
									!formData.namespace_id.trim() ||
									!namespaces.some(ns => ns.id === formData.namespace_id && ns.is_active)
								}
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
