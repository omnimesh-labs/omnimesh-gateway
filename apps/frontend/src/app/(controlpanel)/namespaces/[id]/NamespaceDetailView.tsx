'use client';

import { useState, useEffect, useMemo } from 'react';
import { useParams, useRouter } from 'next/navigation';
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
	Card,
	CardContent,
	GridLegacy as Grid,
	Tabs,
	Tab,
	List,
	ListItem,
	ListItemText,
	ListItemIcon,
	Divider,
	Dialog,
	DialogTitle,
	DialogContent,
	DialogActions,
	TextField,
	FormControl,
	InputLabel,
	Select,
	MenuItem,
	Alert,
	CircularProgress,
	Breadcrumbs,
	Link,
	Stack,
	FormControlLabel,
	Switch,
	Accordion,
	AccordionSummary,
	AccordionDetails
} from '@mui/material';
import LazyDataTable from '@/components/data-table/LazyDataTable';
import SvgIcon from '@fuse/core/SvgIcon';
import { useSnackbar } from 'notistack';
import { namespaceApi, serverApi, endpointApi } from '@/lib/client-api';
import type { MCPServer, Namespace, Endpoint } from '@/lib/types';

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

interface NamespaceDetails extends Omit<Namespace, 'servers'> {
	servers?: MCPServer[];
	endpoints?: Endpoint[];
	stats?: {
		total_servers: number;
		active_servers: number;
		total_sessions: number;
		total_requests: number;
	};
}

