'use client';

import { useState, useMemo, useCallback } from 'react';
import { enqueueSnackbar } from 'notistack';
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
// Mock types
interface Tool {
	id: string;
	name: string;
	description: string;
	function_name: string;
	parameters: any;
	is_active: boolean;
	created_at: string;
	updated_at: string;
}

interface CreateToolRequest {
	name: string;
	description: string;
	function_name: string;
	parameters: any;
}

interface UpdateToolRequest extends CreateToolRequest {
	id: string;
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

// Use the ApiTool type from the API
type Tool = ApiTool;

// Form data type that includes both create and edit fields
interface ToolFormData {
	name: string;
	function_name: string;
	category: string;
	description: string;
	implementation_type: string;
	schema?: any;  // eslint-disable-line @typescript-eslint/no-explicit-any
	is_public: boolean;
	is_active?: boolean;
}

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
	const [viewModalOpen, setViewModalOpen] = useState(false);
	const [editingTool, setEditingTool] = useState<Tool | null>(null);
	const [viewingTool, setViewingTool] = useState<Tool | null>(null);
	const [formData, setFormData] = useState<ToolFormData>({
		name: '',
		function_name: '',
		category: 'general',
		description: '',
		implementation_type: 'internal',
		schema: {},
		is_public: false,
		is_active: true
	});

	// API hooks
	// Mock data
	const [tools, setTools] = useState<Tool[]>([
		{
			id: '1',
			name: 'Get Weather',
			description: 'Retrieves current weather information for a location',
			function_name: 'get_weather',
			parameters: {
				type: 'object',
				properties: {
					location: { type: 'string', description: 'City name' }
				},
				required: ['location']
			},
			is_active: true,
			created_at: '2024-01-01T00:00:00Z',
			updated_at: '2024-01-01T00:00:00Z'
		},
		{
			id: '2',
			name: 'Send Email',
			description: 'Sends an email to a recipient',
			function_name: 'send_email',
			parameters: {
				type: 'object',
				properties: {
					to: { type: 'string', description: 'Email address' },
					subject: { type: 'string', description: 'Email subject' },
					body: { type: 'string', description: 'Email body' }
				},
				required: ['to', 'subject', 'body']
			},
			is_active: true,
			created_at: '2024-01-01T00:00:00Z',
			updated_at: '2024-01-01T00:00:00Z'
		}
	]);
	const isLoading = false;

	// Mock functions
	const createTool = {
		mutate: (data: CreateToolRequest) => {
			const newTool: Tool = {
				id: Date.now().toString(),
				...data,
				is_active: true,
				created_at: new Date().toISOString(),
				updated_at: new Date().toISOString()
			};
			setTools(prev => [...prev, newTool]);
			enqueueSnackbar('Tool created successfully', { variant: 'success' });
		},
		isPending: false
	};

	const updateTool = {
		mutate: (data: UpdateToolRequest) => {
			setTools(prev => prev.map(t => t.id === data.id ? { ...t, ...data, updated_at: new Date().toISOString() } : t));
			enqueueSnackbar('Tool updated successfully', { variant: 'success' });
		},
		isPending: false
	};

	const deleteTool = {
		mutate: (id: string) => {
			setTools(prev => prev.filter(t => t.id !== id));
			enqueueSnackbar('Tool deleted successfully', { variant: 'success' });
		},
		isPending: false
	};
	const [togglingToolId, setTogglingToolId] = useState<string | null>(null);



	const handleCreateTool = () => {
		setFormData({
			name: '',
			function_name: '',
			category: 'general',
			description: '',
			implementation_type: 'internal',
			schema: {},
			is_public: false,
			is_active: true
		});
		setEditingTool(null);
		setCreateModalOpen(true);
	};

	const handleViewTool = (tool: Tool) => {
		setViewingTool(tool);
		setViewModalOpen(true);
	};

