'use client';

import { useState, useMemo, useEffect, useCallback } from 'react';
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
	Tabs
} from '@mui/material';
import LazyDataTable from '@/components/data-table/LazyDataTable';
import SvgIcon from '@fuse/core/SvgIcon';
import { useSnackbar } from 'notistack';
import { namespaceApi, serverApi, type MCPServer, type Namespace } from '@/lib/api';

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
	const [namespaces, setNamespaces] = useState<Namespace[]>([]);
	const [isLoading, setIsLoading] = useState(false);
	const [servers, setServers] = useState<MCPServer[]>([]);
	const [loadingServers, setLoadingServers] = useState(false);
	const [togglingNamespaceId, setTogglingNamespaceId] = useState<string | null>(null);
	const { enqueueSnackbar } = useSnackbar();

	// Fetch namespaces and servers on mount
	useEffect(() => {
		fetchNamespaces();
		fetchServers();
	}, []);

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
		setCreateModalOpen(true);
	};

	const handleViewNamespace = (namespace: Namespace) => {
		setViewingNamespace(namespace);
		setFormData({
			name: namespace.name,
			description: namespace.description || '',
			is_active: namespace.is_active,
			servers: namespace.servers || []
		});
		setModalTab(0);
		setViewModalOpen(true);
	};

	const handleEditNamespace = (namespace: Namespace) => {
		setFormData({
			name: namespace.name,
			description: namespace.description || '',
			is_active: namespace.is_active,
			servers: namespace.servers || []
		});
		setEditingNamespace(namespace);
		setCreateModalOpen(true);
	};

	const handleSaveNamespace = async () => {
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
										<SvgIcon size={18}>lucide:edit</SvgIcon>
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
									onChange={(e) => setFormData((prev) => ({ ...prev, name: e.target.value }))}
									fullWidth
									required
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
									{loadingServers ? (
										<Typography
											variant="caption"
											color="textSecondary"
										>
											Loading servers...
										</Typography>
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
								disabled={!formData.name.trim()}
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
										onChange={(e) => setFormData((prev) => ({ ...prev, name: e.target.value }))}
										fullWidth
										required
										helperText="The display name for this namespace"
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
										label={`Namespace is ${formData.is_active ? 'Active' : 'Inactive'}`}
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
										{loadingServers ? (
											<Typography
												variant="caption"
												color="textSecondary"
											>
												Loading servers...
											</Typography>
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
						</DialogContent>
						<DialogActions>
							<Button onClick={handleCloseViewModal}>Cancel</Button>
							<Button
								variant="contained"
								onClick={handleUpdateFromView}
								disabled={!formData.name.trim()}
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
