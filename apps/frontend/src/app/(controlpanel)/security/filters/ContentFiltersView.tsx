'use client';

import { useState, useMemo, useCallback, useEffect } from 'react';
import { MRT_ColumnDef, MRT_Row } from 'material-react-table';
import PageSimple from '@fuse/core/PageSimple';
import { styled } from '@mui/material/styles';
import {
	Typography,
	Button,
	Chip,
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
import DataTable from '@/components/data-table/DataTable';
import SvgIcon from '@fuse/core/SvgIcon';
import { useSnackbar } from 'notistack';
import { contentFilterApi } from '@/lib/client-api';
import type { ContentFilter, CreateContentFilterRequest, UpdateContentFilterRequest } from '@/lib/types';

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

// UI-specific mapping to match the backend types
interface UIContentFilter extends Omit<ContentFilter, 'enabled'> {
	is_active: boolean;
}

const getFilterTypeColor = (type: string): 'primary' | 'warning' | 'error' | 'success' => {
	switch (type) {
		case 'pii':
			return 'error';
		case 'resource':
			return 'warning';
		case 'deny':
			return 'primary';
		case 'regex':
			return 'success';
		default:
			return 'primary';
	}
};

function ContentFiltersView() {
	const [createModalOpen, setCreateModalOpen] = useState(false);
	const [editingFilter, setEditingFilter] = useState<ContentFilter | null>(null);
	const [filters, setFilters] = useState<ContentFilter[]>([]);
	const [isLoading, setIsLoading] = useState(false);
	const [formData, setFormData] = useState({
		name: '',
		description: '',
		type: 'pii' as 'pii' | 'resource' | 'deny' | 'regex',
		enabled: true,
		priority: 100,
		config: {} as Record<string, any>
	});
	const { enqueueSnackbar } = useSnackbar();

	// Fetch filters on mount
	useEffect(() => {
		fetchFilters();
	}, []);

	const fetchFilters = async () => {
		setIsLoading(true);
		try {
			const data = await contentFilterApi.listFilters();
			setFilters(data);
		} catch (error) {
			enqueueSnackbar('Failed to fetch content filters', { variant: 'error' });
			console.error('Error fetching filters:', error);
			setFilters([]);
		} finally {
			setIsLoading(false);
		}
	};

	// Transform filters for UI display (map 'enabled' to 'is_active')
	const uiFilters = useMemo<UIContentFilter[]>(
		() => filters.map(filter => ({
			...filter,
			is_active: filter.enabled
		})),
		[filters]
	);

	const columns = useMemo<MRT_ColumnDef<UIContentFilter>[]>(
		() => [
			{
				accessorKey: 'name',
				header: 'Name',
				size: 200
			},
			{
				accessorKey: 'description',
				header: 'Description',
				size: 300
			},
			{
				accessorKey: 'type',
				header: 'Type',
				size: 150,
				Cell: ({ cell }) => (
					<Chip
						label={cell.getValue<string>()}
						color={getFilterTypeColor(cell.getValue<string>())}
						size="small"
						variant="filled"
					/>
				)
			},
			{
				accessorKey: 'priority',
				header: 'Priority',
				size: 100
			},
			{
				accessorKey: 'is_active',
				header: 'Status',
				size: 100,
				Cell: ({ cell }) => (
					<Chip
						label={cell.getValue<boolean>() ? 'Active' : 'Inactive'}
						color={cell.getValue<boolean>() ? 'success' : 'default'}
						size="small"
					/>
				)
			}
		],
		[]
	);

	const handleCreateFilter = useCallback(() => {
		setFormData({
			name: '',
			description: '',
			type: 'pii',
			enabled: true,
			priority: 100,
			config: {}
		});
		setEditingFilter(null);
		setCreateModalOpen(true);
	}, []);

	const handleEditFilter = useCallback((filter: ContentFilter) => {
		setFormData({
			name: filter.name,
			description: filter.description || '',
			type: filter.type,
			enabled: filter.enabled,
			priority: filter.priority,
			config: filter.config
		});
		setEditingFilter(filter);
		setCreateModalOpen(true);
	}, []);

	const handleDeleteFilter = useCallback(
		async (filter: ContentFilter) => {
			try {
				await contentFilterApi.deleteFilter(filter.id);
				enqueueSnackbar(`Filter "${filter.name}" deleted successfully`, { variant: 'success' });
				fetchFilters(); // Refresh the list
			} catch (error) {
				enqueueSnackbar(`Failed to delete filter "${filter.name}"`, { variant: 'error' });
				console.error('Error deleting filter:', error);
			}
		},
		[enqueueSnackbar]
	);

	const handleSaveFilter = useCallback(async () => {
		try {
			if (editingFilter) {
				const updateData: UpdateContentFilterRequest = {
					name: formData.name,
					description: formData.description,
					priority: formData.priority,
					enabled: formData.enabled,
					config: formData.config
				};
				await contentFilterApi.updateFilter(editingFilter.id, updateData);
				enqueueSnackbar('Filter updated successfully', { variant: 'success' });
			} else {
				const createData: CreateContentFilterRequest = {
					name: formData.name,
					description: formData.description,
					type: formData.type,
					enabled: formData.enabled,
					priority: formData.priority,
					config: formData.config
				};
				await contentFilterApi.createFilter(createData);
				enqueueSnackbar('Filter created successfully', { variant: 'success' });
			}

			setCreateModalOpen(false);
			setEditingFilter(null);
			fetchFilters(); // Refresh the list
		} catch (error) {
			const action = editingFilter ? 'update' : 'create';
			enqueueSnackbar(`Failed to ${action} filter`, { variant: 'error' });
			console.error(`Error ${action}ing filter:`, error);
		}
	}, [editingFilter, formData, enqueueSnackbar]);

	return (
		<Root
			header={
				<div className="p-6">
					<div className="flex items-center justify-between">
						<div>
							<Typography variant="h4">Content Filters</Typography>
							<Typography
								variant="body1"
								color="textSecondary"
								className="mt-1"
							>
								Manage content filtering rules for input and output
							</Typography>
						</div>
						<Button
							variant="contained"
							color="primary"
							startIcon={<SvgIcon>lucide:plus</SvgIcon>}
							onClick={handleCreateFilter}
						>
							Create Filter
						</Button>
					</div>
				</div>
			}
			content={
				<div className="p-6">
					<DataTable
						columns={columns}
						data={uiFilters}
						state={{ isLoading }}
						enableRowActions
						renderRowActionMenuItems={({ row, closeMenu }) => [
							<MenuItem
								key="edit"
								onClick={() => {
									// Convert UIContentFilter back to ContentFilter for editing
									const filter: ContentFilter = {
										...row.original,
										enabled: row.original.is_active
									};
									handleEditFilter(filter);
									closeMenu();
								}}
							>
								<SvgIcon size={16}>lucide:edit</SvgIcon>
								<span className="ml-2">Edit</span>
							</MenuItem>,
							<MenuItem
								key="delete"
								onClick={() => {
									// Convert UIContentFilter back to ContentFilter for deletion
									const filter: ContentFilter = {
										...row.original,
										enabled: row.original.is_active
									};
									handleDeleteFilter(filter);
									closeMenu();
								}}
							>
								<SvgIcon size={16}>lucide:trash</SvgIcon>
								<span className="ml-2">Delete</span>
							</MenuItem>
						]}
					/>

					<Dialog
						open={createModalOpen}
						onClose={() => setCreateModalOpen(false)}
						maxWidth="sm"
						fullWidth
					>
						<DialogTitle>{editingFilter ? 'Edit Filter' : 'Create New Filter'}</DialogTitle>
						<DialogContent>
							<Stack
								spacing={3}
								className="mt-2"
							>
								<TextField
									label="Name"
									value={formData.name}
									onChange={(e) => setFormData({ ...formData, name: e.target.value })}
									fullWidth
									required
								/>
								<TextField
									label="Description"
									value={formData.description}
									onChange={(e) => setFormData({ ...formData, description: e.target.value })}
									fullWidth
									multiline
									rows={2}
								/>
								<TextField
									select
									label="Type"
									value={formData.type}
									onChange={(e) =>
										setFormData({
											...formData,
											type: e.target.value as 'pii' | 'resource' | 'deny' | 'regex'
										})
									}
									fullWidth
									disabled={editingFilter !== null} // Don't allow changing type when editing
								>
									<MenuItem value="pii">PII Detection</MenuItem>
									<MenuItem value="resource">Resource Filter</MenuItem>
									<MenuItem value="deny">Content Blocking</MenuItem>
									<MenuItem value="regex">Regex Pattern</MenuItem>
								</TextField>
								<TextField
									label="Priority"
									type="number"
									value={formData.priority}
									onChange={(e) => setFormData({ ...formData, priority: parseInt(e.target.value) })}
									fullWidth
									helperText="Higher priority filters are applied first"
								/>
								<FormControlLabel
									control={
										<Switch
											checked={formData.enabled}
											onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })}
										/>
									}
									label="Enabled"
								/>
							</Stack>
						</DialogContent>
						<DialogActions>
							<Button onClick={() => setCreateModalOpen(false)}>Cancel</Button>
							<Button
								variant="contained"
								onClick={handleSaveFilter}
							>
								{editingFilter ? 'Update' : 'Create'}
							</Button>
						</DialogActions>
					</Dialog>
				</div>
			}
		/>
	);
}

export default ContentFiltersView;