	const handleEditTool = (tool: Tool) => {
		setFormData({
			name: tool.name,
			function_name: tool.function_name,
			category: tool.category,
			description: tool.description || '',
			implementation_type: tool.implementation_type,
			is_public: Boolean(tool.is_public),
			is_active: Boolean(tool.is_active)
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

	const handleDeleteTool = async (tool: Tool) => {
		if (confirm(`Are you sure you want to delete "${tool.name}"? This action cannot be undone.`)) {
			try {
				await deleteTool.mutateAsync(tool.id);
			} catch (error) {
				// Error handling is done in the mutation hook
				console.error('Failed to delete tool:', error);
			}
		}
	};

	const handleExecuteTool = (tool: Tool) => {
		enqueueSnackbar(`Execute functionality coming soon for ${tool.name}`, { variant: 'info' });
	};

	const handleToggleStatus = useCallback(async (tool: Tool) => {
		setTogglingToolId(tool.id);
		try {
			const updateData: UpdateToolRequest = {
				is_active: !tool.is_active
			};
			await updateTool.mutateAsync({
				id: tool.id,
				data: updateData
			});
		} catch (error) {
			console.error('Failed to update tool status:', error);
		} finally {
			setTogglingToolId(null);
		}
	}, [updateTool, enqueueSnackbar]);

	const columns = useMemo<MRT_ColumnDef<Tool>[]>(
		() => [
			{
				accessorKey: 'is_active',
				header: 'Active',
				size: 80,
				Cell: ({ row }) => (
					<Tooltip title={row.original.is_active ? 'Deactivate tool' : 'Activate tool'}>
						<Switch
							checked={!!row.original.is_active}
							onChange={() => handleToggleStatus(row.original)}
							disabled={togglingToolId === row.original.id}
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
		[handleToggleStatus, togglingToolId]
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
						data={tools}
						enableRowActions
						state={{ isLoading: isLoading }}
						renderRowActions={({ row }) => (
							<Box className="flex items-center space-x-1">
								<Tooltip title="View Details">
									<IconButton
										size="small"
										onClick={() => handleViewTool(row.original)}
									>
										<SvgIcon size={18}>lucide:eye</SvgIcon>
									</IconButton>
								</Tooltip>
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
								/>
								<TextField
									label="Category"
									value={formData.category}
									onChange={(e) => setFormData((prev) => ({ ...prev, category: e.target.value }))}
									select
									fullWidth
									required
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
								/>
								<FormControlLabel
									control={
										<Switch
											checked={Boolean(formData.is_public)}
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
											checked={Boolean(formData.is_active)}
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

					{/* View Tool Dialog */}
					<Dialog
						open={viewModalOpen}
						onClose={() => setViewModalOpen(false)}
						maxWidth="md"
						fullWidth
					>
						<DialogTitle>Tool Details</DialogTitle>
						<DialogContent>
							{viewingTool && (
								<Stack spacing={3} sx={{ mt: 1 }}>
									<Card variant="outlined">
										<CardContent>
											<Stack spacing={2}>
												<Box>
													<Typography variant="subtitle2" color="textSecondary">
														Name
													</Typography>
													<Typography variant="body1">{viewingTool.name}</Typography>
												</Box>

												<Box>
													<Typography variant="subtitle2" color="textSecondary">
														Function Name
													</Typography>
													<Typography variant="body2" className="font-mono">
														{viewingTool.function_name}
													</Typography>
												</Box>

												<Box>
													<Typography variant="subtitle2" color="textSecondary">
														Category
													</Typography>
													<Chip
														size="small"
														label={viewingTool.category}
														color={CATEGORY_COLORS[viewingTool.category as keyof typeof CATEGORY_COLORS] || 'default'}
													/>
												</Box>

												<Box>
													<Typography variant="subtitle2" color="textSecondary">
														Implementation Type
													</Typography>
													<Chip
														size="small"
														label={viewingTool.implementation_type}
														color={viewingTool.implementation_type === 'internal' ? 'primary' : 'secondary'}
													/>
												</Box>

												{viewingTool.description && (
													<Box>
														<Typography variant="subtitle2" color="textSecondary">
															Description
														</Typography>
														<Typography variant="body1">
															{viewingTool.description}
														</Typography>
													</Box>
												)}

												{viewingTool.schema && Object.keys(viewingTool.schema).length > 0 && (
													<Box>
														<Typography variant="subtitle2" color="textSecondary">
															Schema
														</Typography>
														<Card variant="outlined" sx={{ bgcolor: 'grey.50' }}>
															<CardContent>
																<Typography variant="body2" component="pre" sx={{ whiteSpace: 'pre-wrap' }}>
																	{JSON.stringify(viewingTool.schema, null, 2)}
																</Typography>
															</CardContent>
														</Card>
													</Box>
												)}

												<Box>
													<Typography variant="subtitle2" color="textSecondary">
														Usage Count
													</Typography>
													<Typography variant="body1">
														{viewingTool.usage_count || 0} times
													</Typography>
												</Box>

												<Box sx={{ display: 'flex', gap: 2 }}>
													<Box>
														<Typography variant="subtitle2" color="textSecondary">
															Status
														</Typography>
														<Chip
															size="small"
															label={viewingTool.is_active ? 'Active' : 'Inactive'}
															color={viewingTool.is_active ? 'success' : 'default'}
														/>
													</Box>
													<Box>
														<Typography variant="subtitle2" color="textSecondary">
															Visibility
														</Typography>
														<Chip
															size="small"
															label={viewingTool.is_public ? 'Public' : 'Private'}
															color={viewingTool.is_public ? 'info' : 'default'}
														/>
													</Box>
												</Box>

												{(viewingTool.timeout_seconds || viewingTool.max_retries) && (
													<Box>
														<Typography variant="subtitle2" color="textSecondary">
															Configuration
														</Typography>
														<Stack direction="row" spacing={2}>
															{viewingTool.timeout_seconds && (
																<Typography variant="body2">
																	Timeout: {viewingTool.timeout_seconds}s
																</Typography>
															)}
															{viewingTool.max_retries && (
																<Typography variant="body2">
																	Max Retries: {viewingTool.max_retries}
																</Typography>
															)}
														</Stack>
													</Box>
												)}

												<Divider />

												<Box>
													<Typography variant="subtitle2" color="textSecondary">
														Created
													</Typography>
													<Typography variant="body2">
														{new Date(viewingTool.created_at).toLocaleString()}
													</Typography>
												</Box>

												<Box>
													<Typography variant="subtitle2" color="textSecondary">
														Last Updated
													</Typography>
													<Typography variant="body2">
														{new Date(viewingTool.updated_at).toLocaleString()}
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
							{viewingTool && (
								<Button
									variant="contained"
									color="primary"
									onClick={() => {
										setViewModalOpen(false);
										handleEditTool(viewingTool);
									}}
								>
									Edit Tool
								</Button>
							)}
						</DialogActions>
					</Dialog>
				</div>
			}
		/>
	);
}

export default ToolsView;
