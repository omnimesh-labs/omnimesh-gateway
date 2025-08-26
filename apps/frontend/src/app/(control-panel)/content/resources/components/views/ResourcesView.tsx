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
	MenuItem,
	FormControlLabel,
	Switch
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

interface Resource {
	id: string;
	name: string;
	resource_type: string;
	uri: string;
	description?: string;
	size_bytes?: number;
	is_active: boolean;
	created_at: string;
	updated_at: string;
}

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
	const [editingResource, setEditingResource] = useState<Resource | null>(null);
	const [formData, setFormData] = useState({
		name: '',
		resource_type: 'file',
		uri: '',
		description: '',
		is_active: true
	});
	const { enqueueSnackbar } = useSnackbar();

	// Mock data
	const mockResources: Resource[] = [
		{
			id: '1',
			name: 'User Database',
			resource_type: 'database',
			uri: 'postgres://db.example.com:5432/users',
			description: 'Main user database',
			size_bytes: 5242880000,
			is_active: true,
			created_at: '2024-01-15T10:00:00Z',
			updated_at: '2024-01-20T15:30:00Z'
		},
		{
			id: '2',
			name: 'API Documentation',
			resource_type: 'url',
			uri: 'https://docs.example.com/api',
			description: 'API documentation site',
			size_bytes: undefined,
			is_active: true,
			created_at: '2024-01-10T09:00:00Z',
			updated_at: '2024-01-18T14:00:00Z'
		},
		{
			id: '3',
			name: 'Config File',
			resource_type: 'file',
			uri: '/configs/app.yaml',
			description: 'Application configuration',
			size_bytes: 2048,
			is_active: false,
			created_at: '2024-01-12T11:00:00Z',
			updated_at: '2024-01-22T16:00:00Z'
		},
		{
			id: '4',
			name: 'Weather API',
			resource_type: 'api',
			uri: 'https://api.weather.com/v1',
			description: 'External weather service',
			size_bytes: undefined,
			is_active: true,
			created_at: '2024-01-14T08:00:00Z',
			updated_at: '2024-01-21T10:00:00Z'
		}
	];

	const handleCreateResource = () => {
		setFormData({
			name: '',
			resource_type: 'file',
			uri: '',
			description: '',
			is_active: true
		});
		setEditingResource(null);
		setCreateModalOpen(true);
	};

	const handleEditResource = (resource: Resource) => {
		setFormData({
			name: resource.name,
			resource_type: resource.resource_type,
			uri: resource.uri,
			description: resource.description || '',
			is_active: resource.is_active
		});
		setEditingResource(resource);
		setCreateModalOpen(true);
	};

	const handleSaveResource = () => {
		const action = editingResource ? 'updated' : 'created';
		enqueueSnackbar(`Resource ${action} successfully (demo)`, { variant: 'success' });
		setCreateModalOpen(false);
		setEditingResource(null);
	};

	const handleDeleteResource = (resource: Resource) => {
		enqueueSnackbar(`Delete functionality coming soon for ${resource.name}`, { variant: 'info' });
	};

	const formatFileSize = (bytes?: number) => {
		if (!bytes) return '-';

		const sizes = ['B', 'KB', 'MB', 'GB'];
		const i = Math.floor(Math.log(bytes) / Math.log(1024));
		return `${(bytes / Math.pow(1024, i)).toFixed(2)} ${sizes[i]}`;
	};

	const columns = useMemo<MRT_ColumnDef<Resource>[]>(
		() => [
			{
				accessorKey: 'name',
				header: 'Name',
				size: 200,
				Cell: ({ row }) => {
					const iconName =
						RESOURCE_TYPE_ICONS[row.original.resource_type as keyof typeof RESOURCE_TYPE_ICONS] ||
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
				accessorKey: 'resource_type',
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
		[]
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
						data={mockResources}
						enableRowActions
						renderRowActions={({ row }) => (
							<Box className="flex items-center space-x-1">
								<Tooltip title="View Details">
									<IconButton
										size="small"
										onClick={() =>
											enqueueSnackbar(`View details for ${row.original.name}`, {
												variant: 'info'
											})
										}
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
									label="Resource Type"
									value={formData.resource_type}
									onChange={(e) =>
										setFormData((prev) => ({ ...prev, resource_type: e.target.value }))
									}
									select
									fullWidth
									required
								>
									<MenuItem value="file">File</MenuItem>
									<MenuItem value="url">URL</MenuItem>
									<MenuItem value="database">Database</MenuItem>
									<MenuItem value="api">API</MenuItem>
									<MenuItem value="memory">Memory</MenuItem>
									<MenuItem value="custom">Custom</MenuItem>
								</TextField>
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
				</div>
			}
		/>
	);
}

export default ResourcesView;
