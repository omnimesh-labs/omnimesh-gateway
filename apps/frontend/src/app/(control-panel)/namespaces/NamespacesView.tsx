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
	Switch
} from '@mui/material';
import LazyDataTable from '@/components/data-table/LazyDataTable';
import SvgIcon from '@fuse/core/SvgIcon';
import { useSnackbar } from 'notistack';
import { namespaceApi } from '@/lib/api';

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

interface Namespace {
	id: string;
	name: string;
	slug: string;
	description?: string;
	server_count: number;
	is_active: boolean;
	created_at: string;
	updated_at: string;
}

function NamespacesView() {
	const [createModalOpen, setCreateModalOpen] = useState(false);
	const [editingNamespace, setEditingNamespace] = useState<Namespace | null>(null);
	const [formData, setFormData] = useState({
		name: '',
		description: '',
		is_active: true
	});
	const [namespaces, setNamespaces] = useState<Namespace[]>([]);
	const [isLoading, setIsLoading] = useState(false);
	const { enqueueSnackbar } = useSnackbar();

	// Fetch namespaces on mount
	useEffect(() => {
		fetchNamespaces();
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

	const handleCreateNamespace = () => {
		setFormData({ name: '', description: '', is_active: true });
		setEditingNamespace(null);
		setCreateModalOpen(true);
	};

	const handleEditNamespace = (namespace: Namespace) => {
		setFormData({
			name: namespace.name,
			description: namespace.description || '',
			is_active: namespace.is_active
		});
		setEditingNamespace(namespace);
		setCreateModalOpen(true);
	};

	const handleSaveNamespace = async () => {
		try {
			if (editingNamespace) {
				await namespaceApi.updateNamespace(editingNamespace.id, formData);
				enqueueSnackbar('Namespace updated successfully', { variant: 'success' });
			} else {
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

	const columns = useMemo<MRT_ColumnDef<Namespace>[]>(
		() => [
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
								{row.original.slug}
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
								<Tooltip title="View Servers">
									<IconButton size="small">
										<SvgIcon size={18}>lucide:server</SvgIcon>
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
				</div>
			}
		/>
	);
}

export default NamespacesView;
