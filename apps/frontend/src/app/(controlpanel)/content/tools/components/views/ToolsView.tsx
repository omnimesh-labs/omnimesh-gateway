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
import { useTools, useCreateTool, useUpdateTool, useDeleteTool, useExecuteTool } from '../../api/hooks/useTools';
import type { Tool, CreateToolRequest, UpdateToolRequest } from '@/lib/types';

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

	const { data: tools = [], isLoading } = useTools();
	const createTool = useCreateTool();
	const updateTool = useUpdateTool();
	const deleteTool = useDeleteTool();
	const executeTool = useExecuteTool();

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
			schema: tool.schema || {},
			is_public: Boolean(tool.is_public),
			is_active: Boolean(tool.is_active)
		});
		setEditingTool(tool);
		setCreateModalOpen(true);
	};

	const handleSaveTool = () => {
		if (editingTool) {
			// Update existing tool
			const updateData: UpdateToolRequest = {
				name: formData.name,
				function_name: formData.function_name,
				category: formData.category,
				description: formData.description,
				implementation_type: formData.implementation_type,
				schema: formData.schema,
				is_public: formData.is_public,
				is_active: formData.is_active
			};
			updateTool.mutate({
				id: editingTool.id,
				data: updateData
			}, {
				onSuccess: () => {
					enqueueSnackbar('Tool updated successfully', { variant: 'success' });
					setCreateModalOpen(false);
					setEditingTool(null);
				},
				onError: (error) => {
					enqueueSnackbar(`Failed to update tool: ${error.message}`, { variant: 'error' });
				}
			});
		} else {
			// Create new tool
			const createData: CreateToolRequest = {
				name: formData.name,
				function_name: formData.function_name,
				category: formData.category,
				description: formData.description,
				implementation_type: formData.implementation_type,
				schema: formData.schema || {},
				is_public: formData.is_public
			};
			createTool.mutate(createData, {
				onSuccess: () => {
					enqueueSnackbar('Tool created successfully', { variant: 'success' });
					setCreateModalOpen(false);
					setEditingTool(null);
				},
				onError: (error) => {
					enqueueSnackbar(`Failed to create tool: ${error.message}`, { variant: 'error' });
				}
			});
		}
	};

	const handleDeleteTool = async (tool: Tool) => {
		if (confirm(`Are you sure you want to delete "${tool.name}"? This action cannot be undone.`)) {
			deleteTool.mutate(tool.id, {
				onSuccess: () => {
					enqueueSnackbar('Tool deleted successfully', { variant: 'success' });
				},
				onError: (error) => {
					enqueueSnackbar(`Failed to delete tool: ${error.message}`, { variant: 'error' });
				}
			});
		}
	};

	const handleExecuteTool = (tool: Tool) => {
		executeTool.mutate(tool.id, {
			onSuccess: () => {
				enqueueSnackbar(`Tool "${tool.name}" executed successfully`, { variant: 'success' });
			},
			onError: (error) => {
				enqueueSnackbar(`Failed to execute tool: ${error.message}`, { variant: 'error' });
			}
		});
	};

	const handleToggleStatus = useCallback(async (tool: Tool) => {
		setTogglingToolId(tool.id);
		const updateData: UpdateToolRequest = {
			is_active: !tool.is_active
		};
		updateTool.mutate({
			id: tool.id,
			data: updateData
		}, {
			onSuccess: () => {
				enqueueSnackbar(`Tool ${tool.is_active ? 'deactivated' : 'activated'} successfully`, { variant: 'success' });
			},
			onError: (error) => {
				enqueueSnackbar(`Failed to update tool status: ${error.message}`, { variant: 'error' });
			},
			onSettled: () => {
				setTogglingToolId(null);
			}
		});
	}, [updateTool]);

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
					const isInactive = !row.original.is_active;
					return (
						<Box className="flex items-center space-x-2" sx={{ opacity: isInactive ? 0.6 : 1 }}>
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
				Cell: ({ cell, row }) => {
					const color = CATEGORY_COLORS[cell.getValue<string>() as keyof typeof CATEGORY_COLORS] || 'default';
					const isInactive = !row.original.is_active;
					return (
						<Box sx={{ opacity: isInactive ? 0.6 : 1 }}>
							<Chip
								size="small"
								label={cell.getValue<string>()}
								color={color}
							/>
						</Box>
					);
				}
			},
			{
				accessorKey: 'implementation_type',
				header: 'Type',
				size: 120,
				Cell: ({ cell, row }) => {
					const isInactive = !row.original.is_active;
					return (
						<Box sx={{ opacity: isInactive ? 0.6 : 1 }}>
							<Chip
								size="small"
								label={cell.getValue<string>()}
								variant="outlined"
							/>
						</Box>
					);
				}
			},
			{
				accessorKey: 'usage_count',
				header: 'Usage',
				size: 100,
				Cell: ({ cell, row }) => {
					const isInactive = !row.original.is_active;
					return (
						<Box sx={{ opacity: isInactive ? 0.6 : 1 }}>
							<Chip
								size="small"
								label={cell.getValue<number>() || 0}
								variant="outlined"
							/>
						</Box>
					);
				}
			},
			{
				accessorKey: 'is_public',
				header: 'Access',
				size: 100,
				Cell: ({ cell, row }) => {
					const isInactive = !row.original.is_active;
					return (
						<Box className="flex items-center" sx={{ opacity: isInactive ? 0.6 : 1 }}>
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
					);
				}
			},
			{
				accessorKey: 'created_at',
				header: 'Created',
				size: 150,
				Cell: ({ cell, row }) => {
					const date = new Date(cell.getValue<string>());
					const isInactive = !row.original.is_active;
					return (
						<Typography
							variant="body2"
							sx={{ opacity: isInactive ? 0.6 : 1 }}
						>
							{date.toLocaleDateString('en-US', {
								year: 'numeric',
								month: 'short',
								day: 'numeric'
							})}
						</Typography>
					);
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
								<TextField
									label="Schema (JSON)"
									value={JSON.stringify(formData.schema || {}, null, 2)}
									onChange={(e) => {
										try {
											const schema = JSON.parse(e.target.value);
											setFormData((prev) => ({ ...prev, schema }));
										} catch {
											// Invalid JSON, don't update
										}
									}}
									fullWidth
									multiline
									rows={6}
									placeholder='{"type": "object", "properties": {...}, "required": [...]}'
									helperText="Define the JSON schema for tool parameters"
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
								disabled={
									!formData.name.trim() ||
									!formData.function_name.trim() ||
									createTool.isPending ||
									updateTool.isPending
								}
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
