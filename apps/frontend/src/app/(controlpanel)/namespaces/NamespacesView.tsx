'use client';

import { useState, useMemo, useEffect, useCallback } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
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
	Checkbox,
	Grid,
	Card,
	CardContent,
	Alert,
	Tab,
	Tabs,
	CircularProgress
} from '@mui/material';
import LazyDataTable from '@/components/data-table/LazyDataTable';
import SvgIcon from '@fuse/core/SvgIcon';
import { useSnackbar } from 'notistack';
import { namespaceApi, serverApi } from '@/lib/client-api';
import type { MCPServer, Namespace } from '@/lib/types';

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



function NamespacesView() {
	const searchParams = useSearchParams();
	const router = useRouter();
	const [createModalOpen, setCreateModalOpen] = useState(false);
	const [viewModalOpen, setViewModalOpen] = useState(false);
	const [editingNamespace, setEditingNamespace] = useState<Namespace | null>(null);
	const [viewingNamespace, setViewingNamespace] = useState<Namespace | null>(null);
	const [modalTab, setModalTab] = useState(0);
	const [formData, setFormData] = useState({
		name: '',
		description: '',
		is_active: true,
		servers: [] as string[]
	});
	const [nameError, setNameError] = useState('');
	const [namespaces, setNamespaces] = useState<Namespace[]>([]);
	const [isLoading, setIsLoading] = useState(false);
	const [servers, setServers] = useState<MCPServer[]>([]);
	const [loadingServers, setLoadingServers] = useState(false);
	const [loadingNamespaceDetails, setLoadingNamespaceDetails] = useState(false);
	const [togglingNamespaceId, setTogglingNamespaceId] = useState<string | null>(null);
	const { enqueueSnackbar } = useSnackbar();

	// Validate namespace name - only alphanumeric, underscores, and hyphens allowed
	const validateNamespaceName = (name: string): string => {
		if (!name.trim()) {
			return 'Name is required';
		}

		// Check for invalid characters (anything that's not alphanumeric, underscore, or hyphen)
		const invalidChars = /[^a-zA-Z0-9_-]/;
		if (invalidChars.test(name)) {
			return 'Name can only contain alphanumeric characters, underscores, and hyphens';
		}

		return '';
	};

	// Handle name change with validation
	const handleNameChange = (name: string) => {
		setFormData(prev => ({ ...prev, name }));
		const error = validateNamespaceName(name);
		setNameError(error);
	};

	// Fetch namespaces and servers on mount
	useEffect(() => {
		fetchNamespaces();
		fetchServers();
	}, []);

	// Check for action query parameter to open create modal
	useEffect(() => {
		const action = searchParams.get('action');
		if (action === 'create') {
			handleCreateNamespace();
		}
	}, [searchParams]);

	const fetchNamespaces = useCallback(async () => {
		setIsLoading(true);
		try {
			const data = await namespaceApi.listNamespaces();
			setNamespaces(data);
		} catch (error) {
			enqueueSnackbar('Failed to fetch namespaces', { variant: 'error' });
			console.error('Error fetching namespaces:', error);
			setNamespaces([]); // Set empty array on error
		} finally {
			setIsLoading(false);
		}
	}, [enqueueSnackbar]);

	const fetchServers = useCallback(async () => {
		setLoadingServers(true);
		try {
			const data = await serverApi.listServers();
			setServers(data);
		} catch (error) {
			enqueueSnackbar('Failed to fetch servers', { variant: 'error' });
			console.error('Error fetching servers:', error);
			setServers([]);
		} finally {
			setLoadingServers(false);
		}
	}, [enqueueSnackbar]);

	const handleCreateNamespace = () => {
		setFormData({ name: '', description: '', is_active: true, servers: [] });
		setEditingNamespace(null);
		setNameError('');
		setCreateModalOpen(true);
	};

	const handleViewNamespace = async (namespace: Namespace) => {
		setViewingNamespace(namespace);
		setModalTab(0);
		setViewModalOpen(true);
		setLoadingNamespaceDetails(true);

		// Fetch full namespace details including servers
		try {
			const fullNamespace = await namespaceApi.getNamespace(namespace.id);
			setViewingNamespace(fullNamespace);
			// Extract server IDs from the servers array (which may be objects or strings)
			const serverIds = fullNamespace.servers
				? fullNamespace.servers.map((s: any) => typeof s === 'string' ? s : s.server_id)
				: [];
			setFormData({
				name: fullNamespace.name,
				description: fullNamespace.description || '',
				is_active: fullNamespace.is_active,
				servers: serverIds
			});
		} catch (error) {
			enqueueSnackbar('Failed to fetch namespace details', { variant: 'error' });
			console.error('Error fetching namespace details:', error);
		} finally {
			setLoadingNamespaceDetails(false);
		}
	};

	const handleEditNamespace = async (namespace: Namespace) => {
		setEditingNamespace(namespace);
		setCreateModalOpen(true);
		setLoadingNamespaceDetails(true);
		setNameError('');

		// Set initial data from list (servers might not be populated yet)
		setFormData({
			name: namespace.name,
			description: namespace.description || '',
			is_active: namespace.is_active,
			servers: []
		});

		// Fetch full namespace details including servers
		try {
			const fullNamespace = await namespaceApi.getNamespace(namespace.id);
			setEditingNamespace(fullNamespace);
			// Extract server IDs from the servers array (which may be objects or strings)
			const serverIds = fullNamespace.servers
				? fullNamespace.servers.map((s: any) => typeof s === 'string' ? s : s.server_id)
				: [];
			setFormData({
				name: fullNamespace.name,
				description: fullNamespace.description || '',
				is_active: fullNamespace.is_active,
				servers: serverIds
			});
		} catch (error) {
			enqueueSnackbar('Failed to fetch namespace details', { variant: 'error' });
			console.error('Error fetching namespace details:', error);
		} finally {
			setLoadingNamespaceDetails(false);
		}
	};

	const handleSaveNamespace = async () => {
		// Validate name before saving
		const error = validateNamespaceName(formData.name);
		if (error) {
			setNameError(error);
			return;
		}

		try {
			if (editingNamespace) {
				// For updates, use server_ids field
				const updateData = {
					...formData,
					server_ids: formData.servers
				};
				const { servers: _, ...finalUpdateData } = updateData;
				await namespaceApi.updateNamespace(editingNamespace.id, finalUpdateData);
				enqueueSnackbar('Namespace updated successfully', { variant: 'success' });
			} else {
				// For creation, servers field is correct
				await namespaceApi.createNamespace(formData);
				enqueueSnackbar('Namespace created successfully', { variant: 'success' });
			}

			setCreateModalOpen(false);
			setEditingNamespace(null);
			fetchNamespaces(); // Refresh the list
		} catch (error) {
			const action = editingNamespace ? 'update' : 'create';
			enqueueSnackbar(`Failed to ${action} namespace`, { variant: 'error' });
			console.error(`Error ${action}ing namespace:`, error);
		}
	};

	const handleCloseViewModal = () => {
		setViewModalOpen(false);
		setViewingNamespace(null);
		setModalTab(0);
	};

	const handleUpdateFromView = async () => {
		if (!viewingNamespace) return;

		// Validate name before saving
		const error = validateNamespaceName(formData.name);
		if (error) {
			setNameError(error);
			return;
		}

		try {
			const { servers: _, ...updateData } = {
				...formData,
				server_ids: formData.servers
			};
			await namespaceApi.updateNamespace(viewingNamespace.id, updateData);
			enqueueSnackbar('Namespace updated successfully', { variant: 'success' });
			handleCloseViewModal();
			fetchNamespaces();
		} catch (error) {
			enqueueSnackbar('Failed to update namespace', { variant: 'error' });
			console.error('Error updating namespace:', error);
		}
	};

	const handleDeleteNamespace = async (namespace: Namespace) => {
		try {
			await namespaceApi.deleteNamespace(namespace.id);
			enqueueSnackbar(`Namespace ${namespace.name} deleted successfully`, { variant: 'success' });
			fetchNamespaces(); // Refresh the list
		} catch (error) {
			enqueueSnackbar(`Failed to delete namespace ${namespace.name}`, { variant: 'error' });
			console.error('Error deleting namespace:', error);
		}
	};

	const handleToggleNamespaceStatus = async (namespace: Namespace) => {
		setTogglingNamespaceId(namespace.id);
		try {
			const updatedNamespace = await namespaceApi.updateNamespace(namespace.id, {
				is_active: !namespace.is_active
			});

			// Update the local state with the updated namespace
			setNamespaces(prev => prev.map(ns =>
				ns.id === namespace.id ? updatedNamespace : ns
			));

			// Show specific toast based on activation/deactivation
			if (!namespace.is_active) {
				enqueueSnackbar('Namespace activated successfully', { variant: 'success' });
			} else {
				enqueueSnackbar('Namespace deactivated successfully', { variant: 'success' });
			}
		} catch (error) {
			const message = error instanceof Error ? error.message : 'Failed to update namespace status';
			enqueueSnackbar(message, { variant: 'error' });
			console.error('Error updating namespace status:', error);
			// Refresh data on error to ensure UI is in sync
			fetchNamespaces();
		} finally {
			setTogglingNamespaceId(null);
		}
	};

	const columns = useMemo<MRT_ColumnDef<Namespace>[]>(
		() => [
			{
				accessorKey: 'is_active',
				header: 'Active',
				size: 80,
				Cell: ({ row }) => (
					<Tooltip title={row.original.is_active ? 'Deactivate namespace' : 'Activate namespace'}>
						<Switch
							checked={!!row.original.is_active}
							onChange={() => handleToggleNamespaceStatus(row.original)}
							disabled={togglingNamespaceId === row.original.id}
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
						<SvgIcon size={20}>lucide:folder</SvgIcon>
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
								{row.original.id}
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
				accessorKey: 'server_count',
				header: 'Servers',
				size: 100,
				Cell: ({ cell }) => (
					<Chip
						size="small"
						label={cell.getValue<number>()}
						variant="outlined"
					/>
				)
			},
			{
				accessorKey: 'created_at',
				header: 'Created',
				size: 150,
				Cell: ({ cell }) => {
					const date = new Date(cell.getValue<string>());
					return date.toLocaleDateString('en-US', {
						year: 'numeric',
						month: 'short',
						day: 'numeric'
					});
				}
			}
		],
		[togglingNamespaceId]
	);

	return (
		<Root
			header={
				<div className="p-6">
					<div className="flex items-center justify-between">
						<div>
							<Typography variant="h4">Namespaces</Typography>
							<Typography
								variant="body1"
								color="textSecondary"
								className="mt-1"
							>
								Group and organize your MCP servers into logical namespaces
							</Typography>
						</div>
						<Button
							variant="contained"
							color="primary"
							startIcon={<SvgIcon>lucide:plus</SvgIcon>}
							onClick={handleCreateNamespace}
						>
							Create Namespace
						</Button>
					</div>
				</div>
			}
			content={
				<div className="p-6">
					<LazyDataTable
						columns={columns}
						data={namespaces}
						state={{ isLoading }}
						enableRowActions
						renderRowActions={({ row }) => (
							<Box className="flex items-center space-x-1">
								<Tooltip title="View Details">
									<IconButton
										size="small"
										onClick={() => handleViewNamespace(row.original)}
									>
										<SvgIcon size={18}>lucide:eye</SvgIcon>
									</IconButton>
								</Tooltip>
								<Tooltip title="Edit Namespace">
									<IconButton
										size="small"
										onClick={() => handleEditNamespace(row.original)}
									>
										<SvgIcon size={18}>lucide:pencil</SvgIcon>
									</IconButton>
								</Tooltip>
								<Tooltip title="Delete Namespace">
									<IconButton
										size="small"
										color="error"
										onClick={() => handleDeleteNamespace(row.original)}
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

					{/* Create/Edit Namespace Dialog */}
					<Dialog
						open={createModalOpen}
						onClose={() => setCreateModalOpen(false)}
						maxWidth="sm"
						fullWidth
					>
						<DialogTitle>{editingNamespace ? 'Edit Namespace' : 'Create Namespace'}</DialogTitle>
						<DialogContent>
							<Stack
								spacing={3}
								sx={{ mt: 1 }}
							>
								<TextField
									label="Name"
									value={formData.name}
									onChange={(e) => handleNameChange(e.target.value)}
									fullWidth
									required
									error={!!nameError}
									helperText={nameError || 'Only alphanumeric characters, underscores, and hyphens are allowed'}
								/>
								<TextField
									label="Description"
									value={formData.description}
									onChange={(e) => setFormData((prev) => ({ ...prev, description: e.target.value }))}
									fullWidth
									multiline
									rows={3}
								/>
								<Box>
									<Typography variant="body2" color="textSecondary" gutterBottom>
										Select Servers
									</Typography>
									{loadingServers || loadingNamespaceDetails ? (
										<Box sx={{ display: 'flex', alignItems: 'center', gap: 1, p: 2 }}>
											<CircularProgress size={20} />
											<Typography
												variant="caption"
												color="textSecondary"
											>
												Loading servers...
											</Typography>
										</Box>
									) : (
										<Box
											sx={{
												maxHeight: 200,
												overflowY: 'auto',
												border: 1,
												borderColor: 'divider',
												borderRadius: 1,
												p: 1
											}}
										>
											{servers.length === 0 ? (
												<Typography variant="body2" color="textSecondary" sx={{ p: 1 }}>
													No servers available
												</Typography>
											) : (
												servers.map((server) => (
													<FormControlLabel
														key={server.id}
														control={
															<Checkbox
																checked={formData.servers.includes(server.id)}
																onChange={(e) => {
																	if (e.target.checked) {
																		setFormData((prev) => ({
																			...prev,
																			servers: [...prev.servers, server.id]
																		}));
																	} else {
																		setFormData((prev) => ({
																			...prev,
																			servers: prev.servers.filter((id) => id !== server.id)
																		}));
																	}
																}}
															/>
														}
														label={
															<Box>
																<Typography variant="body2" component="div">
																	{server.name}
																</Typography>
																{/* {server.description && (
																	<Typography variant="caption" color="textSecondary">
																		{server.description}
																	</Typography>
																)} */}
															</Box>
														}
														sx={{ alignItems: 'flex-start', width: '100%', ml: 0 }}
													/>
												))
											)}
										</Box>
									)}
									{formData.servers.length > 0 && (
										<Box sx={{ mt: 2 }}>
											<Typography variant="body2" color="textSecondary" gutterBottom>
												Selected ({formData.servers.length})
											</Typography>
											<Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
												{formData.servers.map((serverId) => {
													const server = servers.find((s) => s.id === serverId);
													return (
														<Chip
															key={serverId}
															label={server?.name || 'Unknown'}
															size="small"
															variant="outlined"
															onDelete={() => {
																setFormData((prev) => ({
																	...prev,
																	servers: prev.servers.filter((id) => id !== serverId)
																}));
															}}
														/>
													);
												})}
											</Stack>
										</Box>
									)}
								</Box>
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
								onClick={handleSaveNamespace}
								disabled={!formData.name.trim() || !!nameError}
							>
								{editingNamespace ? 'Update' : 'Create'}
							</Button>
						</DialogActions>
					</Dialog>

					{/* View/Edit Namespace Modal */}
					<Dialog
						open={viewModalOpen}
						onClose={handleCloseViewModal}
						maxWidth="md"
						fullWidth
					>
						<DialogTitle>
							<Box className="flex items-center justify-between">
								<Typography variant="h6">
									{viewingNamespace?.name || 'Namespace Details'}
								</Typography>
								<Chip
									size="small"
									label={viewingNamespace?.is_active ? 'Active' : 'Inactive'}
									color={viewingNamespace?.is_active ? 'success' : 'default'}
								/>
							</Box>
						</DialogTitle>
						<DialogContent>
							<Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 3 }}>
								<Tabs value={modalTab} onChange={(e, newValue) => setModalTab(newValue)}>
									<Tab label="Overview" />
									<Tab label="Settings" />
									<Tab label="Servers" />
									<Tab label="Endpoint" />
								</Tabs>
							</Box>

							{/* Overview Tab */}
							{modalTab === 0 && (
								<Stack spacing={3}>
									<Card variant="outlined">
										<CardContent>
											<Typography variant="h6" gutterBottom>
												Namespace Information
											</Typography>
											<Grid container spacing={2}>
												<Grid item xs={6}>
													<Typography variant="body2" color="textSecondary">
														Name
													</Typography>
													<Typography variant="body1">
														{viewingNamespace?.name}
													</Typography>
												</Grid>
												<Grid item xs={6}>
													<Typography variant="body2" color="textSecondary">
														ID
													</Typography>
													<Typography variant="body1">
														{viewingNamespace?.id}
													</Typography>
												</Grid>
												<Grid item xs={6}>
													<Typography variant="body2" color="textSecondary">
														Server Count
													</Typography>
													<Typography variant="body1">
														{viewingNamespace?.server_count || 0} servers
													</Typography>
												</Grid>
												<Grid item xs={6}>
													<Typography variant="body2" color="textSecondary">
														Created
													</Typography>
													<Typography variant="body1">
														{viewingNamespace?.created_at ? new Date(viewingNamespace.created_at).toLocaleDateString('en-US', {
															year: 'numeric',
															month: 'short',
															day: 'numeric'
														}) : 'N/A'}
													</Typography>
												</Grid>
												<Grid item xs={12}>
													<Typography variant="body2" color="textSecondary">
														Description
													</Typography>
													<Typography variant="body1">
														{viewingNamespace?.description || 'No description provided'}
													</Typography>
												</Grid>
											</Grid>
										</CardContent>
									</Card>
								</Stack>
							)}

							{/* Settings Tab */}
							{modalTab === 1 && (
								<Stack spacing={3}>
									<Alert severity="info">
										Modify namespace settings. Changes will be saved when you click Update.
									</Alert>
									<TextField
										label="Name"
										value={formData.name}
										onChange={(e) => handleNameChange(e.target.value)}
										fullWidth
										required
										error={!!nameError}
										helperText={nameError || 'Only alphanumeric characters, underscores, and hyphens are allowed'}
									/>
									<TextField
										label="Description"
										value={formData.description}
										onChange={(e) => setFormData((prev) => ({ ...prev, description: e.target.value }))}
										fullWidth
										multiline
										rows={3}
										helperText="Optional description to help identify this namespace"
									/>
									<FormControlLabel
										control={
											<Switch
												checked={formData.is_active}
												onChange={(e) =>
													setFormData((prev) => ({ ...prev, is_active: e.target.checked }))
												}
												color={formData.is_active ? 'success' : 'default'}
											/>
										}
										label={`${formData.is_active ? 'Active' : 'Inactive'}`}
									/>
									{!formData.is_active && (
										<Alert severity="warning">
											Deactivating this namespace will make it unavailable for use, but data will be preserved.
										</Alert>
									)}
								</Stack>
							)}

							{/* Servers Tab */}
							{modalTab === 2 && (
								<Stack spacing={3}>
									<Alert severity="info">
										Manage which servers are associated with this namespace.
									</Alert>
									<Box>
										<Typography variant="body2" color="textSecondary" gutterBottom>
											Select Servers
										</Typography>
										{loadingServers || loadingNamespaceDetails ? (
											<Box sx={{ display: 'flex', alignItems: 'center', gap: 1, p: 2 }}>
												<CircularProgress size={20} />
												<Typography
													variant="caption"
													color="textSecondary"
												>
													Loading servers...
												</Typography>
											</Box>
										) : (
											<Box
												sx={{
													maxHeight: 200,
													overflowY: 'auto',
													border: 1,
													borderColor: 'divider',
													borderRadius: 1,
													p: 1
												}}
											>
												{servers.length === 0 ? (
													<Typography variant="body2" color="textSecondary" sx={{ p: 1 }}>
														No servers available
													</Typography>
												) : (
													servers.map((server) => (
														<FormControlLabel
															key={server.id}
															control={
																<Checkbox
																	checked={formData.servers.includes(server.id)}
																	onChange={(e) => {
																		if (e.target.checked) {
																			setFormData((prev) => ({
																				...prev,
																				servers: [...prev.servers, server.id]
																			}));
																		} else {
																			setFormData((prev) => ({
																				...prev,
																				servers: prev.servers.filter((id) => id !== server.id)
																			}));
																		}
																	}}
																/>
															}
															label={
																<Box>
																	<Typography variant="body2" component="div">
																		{server.name}
																	</Typography>
																	{server.description && (
																		<Typography variant="caption" color="textSecondary">
																			{server.description}
																		</Typography>
																	)}
																</Box>
															}
															sx={{ alignItems: 'flex-start', width: '100%', ml: 0 }}
														/>
													))
												)}
											</Box>
										)}
									</Box>
									{formData.servers.length > 0 && (
										<Box>
											<Typography variant="body2" color="textSecondary" gutterBottom>
												Selected Servers ({formData.servers.length})
											</Typography>
											<Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
												{formData.servers.map((serverId) => {
													const server = servers.find((s) => s.id === serverId);
													return (
														<Chip
															key={serverId}
															label={server?.name || 'Unknown'}
															size="small"
															variant="outlined"
															onDelete={() => {
																setFormData((prev) => ({
																	...prev,
																	servers: prev.servers.filter((id) => id !== serverId)
																}));
															}}
														/>
													);
												})}
											</Stack>
										</Box>
									)}
								</Stack>
							)}

							{/* Endpoint Tab */}
							{modalTab === 3 && (
								<Stack spacing={3}>
									<Alert severity="info">
										Configure public endpoint for this namespace to make it accessible via custom URLs.
									</Alert>

									{viewingNamespace?.endpoint ? (
										<Card variant="outlined">
											<CardContent>
												<Box sx={{ display: 'flex', justifyContent: 'between', alignItems: 'center', mb: 2 }}>
													<Typography
														variant="h6"
														sx={{
															cursor: 'pointer',
															color: 'primary.main',
															'&:hover': { textDecoration: 'underline' }
														}}
														onClick={() => {
															if (viewingNamespace?.endpoint) {
																handleCloseViewModal();
																router.push(`/endpoints?highlight=${viewingNamespace.endpoint.id}`);
															}
														}}
													>
														Endpoint Configuration
														<SvgIcon size={16} sx={{ ml: 1 }}>lucide:external-link</SvgIcon>
													</Typography>
													<Chip
														size="small"
														label={viewingNamespace.endpoint.is_active ? 'Active' : 'Inactive'}
														color={viewingNamespace.endpoint.is_active ? 'success' : 'default'}
														variant="outlined"
													/>
												</Box>

												<Grid container spacing={2}>
													<Grid item xs={12}>
														<Typography variant="body2" color="textSecondary">
															Endpoint Name
														</Typography>
														<Typography variant="body1" sx={{ fontFamily: 'monospace', fontSize: '0.875rem' }}>
															{viewingNamespace.endpoint.name}
														</Typography>
													</Grid>

													{viewingNamespace.endpoint.description && (
														<Grid item xs={12}>
															<Typography variant="body2" color="textSecondary">
																Description
															</Typography>
															<Typography variant="body1">
																{viewingNamespace.endpoint.description}
															</Typography>
														</Grid>
													)}

													{viewingNamespace.endpoint.urls && (
														<Grid item xs={12}>
															<Typography variant="body2" color="textSecondary" gutterBottom>
																Endpoint URLs
															</Typography>
															<Stack spacing={1}>
																<Box>
																	<Typography variant="caption" color="textSecondary">
																		SSE:
																	</Typography>
																	<Typography variant="body2" sx={{ fontFamily: 'monospace', ml: 1 }}>
																		{viewingNamespace.endpoint.urls.sse}
																	</Typography>
																</Box>
																<Box>
																	<Typography variant="caption" color="textSecondary">
																		HTTP:
																	</Typography>
																	<Typography variant="body2" sx={{ fontFamily: 'monospace', ml: 1 }}>
																		{viewingNamespace.endpoint.urls.http}
																	</Typography>
																</Box>
																<Box>
																	<Typography variant="caption" color="textSecondary">
																		WebSocket:
																	</Typography>
																	<Typography variant="body2" sx={{ fontFamily: 'monospace', ml: 1 }}>
																		{viewingNamespace.endpoint.urls.websocket}
																	</Typography>
																</Box>
															</Stack>
														</Grid>
													)}

													<Grid item xs={6}>
														<Typography variant="body2" color="textSecondary">
															Rate Limit
														</Typography>
														<Typography variant="body1">
															{viewingNamespace.endpoint.rate_limit_requests} requests per {viewingNamespace.endpoint.rate_limit_window}s
														</Typography>
													</Grid>

													<Grid item xs={6}>
														<Typography variant="body2" color="textSecondary">
															Authentication Methods
														</Typography>
														<Stack direction="row" spacing={1} sx={{ mt: 0.5 }}>
															{viewingNamespace.endpoint.enable_api_key_auth && (
																<Chip label="API Key" size="small" variant="outlined" />
															)}
															{viewingNamespace.endpoint.enable_oauth && (
																<Chip label="OAuth" size="small" variant="outlined" />
															)}
															{viewingNamespace.endpoint.enable_public_access && (
																<Chip label="Public Access" size="small" variant="outlined" color="warning" />
															)}
															{!viewingNamespace.endpoint.enable_api_key_auth && !viewingNamespace.endpoint.enable_oauth && !viewingNamespace.endpoint.enable_public_access && (
																<Typography variant="caption" color="textSecondary">
																	No authentication configured
																</Typography>
															)}
														</Stack>
													</Grid>
												</Grid>
											</CardContent>
										</Card>
									) : (
										<Card variant="outlined">
											<CardContent sx={{ textAlign: 'center', py: 4 }}>
												<Typography variant="h6" gutterBottom>
													No Endpoint Configured
												</Typography>
												<Typography variant="body2" color="textSecondary" gutterBottom={true}>
													Create a public endpoint to make this namespace accessible via custom URLs with authentication and rate limiting.
												</Typography>
												<Button
													variant="contained"
													size="medium"
													startIcon={<SvgIcon size={16}>lucide:plus</SvgIcon>}
													onClick={() => {
														handleCloseViewModal();
														router.push(`/endpoints?action=create&namespace_id=${viewingNamespace?.id}`);
													}}
													sx={{ mt: 4 }}
												>
													Create Endpoint
												</Button>
											</CardContent>
										</Card>
									)}
								</Stack>
							)}
						</DialogContent>
						<DialogActions>
							<Button onClick={handleCloseViewModal}>Cancel</Button>
							<Button
								variant="contained"
								onClick={handleUpdateFromView}
								disabled={!formData.name.trim() || !!nameError}
							>
								Update
							</Button>
						</DialogActions>
					</Dialog>
				</div>
			}
		/>
	);
}

export default NamespacesView;
