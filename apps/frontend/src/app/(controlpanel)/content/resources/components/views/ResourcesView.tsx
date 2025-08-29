'use client';

import { useState, useMemo, useCallback } from 'react';
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
	MenuItem,
	FormControlLabel,
	Switch,
	Divider,
	Card,
	CardContent
} from '@mui/material';
import LazyDataTable from '@/components/data-table/LazyDataTable';
import SvgIcon from '@fuse/core/SvgIcon';
import { useSnackbar } from 'notistack';

// Mock types
interface Resource {
	id: string;
	name: string;
	type: string;
	content_type: string;
	description: string;
	uri: string;
	is_active: boolean;
	created_at: string;
	updated_at: string;
}

interface CreateResourceRequest {
	name: string;
	type: string;
	content_type: string;
	description: string;
	uri: string;
}

interface UpdateResourceRequest {
	name?: string;
	type?: string;
	content_type?: string;
	description?: string;
	uri?: string;
	is_active?: boolean;
}

interface ResourceFormData {
	name: string;
	type: string;
	content_type: string;
	description: string;
	uri: string;
	is_active: boolean;
}

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


const RESOURCE_TYPE_ICONS = {
	file: 'lucide:file-text',
	url: 'lucide:link',
	database: 'lucide:database',
	api: 'lucide:globe',
	memory: 'lucide:hard-drive',
	custom: 'lucide:code'
};

const RESOURCE_TYPE_COLORS = {
	file: 'primary',
	url: 'success',
	database: 'secondary',
	api: 'warning',
	memory: 'default',
	custom: 'info'
} as const;