function NamespaceDetailView() {
	const params = useParams();
	const router = useRouter();
	const namespaceId = params.id as string;
	const { enqueueSnackbar } = useSnackbar();

	const [namespace, setNamespace] = useState<NamespaceDetails | null>(null);
	const [loading, setLoading] = useState(true);
	const [tabValue, setTabValue] = useState(0);
	const [editDialogOpen, setEditDialogOpen] = useState(false);
	const [assignServerDialogOpen, setAssignServerDialogOpen] = useState(false);
	const [availableServers, setAvailableServers] = useState<MCPServer[]>([]);
	const [selectedServers, setSelectedServers] = useState<string[]>([]);
	const [editFormData, setEditFormData] = useState({
		name: '',
		description: ''
	});
	const [editEndpointModalOpen, setEditEndpointModalOpen] = useState(false);
	const [editingEndpoint, setEditingEndpoint] = useState<Endpoint | null>(null);
	const [endpointFormData, setEndpointFormData] = useState({
		name: '',
		description: '',
		enable_api_key_auth: true,
		enable_oauth: false,
		enable_public_access: false,
		rate_limit_requests: 100,
		rate_limit_window: 3600,
		is_active: true
	});

	const loadNamespaceDetails = async () => {
		try {
			setLoading(true);
			const [namespaceData, allServers, allEndpoints] = await Promise.all([
				namespaceApi.getNamespace(namespaceId),
				serverApi.listServers(),
				endpointApi.listEndpoints()
			]);

			// Get servers assigned to this namespace (assuming they're returned with the namespace data)
			const assignedServerIds = namespaceData.servers || [];
			const assignedServers = allServers.filter((s) => assignedServerIds.includes(s.id));

			// Get endpoints for this namespace
			const namespaceEndpoints = allEndpoints.filter(e => e.namespace_id === namespaceId);

			// Calculate stats
			const stats = {
				total_servers: assignedServers.length,
				active_servers: assignedServers.filter((s) => s.is_active).length,
				total_sessions: 0, // Would come from a sessions API
				total_requests: 0 // Would come from metrics API
			};

			setNamespace({
				...namespaceData,
				servers: assignedServers,
				endpoints: namespaceEndpoints,
				stats
			});

			// Available servers are those not assigned to any namespace
			const unassignedServers = allServers.filter((s) => !assignedServerIds.includes(s.id));
			setAvailableServers(unassignedServers);
		} catch (error) {
			console.error('Failed to load namespace:', error);
			enqueueSnackbar('Failed to load namespace details', { variant: 'error' });
			router.push('/namespaces');
		} finally {
			setLoading(false);
		}
	};

	useEffect(() => {
		loadNamespaceDetails();
	}, [namespaceId]);

	const handleEditNamespace = () => {
		if (!namespace) return;

		setEditFormData({
			name: namespace.name,
			description: namespace.description || ''
		});
		setEditDialogOpen(true);
	};

	const handleSaveNamespace = async () => {
		try {
			await namespaceApi.updateNamespace(namespaceId, editFormData);
			enqueueSnackbar('Namespace updated successfully', { variant: 'success' });
			setEditDialogOpen(false);
			loadNamespaceDetails();
		} catch (error) {
			enqueueSnackbar('Failed to update namespace', { variant: 'error' });
		}
	};

	const handleDeleteNamespace = async () => {
		if (!confirm('Are you sure you want to delete this namespace? This action cannot be undone.')) {
			return;
		}

		try {
			await namespaceApi.deleteNamespace(namespaceId);
			enqueueSnackbar('Namespace deleted successfully', { variant: 'success' });
			router.push('/namespaces');
		} catch (_error) {
			enqueueSnackbar('Failed to delete namespace', { variant: 'error' });
		}
	};

	const handleAssignServers = async () => {
		try {
			// Assign each selected server to the namespace
			await Promise.all(
				selectedServers.map((serverId) => namespaceApi.addServerToNamespace(namespaceId, serverId))
			);
			enqueueSnackbar('Servers assigned successfully', { variant: 'success' });
			setAssignServerDialogOpen(false);
			setSelectedServers([]);
			loadNamespaceDetails();
		} catch (error) {
			enqueueSnackbar('Failed to assign servers', { variant: 'error' });
		}
	};

	const handleRemoveServer = async (serverId: string) => {
		if (!confirm('Are you sure you want to remove this server from the namespace?')) {
			return;
		}

		try {
			await namespaceApi.removeServerFromNamespace(namespaceId, serverId);
			enqueueSnackbar('Server removed successfully', { variant: 'success' });
			loadNamespaceDetails();
		} catch (_error) {
			enqueueSnackbar('Failed to remove server', { variant: 'error' });
		}
	};

	const handleEditEndpoint = (endpoint: Endpoint) => {
		setEndpointFormData({
			name: endpoint.name,
			description: endpoint.description || '',
			enable_api_key_auth: endpoint.enable_api_key_auth,
			enable_oauth: endpoint.enable_oauth,
			enable_public_access: endpoint.enable_public_access,
			rate_limit_requests: endpoint.rate_limit_requests,
			rate_limit_window: endpoint.rate_limit_window,
			is_active: endpoint.is_active
		});
		setEditingEndpoint(endpoint);
		setEditEndpointModalOpen(true);
	};

	const handleSaveEndpoint = async () => {
		if (!editingEndpoint) return;

		try {
			await endpointApi.updateEndpoint(editingEndpoint.id, endpointFormData);
			enqueueSnackbar('Endpoint updated successfully', { variant: 'success' });
			setEditEndpointModalOpen(false);
			setEditingEndpoint(null);
			loadNamespaceDetails(); // Refresh the namespace data
		} catch (error) {
			enqueueSnackbar('Failed to update endpoint', { variant: 'error' });
			console.error('Error updating endpoint:', error);
		}
	};

	const handleDeleteEndpoint = async (endpoint: Endpoint) => {
		if (!confirm(`Are you sure you want to delete endpoint "${endpoint.name}"?`)) {
			return;
		}

		try {
			await endpointApi.deleteEndpoint(endpoint.id);
			enqueueSnackbar(`Endpoint ${endpoint.name} deleted successfully`, { variant: 'success' });
			loadNamespaceDetails(); // Refresh the namespace data
		} catch (error) {
			enqueueSnackbar(`Failed to delete endpoint ${endpoint.name}`, { variant: 'error' });
			console.error('Error deleting endpoint:', error);
		}
	};

	const endpointColumns = useMemo<MRT_ColumnDef<Endpoint>[]>(
		() => [
			{
				accessorKey: 'name',
				header: 'Endpoint Name',
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
							{row.original.description && (
								<Typography
									variant="caption"
									color="textSecondary"
								>
									{row.original.description}
								</Typography>
							)}
						</Box>
					</Box>
				)
			},
			{
				accessorKey: 'enable_api_key_auth',
				header: 'Auth Type',
				size: 150,
				Cell: ({ row }) => {
					const endpoint = row.original;
					const authTypes = [];
					if (endpoint.enable_api_key_auth) authTypes.push('API Key');
					if (endpoint.enable_oauth) authTypes.push('OAuth');
					if (endpoint.enable_public_access) authTypes.push('Public');

					return (
						<Box className="flex flex-wrap gap-1">
							{authTypes.map((type) => (
								<Chip
									key={type}
									size="small"
									label={type}
									variant="outlined"
								/>
							))}
						</Box>
					);
				}
			},
			{
				accessorKey: 'rate_limit_requests',
				header: 'Rate Limit',
				size: 120,
				Cell: ({ row }) => {
					const endpoint = row.original;
					return (
						<Typography variant="body2">
							{endpoint.rate_limit_requests}/{endpoint.rate_limit_window}s
						</Typography>
					);
				}
			},
			{
				accessorKey: 'is_active',
				header: 'Status',
				size: 100,
				Cell: ({ cell }) => {
					const isActive = cell.getValue<boolean>();
					return (
						<Chip
							size="small"
							label={isActive ? 'Active' : 'Inactive'}
							color={isActive ? 'success' : 'default'}
						/>
					);
				}
			}
		],
		[]
	);

	const serverColumns = useMemo<MRT_ColumnDef<MCPServer>[]>(
		() => [
			{
				accessorKey: 'name',
				header: 'Server Name',
				size: 200,
				Cell: ({ row }) => (
					<Box className="flex items-center space-x-2">
						<SvgIcon size={20}>lucide:server</SvgIcon>
						<Box>
							<Typography
								variant="body2"
								className="font-medium"
							>
								{row.original.name}
							</Typography>
							{row.original.description && (
								<Typography
									variant="caption"
									color="textSecondary"
								>
									{row.original.description}
								</Typography>
							)}
						</Box>
					</Box>
				)
			},
			{
				accessorKey: 'protocol',
				header: 'Protocol',
				size: 100,
				Cell: ({ cell }) => (
					<Chip
						size="small"
						label={cell.getValue<string>()}
						variant="outlined"
					/>
				)
			},
			{
				accessorKey: 'status',
				header: 'Status',
				size: 100,
				Cell: ({ cell }) => {
					const status = cell.getValue<string>();
					const color =
						status === 'active'
							? 'success'
							: status === 'inactive'
								? 'default'
								: status === 'unhealthy'
									? 'error'
									: 'warning';
					return (
						<Chip
							size="small"
							label={status}
							color={
								color as 'default' | 'primary' | 'secondary' | 'error' | 'info' | 'success' | 'warning'
							}
						/>
					);
				}
			},
			{
				accessorKey: 'version',
				header: 'Version',
				size: 100
			}
		],
		[]
	);

	if (loading) {
		return (
			<Box className="flex h-screen items-center justify-center">
				<CircularProgress />
			</Box>
		);
	}

	if (!namespace) {
		return (
			<Box className="p-6">
				<Alert severity="error">Namespace not found</Alert>
			</Box>
		);
	}

	return (
		<Root
			header={
				<div className="p-6">
					<Breadcrumbs className="mb-4">
						<Link
							color="inherit"
							href="/namespaces"
							onClick={(e) => {
								e.preventDefault();
								router.push('/namespaces');
							}}
							className="cursor-pointer hover:underline"
						>
							Namespaces
						</Link>
						<Typography color="text.primary">{namespace.name}</Typography>
					</Breadcrumbs>

					<div className="flex items-center justify-between">
						<div>
							<Typography variant="h4">{namespace.name}</Typography>
							<Typography
								variant="body1"
								color="textSecondary"
								className="mt-1"
							>
								{namespace.description || 'No description provided'}
							</Typography>
							<Box className="mt-2 flex gap-2">
								<Chip
									size="small"
									label={namespace.is_active ? 'Active' : 'Inactive'}
									color={namespace.is_active ? 'success' : 'default'}
								/>
								{namespace.metadata?.environment && (
									<Chip
										size="small"
										label={`Environment: ${namespace.metadata.environment}`}
										variant="outlined"
									/>
								)}
								{namespace.metadata?.region && (
									<Chip
										size="small"
										label={`Region: ${namespace.metadata.region}`}
										variant="outlined"
									/>
								)}
							</Box>
						</div>
						<Box className="flex gap-2">
							<Button
								variant="outlined"
								startIcon={<SvgIcon>lucide:edit</SvgIcon>}
								onClick={handleEditNamespace}
							>
								Edit
							</Button>
							<Button
								variant="outlined"
								color="error"
								startIcon={<SvgIcon>lucide:trash-2</SvgIcon>}
								onClick={handleDeleteNamespace}
							>
								Delete
							</Button>
						</Box>
					</div>
				</div>
			}
			content={
				<div className="p-6">
					{/* Stats Cards */}
					{namespace.stats && (
						<Grid
							container
							spacing={3}
							className="mb-6"
						>
							<Grid
								item
								xs={12}
								sm={6}
								md={3}
							>
								<Card>
									<CardContent>
										<Typography variant="h6">{namespace.stats.total_servers}</Typography>
										<Typography
											variant="body2"
											color="textSecondary"
										>
											Total Servers
										</Typography>
									</CardContent>
								</Card>
							</Grid>
							<Grid
								item
								xs={12}
								sm={6}
								md={3}
							>
								<Card>
									<CardContent>
										<Typography
											variant="h6"
											color="success.main"
										>
											{namespace.stats.active_servers}
										</Typography>
										<Typography
											variant="body2"
											color="textSecondary"
										>
											Active Servers
										</Typography>
									</CardContent>
								</Card>
							</Grid>
							<Grid
								item
								xs={12}
								sm={6}
								md={3}
							>
								<Card>
									<CardContent>
										<Typography variant="h6">{namespace.stats.total_sessions}</Typography>
										<Typography
											variant="body2"
											color="textSecondary"
										>
											Total Sessions
										</Typography>
									</CardContent>
								</Card>
							</Grid>
							<Grid
								item
								xs={12}
								sm={6}
								md={3}
							>
								<Card>
									<CardContent>
										<Typography variant="h6">{namespace.stats.total_requests}</Typography>
										<Typography
											variant="body2"
											color="textSecondary"
										>
											Total Requests
										</Typography>
									</CardContent>
								</Card>
							</Grid>
						</Grid>
					)}

					<Tabs
						value={tabValue}
						onChange={(_, newValue) => setTabValue(newValue)}
						className="mb-4"
					>
						<Tab label={`Servers (${namespace.servers?.length || 0})`} />
						<Tab label={`Endpoints (${namespace.endpoints?.length || 0})`} />
						<Tab label="Settings" />
						<Tab label="Activity" />
					</Tabs>

					{/* Servers Tab */}
					{tabValue === 0 && (
						<Box>
							<Box className="mb-4 flex items-center justify-between">
								<Typography variant="h6">Assigned Servers</Typography>
								<Button
									variant="contained"
									color="primary"
									startIcon={<SvgIcon>lucide:plus</SvgIcon>}
									onClick={() => setAssignServerDialogOpen(true)}
								>
									Assign Servers
								</Button>
							</Box>

							{namespace.servers && namespace.servers.length > 0 ? (
								<LazyDataTable
									columns={serverColumns}
									data={namespace.servers}
									enableRowActions
									renderRowActions={({ row }) => (
										<Box className="flex items-center space-x-1">
											<Tooltip title="View Details">
												<IconButton
													size="small"
													onClick={() => router.push(`/servers?id=${row.original.id}`)}
												>
													<SvgIcon size={18}>lucide:eye</SvgIcon>
												</IconButton>
											</Tooltip>
											<Tooltip title="Remove from Namespace">
												<IconButton
													size="small"
													color="error"
													onClick={() => handleRemoveServer(row.original.id)}
												>
													<SvgIcon size={18}>lucide:x</SvgIcon>
												</IconButton>
											</Tooltip>
										</Box>
									)}
								/>
							) : (
								<Alert severity="info">
									No servers assigned to this namespace. Click "Assign Servers" to add servers.
								</Alert>
							)}
						</Box>
					)}

					{/* Endpoints Tab */}
					{tabValue === 1 && (
						<Box>
							<Box className="mb-4 flex items-center justify-between">
								<Typography variant="h6">Endpoints</Typography>
								<Button
									variant="contained"
									color="primary"
									startIcon={<SvgIcon>lucide:plus</SvgIcon>}
									onClick={() => router.push(`/endpoints?action=create&namespace_id=${namespace.id}`)}
								>
									Create Endpoint
								</Button>
							</Box>

							{namespace.endpoints && namespace.endpoints.length > 0 ? (
								<LazyDataTable
									columns={endpointColumns}
									data={namespace.endpoints}
									enableRowActions
									renderRowActions={({ row }) => (
										<Box className="flex items-center space-x-1">
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
								/>
							) : (
								<Alert severity="info">
									No endpoints configured for this namespace. Click "Create Endpoint" to add your first endpoint.
								</Alert>
							)}
						</Box>
					)}

					{/* Settings Tab */}
					{tabValue === 2 && (
						<Card>
							<CardContent>
								<Typography
									variant="h6"
									className="mb-3"
								>
									Namespace Settings
								</Typography>
								<List>
									<ListItem>
										<ListItemIcon>
											<SvgIcon>lucide:hash</SvgIcon>
										</ListItemIcon>
										<ListItemText
											primary="Namespace ID"
											secondary={namespace.id}
										/>
									</ListItem>
									<Divider />
									<ListItem>
										<ListItemIcon>
											<SvgIcon>lucide:building</SvgIcon>
										</ListItemIcon>
										<ListItemText
											primary="Organization ID"
											secondary={namespace.organization_id}
										/>
									</ListItem>
									<Divider />
									<ListItem>
										<ListItemIcon>
											<SvgIcon>lucide:calendar</SvgIcon>
										</ListItemIcon>
										<ListItemText
											primary="Created"
											secondary={new Date(namespace.created_at).toLocaleString()}
										/>
									</ListItem>
									<Divider />
									<ListItem>
										<ListItemIcon>
											<SvgIcon>lucide:clock</SvgIcon>
										</ListItemIcon>
										<ListItemText
											primary="Last Updated"
											secondary={new Date(namespace.updated_at).toLocaleString()}
										/>
									</ListItem>
									{namespace.metadata &&
										Object.entries(namespace.metadata).map(([key, value]) => (
											<div key={key}>
												<Divider />
												<ListItem>
													<ListItemIcon>
														<SvgIcon>lucide:tag</SvgIcon>
													</ListItemIcon>
													<ListItemText
														primary={key.charAt(0).toUpperCase() + key.slice(1)}
														secondary={String(value)}
													/>
												</ListItem>
											</div>
										))}
								</List>
							</CardContent>
						</Card>
					)}

					{/* Activity Tab */}
					{tabValue === 3 && (
						<Card>
							<CardContent>
								<Typography
									variant="h6"
									className="mb-3"
								>
									Recent Activity
								</Typography>
								<Alert severity="info">Activity tracking will be available in a future update.</Alert>
							</CardContent>
						</Card>
					)}

					{/* Edit Namespace Dialog */}
					<Dialog
						open={editDialogOpen}
						onClose={() => setEditDialogOpen(false)}
						maxWidth="sm"
						fullWidth
					>
						<DialogTitle>Edit Namespace</DialogTitle>
						<DialogContent>
							<Box className="mt-2 space-y-4">
								<TextField
									label="Name"
									value={editFormData.name}
									onChange={(e) => setEditFormData({ ...editFormData, name: e.target.value })}
									fullWidth
									required
								/>
								<TextField
									label="Description"
									value={editFormData.description}
									onChange={(e) => setEditFormData({ ...editFormData, description: e.target.value })}
									fullWidth
									multiline
									rows={3}
								/>
							</Box>
						</DialogContent>
						<DialogActions>
							<Button onClick={() => setEditDialogOpen(false)}>Cancel</Button>
							<Button
								variant="contained"
								onClick={handleSaveNamespace}
							>
								Save Changes
							</Button>
						</DialogActions>
					</Dialog>

					{/* Assign Servers Dialog */}
					<Dialog
						open={assignServerDialogOpen}
						onClose={() => setAssignServerDialogOpen(false)}
						maxWidth="md"
						fullWidth
					>
						<DialogTitle>Assign Servers to Namespace</DialogTitle>
						<DialogContent>
							<Alert
								severity="info"
								className="mb-3"
							>
								Select servers to assign to this namespace. Only unassigned servers are shown.
							</Alert>
							<FormControl fullWidth>
								<InputLabel>Select Servers</InputLabel>
								<Select
									multiple
									value={selectedServers}
									onChange={(e) => setSelectedServers(e.target.value as string[])}
									renderValue={(selected) => `${selected.length} servers selected`}
								>
									{availableServers.map((server) => (
										<MenuItem
											key={server.id}
											value={server.id}
										>
											<Box>
												<Typography variant="body2">{server.name}</Typography>
												<Typography
													variant="caption"
													color="textSecondary"
												>
													{server.protocol} - {server.status}
												</Typography>
											</Box>
										</MenuItem>
									))}
								</Select>
							</FormControl>
						</DialogContent>
						<DialogActions>
							<Button onClick={() => setAssignServerDialogOpen(false)}>Cancel</Button>
							<Button
								variant="contained"
								onClick={handleAssignServers}
								disabled={selectedServers.length === 0}
							>
								Assign {selectedServers.length} Server(s)
							</Button>
						</DialogActions>
					</Dialog>

					{/* Edit Endpoint Dialog */}
					<Dialog
						open={editEndpointModalOpen}
						onClose={() => setEditEndpointModalOpen(false)}
						maxWidth="md"
						fullWidth
					>
						<DialogTitle>Edit Endpoint - {editingEndpoint?.name}</DialogTitle>
						<DialogContent>
							<Stack
								spacing={3}
								sx={{ mt: 1 }}
							>
								<TextField
									label="Name"
									value={endpointFormData.name}
									onChange={(e) => setEndpointFormData((prev) => ({ ...prev, name: e.target.value }))}
									fullWidth
									required
								/>
								<TextField
									label="Description"
									value={endpointFormData.description}
									onChange={(e) => setEndpointFormData((prev) => ({ ...prev, description: e.target.value }))}
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
														checked={endpointFormData.enable_api_key_auth}
														onChange={(e) =>
															setEndpointFormData((prev) => ({
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
														checked={endpointFormData.enable_oauth}
														onChange={(e) =>
															setEndpointFormData((prev) => ({
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
														checked={endpointFormData.enable_public_access}
														onChange={(e) =>
															setEndpointFormData((prev) => ({
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
												value={endpointFormData.rate_limit_requests}
												onChange={(e) =>
													setEndpointFormData((prev) => ({
														...prev,
														rate_limit_requests: parseInt(e.target.value) || 0
													}))
												}
												fullWidth
											/>
											<TextField
												label="Window (seconds)"
												type="number"
												value={endpointFormData.rate_limit_window}
												onChange={(e) =>
													setEndpointFormData((prev) => ({
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
											checked={endpointFormData.is_active}
											onChange={(e) =>
												setEndpointFormData((prev) => ({ ...prev, is_active: e.target.checked }))
											}
										/>
									}
									label="Active"
								/>
							</Stack>
						</DialogContent>
						<DialogActions>
							<Button onClick={() => setEditEndpointModalOpen(false)}>Cancel</Button>
							<Button
								variant="contained"
								onClick={handleSaveEndpoint}
								disabled={!endpointFormData.name.trim()}
							>
								Save Changes
							</Button>
						</DialogActions>
					</Dialog>
				</div>
			}
		/>
	);
}

export default NamespaceDetailView;
