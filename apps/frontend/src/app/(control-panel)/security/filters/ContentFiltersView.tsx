'use client';

import { useState, useMemo, useCallback } from 'react';
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

interface ContentFilter {
	id: string;
	name: string;
	description?: string;
	type: 'input' | 'output' | 'bidirectional';
	scope: 'global' | 'namespace' | 'server';
	is_active: boolean;
	priority: number;
	created_at: string;
	updated_at: string;
	rules: {
		patterns?: string[];
		max_length?: number;
		allowed_types?: string[];
		blocked_keywords?: string[];
	};
}

const getFilterTypeColor = (type: string): 'primary' | 'warning' | 'error' => {
	switch (type) {
		case 'input':
			return 'primary';
		case 'output':
			return 'warning';
		case 'bidirectional':
			return 'error';
		default:
			return 'primary';
	}
};

function ContentFiltersView() {
	const [createModalOpen, setCreateModalOpen] = useState(false);
	const [editingFilter, setEditingFilter] = useState<ContentFilter | null>(null);
	const [formData, setFormData] = useState({
		name: '',
		description: '',
		type: 'input' as 'input' | 'output' | 'bidirectional',
		scope: 'namespace' as 'global' | 'namespace' | 'server',
		is_active: true,
		priority: 100
	});
	const { enqueueSnackbar } = useSnackbar();

	// Mock data - memoized to prevent recreation on each render
	const mockFilters = useMemo<ContentFilter[]>(
		() => [
			{
				id: '1',
				name: 'PII Data Filter',
				description: 'Filters out personally identifiable information',
				type: 'bidirectional',
				scope: 'global',
				is_active: true,
				priority: 100,
				created_at: '2024-01-15T10:00:00Z',
				updated_at: '2024-01-20T15:30:00Z',
				rules: {
					patterns: ['\\b\\d{3}-\\d{2}-\\d{4}\\b', '\\b[A-Z]{2}\\d{6}\\b'],
					blocked_keywords: ['ssn', 'social security', 'credit card']
				}
			},
			{
				id: '2',
				name: 'Profanity Filter',
				description: 'Blocks inappropriate language',
				type: 'output',
				scope: 'namespace',
				is_active: true,
				priority: 90,
				created_at: '2024-01-16T11:00:00Z',
				updated_at: '2024-01-18T14:20:00Z',
				rules: {
					blocked_keywords: ['explicit', 'content', 'list']
				}
			},
			{
				id: '3',
				name: 'Input Size Limiter',
				description: 'Limits the size of incoming requests',
				type: 'input',
				scope: 'server',
				is_active: false,
				priority: 80,
				created_at: '2024-01-17T09:30:00Z',
				updated_at: '2024-01-17T09:30:00Z',
				rules: {
					max_length: 10000,
					allowed_types: ['text/plain', 'application/json']
				}
			},
			{
				id: '4',
				name: 'SQL Injection Prevention',
				description: 'Prevents SQL injection attempts',
				type: 'input',
				scope: 'global',
				is_active: true,
				priority: 95,
				created_at: '2024-01-18T13:00:00Z',
				updated_at: '2024-01-19T10:15:00Z',
				rules: {
					patterns: [
						'(\\bSELECT\\b.*\\bFROM\\b|\\bINSERT\\b.*\\bINTO\\b|\\bUPDATE\\b.*\\bSET\\b|\\bDELETE\\b.*\\bFROM\\b)',
						'(--|#|/\\*|\\*/)',
						'(\'|"|\\\\x00|\\\\n|\\\\r|\\\\x1a)'
					]
				}
			}
		],
		[]
	);

	const columns = useMemo<MRT_ColumnDef<ContentFilter>[]>(
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
				accessorKey: 'scope',
				header: 'Scope',
				size: 120,
				Cell: ({ cell }) => (
					<Chip
						label={cell.getValue<string>()}
						size="small"
						variant="outlined"
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

	const handleCreateFilter = () => {
		setEditingFilter(null);
		setFormData({
			name: '',
			description: '',
			type: 'input',
			scope: 'namespace',
			is_active: true,
			priority: 100
		});
		setCreateModalOpen(true);
	};

	const handleEditFilter = (row: MRT_Row<ContentFilter>) => {
		const filter = row.original;
		setEditingFilter(filter);
		setFormData({
			name: filter.name,
			description: filter.description || '',
			type: filter.type,
			scope: filter.scope,
			is_active: filter.is_active,
			priority: filter.priority
		});
		setCreateModalOpen(true);
	};

	const handleDeleteFilter = useCallback(
		(row: MRT_Row<ContentFilter>) => {
			enqueueSnackbar(`Filter "${row.original.name}" deleted`, { variant: 'success' });
		},
		[enqueueSnackbar]
	);

	const handleSaveFilter = () => {
		if (editingFilter) {
			enqueueSnackbar(`Filter "${formData.name}" updated`, { variant: 'success' });
		} else {
			enqueueSnackbar(`Filter "${formData.name}" created`, { variant: 'success' });
		}

		setCreateModalOpen(false);
	};

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
						data={mockFilters}
						enableRowActions
						renderRowActionMenuItems={({ row, closeMenu }) => [
							<MenuItem
								key="edit"
								onClick={() => {
									handleEditFilter(row);
									closeMenu();
								}}
							>
								<SvgIcon size={16}>lucide:edit</SvgIcon>
								<span className="ml-2">Edit</span>
							</MenuItem>,
							<MenuItem
								key="delete"
								onClick={() => {
									handleDeleteFilter(row);
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
											type: e.target.value as 'input' | 'output' | 'bidirectional'
										})
									}
									fullWidth
								>
									<MenuItem value="input">Input</MenuItem>
									<MenuItem value="output">Output</MenuItem>
									<MenuItem value="bidirectional">Bidirectional</MenuItem>
								</TextField>
								<TextField
									select
									label="Scope"
									value={formData.scope}
									onChange={(e) =>
										setFormData({
											...formData,
											scope: e.target.value as 'global' | 'namespace' | 'server'
										})
									}
									fullWidth
								>
									<MenuItem value="global">Global</MenuItem>
									<MenuItem value="namespace">Namespace</MenuItem>
									<MenuItem value="server">Server</MenuItem>
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
											checked={formData.is_active}
											onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
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