function ResourcesView() {
	const [createModalOpen, setCreateModalOpen] = useState(false);
	const [viewModalOpen, setViewModalOpen] = useState(false);
	const [editingResource, setEditingResource] = useState<Resource | null>(null);
	const [viewingResource, setViewingResource] = useState<Resource | null>(null);
	const [formData, setFormData] = useState<ResourceFormData>({
		name: '',
		type: 'text',
		content_type: 'text/plain',
		uri: '',
		description: '',
		is_active: true
	});

	// API hooks
	const { enqueueSnackbar } = useSnackbar();

	// Mock data
	const [resources, setResources] = useState<Resource[]>([
		{
			id: '1',
			name: 'API Documentation',
			type: 'text',
			content_type: 'text/markdown',
			description: 'API reference documentation',
			uri: 'https://api.example.com/docs',
			is_active: true,
			created_at: '2024-01-01T00:00:00Z',
			updated_at: '2024-01-01T00:00:00Z'
		},
		{
			id: '2',
			name: 'User Guide PDF',
			type: 'text',
			content_type: 'application/pdf',
			description: 'Comprehensive user guide',
			uri: 'https://docs.example.com/user-guide.pdf',
			is_active: true,
			created_at: '2024-01-01T00:00:00Z',
			updated_at: '2024-01-01T00:00:00Z'
		}
	]);
	const isLoading = false;

	// Mock functions
	const createResource = {
		mutate: (data: CreateResourceRequest) => {
			const newResource: Resource = {
				id: Date.now().toString(),
				...data,
				is_active: true,
				created_at: new Date().toISOString(),
				updated_at: new Date().toISOString()
			};
			setResources(prev => [...prev, newResource]);
			enqueueSnackbar('Resource created successfully', { variant: 'success' });
		},
		isPending: false
	};

	const updateResource = {
		mutate: (data: { id: string; data: Partial<UpdateResourceRequest> }) => {
			setResources(prev => prev.map(r => r.id === data.id ? { ...r, ...data.data, updated_at: new Date().toISOString() } : r));
			enqueueSnackbar('Resource updated successfully', { variant: 'success' });
		},
		isPending: false
	};

	const deleteResource = {
		mutate: (id: string) => {
			setResources(prev => prev.filter(r => r.id !== id));
			enqueueSnackbar('Resource deleted successfully', { variant: 'success' });
		},
		isPending: false
	};
	const [togglingResourceId, setTogglingResourceId] = useState<string | null>(null);


	const handleCreateResource = () => {
		setFormData({
			name: '',
			type: 'text',
			content_type: 'text/plain',
			uri: '',
			description: '',
			is_active: true
		});
		setEditingResource(null);
		setCreateModalOpen(true);
	};

	const handleViewResource = (resource: Resource) => {
		setViewingResource(resource);
		setViewModalOpen(true);
	};

	const handleEditResource = (resource: Resource) => {
		setFormData({
			name: resource.name,
			type: resource.type,
			content_type: resource.content_type,
			uri: resource.uri,
			description: resource.description || '',
			is_active: resource.is_active
		});
		setEditingResource(resource);
		setCreateModalOpen(true);
	};

	const handleSaveResource = () => {
		if (editingResource) {
			const updateData: UpdateResourceRequest = {
				name: formData.name,
				type: formData.type,
				content_type: formData.content_type,
				uri: formData.uri,
				description: formData.description
			};
			updateResource.mutate({
				id: editingResource.id,
				data: updateData
			});
		} else {
			const createData: CreateResourceRequest = {
				name: formData.name,
				type: formData.type,
				content_type: formData.content_type,
				uri: formData.uri,
				description: formData.description
			};
			createResource.mutate(createData);
		}
		setCreateModalOpen(false);
		setEditingResource(null);
	};

	const handleDeleteResource = (resource: Resource) => {
		if (confirm(`Are you sure you want to delete "${resource.name}"? This action cannot be undone.`)) {
			deleteResource.mutate(resource.id);
		}
	};

	const handleToggleResourceStatus = useCallback((resource: Resource) => {
		setTogglingResourceId(resource.id);
		const updateData: UpdateResourceRequest = {
			is_active: !resource.is_active
		};
		updateResource.mutate({
			id: resource.id,
			data: updateData
		});
		setTogglingResourceId(null);
	}, [updateResource]);

	const formatFileSize = (bytes?: number) => {
		if (!bytes) return '-';

		const sizes = ['B', 'KB', 'MB', 'GB'];
		const i = Math.floor(Math.log(bytes) / Math.log(1024));
		return `${(bytes / Math.pow(1024, i)).toFixed(2)} ${sizes[i]}`;
	};

	const columns = useMemo<MRT_ColumnDef<Resource>[]>(
		() => [
			{
				accessorKey: 'is_active',
				header: 'Active',
				size: 80,
				Cell: ({ row }) => (
					<Tooltip title={row.original.is_active ? 'Deactivate resource' : 'Activate resource'}>
						<Switch
							checked={!!row.original.is_active}
							onChange={() => handleToggleResourceStatus(row.original)}
							disabled={togglingResourceId === row.original.id}
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
				Cell: ({ row }) => {
					const iconName =
						RESOURCE_TYPE_ICONS[row.original.type as keyof typeof RESOURCE_TYPE_ICONS] ||
						'lucide:code';
					return (
						<Box className="flex items-center space-x-2">
							<SvgIcon size={20}>{iconName}</SvgIcon>
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
					);
				}
			},
			{
				accessorKey: 'type',
				header: 'Type',
				size: 120,
				Cell: ({ cell }) => {
					const color =
						RESOURCE_TYPE_COLORS[cell.getValue<string>() as keyof typeof RESOURCE_TYPE_COLORS] || 'default';
					return (
						<Chip
							size="small"
							label={cell.getValue<string>()}
							color={color}
						/>
					);
				}
			},
			{
				accessorKey: 'uri',
				header: 'URI',
				size: 350,
				Cell: ({ cell }) => (
					<Typography
						variant="body2"
						color="textSecondary"
						className="truncate font-mono text-xs"
						title={cell.getValue<string>()}
					>
						{cell.getValue<string>()}
					</Typography>
				)
			},
			{
				accessorKey: 'size_bytes',
				header: 'Size',
				size: 100,
				Cell: ({ cell }) => <Typography variant="body2">{formatFileSize(cell.getValue<number>())}</Typography>
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
		[handleToggleResourceStatus, togglingResourceId]
	);

	return (
		<Root
			header={
				<div className="p-6">
					<div className="flex items-center justify-between">
						<div>
							<Typography variant="h4">Resources</Typography>
							<Typography
								variant="body1"
								color="textSecondary"
								className="mt-1"
							>
								Manage your MCP gateway resources
							</Typography>
						</div>
						<Button
							variant="contained"
							color="primary"
							startIcon={<SvgIcon>lucide:plus</SvgIcon>}
							onClick={handleCreateResource}
						>
							Create Resource
						</Button>
					</div>
				</div>
			}
			content={
				<div className="p-6">
					<LazyDataTable
						columns={columns}
						data={resources}
						state={{ isLoading: isLoading }}
						enableRowActions
						renderRowActions={({ row }) => (
							<Box className="flex items-center space-x-1">
								<Tooltip title="View Details">
									<IconButton
										size="small"
										onClick={() => handleViewResource(row.original)}
									>
										<SvgIcon size={18}>lucide:eye</SvgIcon>
									</IconButton>
								</Tooltip>
								<Tooltip title="Edit Resource">
									<IconButton
										size="small"
										onClick={() => handleEditResource(row.original)}
									>
										<SvgIcon size={18}>lucide:edit</SvgIcon>
									</IconButton>
								</Tooltip>
								<Tooltip title="Delete Resource">
									<IconButton
										size="small"
										color="error"
										onClick={() => handleDeleteResource(row.original)}
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

					{/* Create/Edit Resource Dialog */}
					<Dialog
						open={createModalOpen}
						onClose={() => setCreateModalOpen(false)}
						maxWidth="md"
						fullWidth
					>
						<DialogTitle>{editingResource ? 'Edit Resource' : 'Create Resource'}</DialogTitle>
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
									label="Type"
									value={formData.type}
									onChange={(e) =>
										setFormData((prev) => ({ ...prev, type: e.target.value }))
									}
									select
									fullWidth
									required
								>
									<MenuItem value="text">Text</MenuItem>
									<MenuItem value="image">Image</MenuItem>
									<MenuItem value="file">File</MenuItem>
									<MenuItem value="url">URL</MenuItem>
									<MenuItem value="database">Database</MenuItem>
									<MenuItem value="api">API</MenuItem>
									<MenuItem value="memory">Memory</MenuItem>
									<MenuItem value="custom">Custom</MenuItem>
								</TextField>
								<TextField
									label="Content Type"
									value={formData.content_type}
									onChange={(e) =>
										setFormData((prev) => ({ ...prev, content_type: e.target.value }))
									}
									fullWidth
									required
									helperText="MIME type (e.g., text/plain, application/json)"
								/>
								<TextField
									label="URI"
									value={formData.uri}
									onChange={(e) => setFormData((prev) => ({ ...prev, uri: e.target.value }))}
									fullWidth
									required
									helperText="Resource location (path, URL, or connection string)"
								/>
								<TextField
									label="Description"
									value={formData.description}
									onChange={(e) => setFormData((prev) => ({ ...prev, description: e.target.value }))}
									fullWidth
									multiline
									rows={3}
								/>
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
								onClick={handleSaveResource}
								disabled={!formData.name.trim() || !formData.uri.trim()}
							>
								{editingResource ? 'Update' : 'Create'}
							</Button>
						</DialogActions>
					</Dialog>

					{/* View Resource Dialog */}
					<Dialog
						open={viewModalOpen}
						onClose={() => setViewModalOpen(false)}
						maxWidth="md"
						fullWidth
					>
						<DialogTitle>Resource Details</DialogTitle>
						<DialogContent>
							{viewingResource && (
								<Stack spacing={3} sx={{ mt: 1 }}>
									<Card variant="outlined">
										<CardContent>
											<Stack spacing={2}>
												<Box>
													<Typography variant="subtitle2" color="textSecondary">
														Name
													</Typography>
													<Typography variant="body1">{viewingResource.name}</Typography>
												</Box>

												<Box>
													<Typography variant="subtitle2" color="textSecondary">
														Type
													</Typography>
													<Chip
														size="small"
														label={viewingResource.type}
														color={RESOURCE_TYPE_COLORS[viewingResource.type as keyof typeof RESOURCE_TYPE_COLORS] || 'default'}
													/>
												</Box>

												<Box>
													<Typography variant="subtitle2" color="textSecondary">
														Content Type
													</Typography>
													<Typography variant="body1">{viewingResource.content_type}</Typography>
												</Box>

												<Box>
													<Typography variant="subtitle2" color="textSecondary">
														URI
													</Typography>
													<Typography variant="body2" className="font-mono break-all">
														{viewingResource.uri}
													</Typography>
												</Box>

												{viewingResource.description && (
													<Box>
														<Typography variant="subtitle2" color="textSecondary">
															Description
														</Typography>
														<Typography variant="body1">
															{viewingResource.description}
														</Typography>
													</Box>
												)}

												{viewingResource.size_bytes && (
													<Box>
														<Typography variant="subtitle2" color="textSecondary">
															Size
														</Typography>
														<Typography variant="body1">
															{formatFileSize(viewingResource.size_bytes)}
														</Typography>
													</Box>
												)}

												<Box>
													<Typography variant="subtitle2" color="textSecondary">
														Status
													</Typography>
													<Chip
														size="small"
														label={viewingResource.is_active ? 'Active' : 'Inactive'}
														color={viewingResource.is_active ? 'success' : 'default'}
													/>
												</Box>

												<Divider />

												<Box>
													<Typography variant="subtitle2" color="textSecondary">
														Created
													</Typography>
													<Typography variant="body2">
														{new Date(viewingResource.created_at).toLocaleString()}
													</Typography>
												</Box>

												<Box>
													<Typography variant="subtitle2" color="textSecondary">
														Last Updated
													</Typography>
													<Typography variant="body2">
														{new Date(viewingResource.updated_at).toLocaleString()}
													</Typography>
												</Box>
											</Stack>
										</CardContent>
									</Card>
								</Stack>
							)}
						</DialogContent>
						<DialogActions>
							<Button onClick={() => setViewModalOpen(false)}>Close</Button>
							{viewingResource && (
								<Button
									variant="contained"
									color="primary"
									onClick={() => {
										setViewModalOpen(false);
										handleEditResource(viewingResource);
									}}
								>
									Edit Resource
								</Button>
							)}
						</DialogActions>
					</Dialog>
				</div>
			}
		/>
	);
}

export default ResourcesView;
