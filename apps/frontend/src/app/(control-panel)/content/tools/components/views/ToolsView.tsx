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
	MenuItem,
	FormControlLabel,
	Switch
} from '@mui/material';
import LazyDataTable from '@/components/data-table/LazyDataTable';
import SvgIcon from '@fuse/core/SvgIcon';
import { useSnackbar } from 'notistack';
import { toolApi, Tool as ApiTool } from '@/lib/api';

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

// Use the ApiTool type from the API
type Tool = ApiTool;

const CATEGORY_ICONS = {
	general: 'lucide:wrench',
	data: 'lucide:database',
	file: 'lucide:file-text',
	web: 'lucide:globe',
	system: 'lucide:terminal',
	ai: 'lucide:brain',
	dev: 'lucide:code',
	custom: 'lucide:package'
};

const CATEGORY_COLORS = {
	general: 'default',
	data: 'primary',
	file: 'success',
	web: 'secondary',
	system: 'warning',
	ai: 'info',
	dev: 'primary',
	custom: 'default'
} as const;

function ToolsView() {
	const [createModalOpen, setCreateModalOpen] = useState(false);
	const [editingTool, setEditingTool] = useState<Tool | null>(null);
	const [tools, setTools] = useState<Tool[]>([]);
	const [loading, setLoading] = useState(true);
	const [formData, setFormData] = useState({
		name: '',
		function_name: '',
		category: 'general',
		description: '',
		implementation_type: 'internal',
		is_public: false,
		is_active: true
	});
	const { enqueueSnackbar } = useSnackbar();

	// Load tools from API
	const loadTools = useCallback(async () => {
		try {
			setLoading(true);
			const response = await toolApi.listTools();
			setTools(response.data || []);
		} catch (error) {
			enqueueSnackbar('Failed to load tools', { variant: 'error' });
			console.error('Error loading tools:', error);
		} finally {
			setLoading(false);
		}
	}, [enqueueSnackbar]);

	useEffect(() => {
		loadTools();
	}, [loadTools]);

	// Mock data fallback (for development)
	const _mockTools: Tool[] = [
		{
			id: '1',
			organization_id: 'org-1',
			name: 'Database Query',
			function_name: 'db_query',
			category: 'data' as const,
			description: 'Execute database queries',
			implementation_type: 'internal' as const,
			timeout_seconds: 30,
			max_retries: 3,
			usage_count: 156,
			is_public: false,
			is_active: true,
			created_at: '2024-01-15T10:00:00Z',
			updated_at: '2024-01-20T15:30:00Z'
		},
		{
			id: '2',
			organization_id: 'org-1',
			name: 'File Upload',
			function_name: 'file_upload',
			category: 'file' as const,
			description: 'Handle file uploads',
			implementation_type: 'external' as const,
			timeout_seconds: 60,
			max_retries: 2,
			usage_count: 89,
			is_public: true,
			is_active: true,
			created_at: '2024-01-10T09:00:00Z',
			updated_at: '2024-01-18T14:00:00Z'
		},
		{
			id: '3',
			organization_id: 'org-1',
			name: 'Web Scraper',
			function_name: 'web_scrape',
			category: 'web' as const,
			description: 'Scrape web content',
			implementation_type: 'webhook' as const,
			timeout_seconds: 45,
			max_retries: 1,
			usage_count: 42,
			is_public: true,
			is_active: false,
			created_at: '2024-01-12T11:00:00Z',
			updated_at: '2024-01-22T16:00:00Z'
		}
	];

	const handleCreateTool = () => {
		setFormData({
			name: '',
			function_name: '',
			category: 'general',
			description: '',
			implementation_type: 'internal',
			is_public: false,
			is_active: true
		});
		setEditingTool(null);
		setCreateModalOpen(true);
	};

	const handleEditTool = (tool: Tool) => {
		setFormData({
			name: tool.name,
			function_name: tool.function_name,
			category: tool.category,
			description: tool.description || '',
			implementation_type: tool.implementation_type,
			is_public: tool.is_public,
			is_active: tool.is_active
		});
		setEditingTool(tool);
		setCreateModalOpen(true);
	};

	const handleSaveTool = () => {
		const action = editingTool ? 'updated' : 'created';
		enqueueSnackbar(`Tool ${action} successfully (demo)`, { variant: 'success' });
		setCreateModalOpen(false);
		setEditingTool(null);
	};

	const handleDeleteTool = (tool: Tool) => {
		enqueueSnackbar(`Delete functionality coming soon for ${tool.name}`, { variant: 'info' });
	};

	const handleExecuteTool = (tool: Tool) => {
		enqueueSnackbar(`Execute functionality coming soon for ${tool.name}`, { variant: 'info' });
	};

	const handleToggleStatus = useCallback(
		async (tool: Tool) => {
			try {
				await toolApi.updateTool(tool.id, { is_active: !tool.is_active });
				enqueueSnackbar(`Tool ${!tool.is_active ? 'activated' : 'deactivated'} successfully`, {
					variant: 'success'
				});
				// Refresh the tools data
				await loadTools();
			} catch (error) {
				enqueueSnackbar('Failed to update tool status', { variant: 'error' });
				console.error('Error updating tool status:', error);
			}
		},
		[loadTools, enqueueSnackbar]
	);

	const columns = useMemo<MRT_ColumnDef<Tool>[]>(
		() => [
			{
				accessorKey: 'name',
				header: 'Name',
				size: 200,
				Cell: ({ row }) => {
					const iconName =
						CATEGORY_ICONS[row.original.category as keyof typeof CATEGORY_ICONS] || 'lucide:package';
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
								<Typography
									variant="caption"
									color="textSecondary"
									className="font-mono"
								>
									{row.original.function_name}
								</Typography>
							</Box>
						</Box>
					);
				}
			},
			{
				accessorKey: 'category',
				header: 'Category',
				size: 120,
				Cell: ({ cell }) => {
					const color = CATEGORY_COLORS[cell.getValue<string>() as keyof typeof CATEGORY_COLORS] || 'default';
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
				accessorKey: 'implementation_type',
				header: 'Type',
				size: 120,
				Cell: ({ cell }) => (
					<Chip
						size="small"
						label={cell.getValue<string>()}
						variant="outlined"
					/>
				)
			},
			{
				accessorKey: 'usage_count',
				header: 'Usage',
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
				accessorKey: 'is_public',
				header: 'Access',
				size: 100,
				Cell: ({ cell }) => (
					<Box className="flex items-center">
						{cell.getValue<boolean>() ? (
							<SvgIcon
								size={18}
								className="text-green-600"
							>
								lucide:unlock
							</SvgIcon>
						) : (
							<SvgIcon
								size={18}
								className="text-gray-400"
							>
								lucide:lock
							</SvgIcon>
						)}
						<Typography
							variant="caption"
							className="ml-1"
						>
							{cell.getValue<boolean>() ? 'Public' : 'Private'}
						</Typography>
					</Box>
				)
			},
			{
				accessorKey: 'is_active',
				header: 'Status',
				size: 120,
				Cell: ({ cell, row }) => (
					<Chip
						size="small"
						label={cell.getValue<boolean>() ? 'Active' : 'Inactive'}
						color={cell.getValue<boolean>() ? 'success' : 'default'}
						onClick={() => handleToggleStatus(row.original)}
						sx={{ cursor: 'pointer' }}
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
		[handleToggleStatus]
	);

	return (
		<Root
			header={
				<div className="p-6">
					<div className="flex items-center justify-between">
						<div>
							<Typography variant="h4">Tools</Typography>
							<Typography
								variant="body1"
								color="textSecondary"
								className="mt-1"
							>
								Manage your MCP tools and functions
							</Typography>
						</div>
						<Button
							variant="contained"
							color="primary"
							startIcon={<SvgIcon>lucide:plus</SvgIcon>}
							onClick={handleCreateTool}
						>
							Create Tool
						</Button>
					</div>
				</div>
			}
			content={
				<div className="p-6">
					<LazyDataTable
						columns={columns}
						data={loading ? [] : tools}
						enableRowActions
						state={{ isLoading: loading }}
						renderRowActions={({ row }) => (
							<Box className="flex items-center space-x-1">
								<Tooltip title="Execute Tool">
									<IconButton
										size="small"
										onClick={() => handleExecuteTool(row.original)}
									>
										<SvgIcon size={18}>lucide:play</SvgIcon>
									</IconButton>
								</Tooltip>
								<Tooltip title="Edit Tool">
									<IconButton
										size="small"
										onClick={() => handleEditTool(row.original)}
									>
										<SvgIcon size={18}>lucide:edit</SvgIcon>
									</IconButton>
								</Tooltip>
								<Tooltip title="Delete Tool">
									<IconButton
										size="small"
										color="error"
										onClick={() => handleDeleteTool(row.original)}
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

					{/* Create/Edit Tool Dialog */}
					<Dialog
						open={createModalOpen}
						onClose={() => setCreateModalOpen(false)}
						maxWidth="md"
						fullWidth
					>
						<DialogTitle>{editingTool ? 'Edit Tool' : 'Create Tool'}</DialogTitle>
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
									sx={{
										'& .MuiOutlinedInput-root': {
											backgroundColor: (theme) => 
												theme.palette.mode === 'light' ? 'rgba(0, 0, 0, 0.02)' : 'inherit'
										}
									}}
								/>
								<TextField
									label="Function Name"
									value={formData.function_name}
									onChange={(e) =>
										setFormData((prev) => ({ ...prev, function_name: e.target.value }))
									}
									fullWidth
									required
									placeholder="Unique identifier for the function"
									sx={{
										'& .MuiOutlinedInput-root': {
											backgroundColor: (theme) => 
												theme.palette.mode === 'light' ? 'rgba(0, 0, 0, 0.02)' : 'inherit'
										}
									}}
								/>
								<TextField
									label="Category"
									value={formData.category}
									onChange={(e) => setFormData((prev) => ({ ...prev, category: e.target.value }))}
									select
									fullWidth
									required
									sx={{
										'& .MuiOutlinedInput-root': {
											backgroundColor: (theme) => 
												theme.palette.mode === 'light' ? 'rgba(0, 0, 0, 0.02)' : 'inherit'
										}
									}}
								>
									<MenuItem value="general">General</MenuItem>
									<MenuItem value="data">Data</MenuItem>
									<MenuItem value="file">File</MenuItem>
									<MenuItem value="web">Web</MenuItem>
									<MenuItem value="system">System</MenuItem>
									<MenuItem value="ai">AI</MenuItem>
									<MenuItem value="dev">Dev</MenuItem>
									<MenuItem value="custom">Custom</MenuItem>
								</TextField>
								<TextField
									label="Implementation Type"
									value={formData.implementation_type}
									onChange={(e) =>
										setFormData((prev) => ({ ...prev, implementation_type: e.target.value }))
									}
									select
									fullWidth
									required
									sx={{
										'& .MuiOutlinedInput-root': {
											backgroundColor: (theme) => 
												theme.palette.mode === 'light' ? 'rgba(0, 0, 0, 0.02)' : 'inherit'
										}
									}}
								>
									<MenuItem value="internal">Internal</MenuItem>
									<MenuItem value="external">External</MenuItem>
									<MenuItem value="webhook">Webhook</MenuItem>
									<MenuItem value="script">Script</MenuItem>
								</TextField>
								<TextField
									label="Description"
									value={formData.description}
									onChange={(e) => setFormData((prev) => ({ ...prev, description: e.target.value }))}
									fullWidth
									multiline
									rows={3}
									sx={{
										'& .MuiOutlinedInput-root': {
											backgroundColor: (theme) => 
												theme.palette.mode === 'light' ? 'rgba(0, 0, 0, 0.02)' : 'inherit'
										}
									}}
								/>
								<FormControlLabel
									control={
										<Switch
											checked={formData.is_public}
											onChange={(e) =>
												setFormData((prev) => ({ ...prev, is_public: e.target.checked }))
											}
										/>
									}
									label="Public Access"
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
								onClick={handleSaveTool}
								disabled={!formData.name.trim() || !formData.function_name.trim()}
							>
								{editingTool ? 'Update' : 'Create'}
							</Button>
						</DialogActions>
					</Dialog>
				</div>
			}
		/>
	);
}

export default ToolsView;
