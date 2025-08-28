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
interface Prompt {
	id: string;
	name: string;
	category: string;
	description: string;
	prompt_template: string;
	is_active: boolean;
	created_at: string;
	updated_at: string;
}

interface CreatePromptRequest {
	name: string;
	category: string;
	description: string;
	prompt_template: string;
	is_active: boolean;
}

interface UpdatePromptRequest extends CreatePromptRequest {
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


const CATEGORY_ICONS = {
	general: 'lucide:message-square',
	coding: 'lucide:code',
	analysis: 'lucide:sparkles',
	creative: 'lucide:palette',
	educational: 'lucide:book-open',
	business: 'lucide:briefcase',
	custom: 'lucide:message-square'
};

const CATEGORY_COLORS = {
	general: 'default',
	coding: 'primary',
	analysis: 'secondary',
	creative: 'info',
	educational: 'success',
	business: 'warning',
	custom: 'default'
} as const;

function PromptsView() {
	const [createModalOpen, setCreateModalOpen] = useState(false);
	const [viewModalOpen, setViewModalOpen] = useState(false);
	const [editingPrompt, setEditingPrompt] = useState<Prompt | null>(null);
	const [viewingPrompt, setViewingPrompt] = useState<Prompt | null>(null);
	const [formData, setFormData] = useState<CreatePromptRequest>({
		name: '',
		category: 'general',
		description: '',
		prompt_template: '',
		is_active: true
	});

	const { enqueueSnackbar } = useSnackbar();

	// Mock data
	const [prompts, setPrompts] = useState<Prompt[]>([
		{
			id: '1',
			name: 'Code Review Assistant',
			category: 'coding',
			description: 'Helps review code for best practices and issues',
			prompt_template: 'Please review the following code for:\n- Best practices\n- Potential bugs\n- Performance issues\n- Security concerns\n\n{code}',
			is_active: true,
			created_at: '2024-01-01T00:00:00Z',
			updated_at: '2024-01-01T00:00:00Z'
		},
		{
			id: '2',
			name: 'Data Analysis Helper',
			category: 'analysis',
			description: 'Analyzes data and provides insights',
			prompt_template: 'Analyze the following data and provide insights:\n\n{data}',
			is_active: true,
			created_at: '2024-01-01T00:00:00Z',
			updated_at: '2024-01-01T00:00:00Z'
		}
	]);
	const isLoading = false;
	const error = null;

	// Mock functions
	const createPrompt = {
		mutate: (data: CreatePromptRequest) => {
			const newPrompt: Prompt = {
				id: Date.now().toString(),
				...data,
				created_at: new Date().toISOString(),
				updated_at: new Date().toISOString()
			};
			setPrompts(prev => [...prev, newPrompt]);
			enqueueSnackbar('Prompt created successfully', { variant: 'success' });
		},
		isPending: false
	};

	const updatePrompt = {
		mutate: (data: UpdatePromptRequest) => {
			setPrompts(prev => prev.map(p => p.id === data.id ? { ...p, ...data, updated_at: new Date().toISOString() } : p));
			enqueueSnackbar('Prompt updated successfully', { variant: 'success' });
		},
		isPending: false
	};

	const deletePrompt = {
		mutate: (id: string) => {
			setPrompts(prev => prev.filter(p => p.id !== id));
			enqueueSnackbar('Prompt deleted successfully', { variant: 'success' });
		},
		isPending: false
	};
	const [togglingPromptId, setTogglingPromptId] = useState<string | null>(null);


	const handleCreatePrompt = () => {
		setFormData({
			name: '',
			category: 'general',
			description: '',
			prompt_template: '',
			is_active: true
		});
		setEditingPrompt(null);
		setCreateModalOpen(true);
	};

	const handleViewPrompt = (prompt: Prompt) => {
		setViewingPrompt(prompt);
		setViewModalOpen(true);
	};

	const handleEditPrompt = (prompt: Prompt) => {
		setFormData({
			name: prompt.name,
			category: prompt.category,
			description: prompt.description || '',
			prompt_template: prompt.prompt_template,
			is_active: prompt.is_active
		});
		setEditingPrompt(prompt);
		setCreateModalOpen(true);
	};

	const handleSavePrompt = async () => {
		try {
			if (editingPrompt) {
				const updateData: UpdatePromptRequest = {
					name: formData.name,
					category: formData.category,
					description: formData.description,
					prompt_template: formData.prompt_template
				};
				await updatePrompt.mutateAsync({
					id: editingPrompt.id,
					data: updateData
				});
			} else {
				await createPrompt.mutateAsync(formData);
			}
			setCreateModalOpen(false);
			setEditingPrompt(null);
		} catch (error) {
			// Error handling is done in the mutation hooks
			console.error('Failed to save prompt:', error);
		}
	};

	const handleDeletePrompt = async (prompt: Prompt) => {
		if (confirm(`Are you sure you want to delete "${prompt.name}"? This action cannot be undone.`)) {
			try {
				await deletePrompt.mutateAsync(prompt.id);
			} catch (error) {
				// Error handling is done in the mutation hook
				console.error('Failed to delete prompt:', error);
			}
		}
	};

	const handleTogglePromptStatus = useCallback(async (prompt: Prompt) => {
		setTogglingPromptId(prompt.id);
		try {
			const updateData: UpdatePromptRequest = {
				is_active: !prompt.is_active
			};
			const updatedPrompt = await updatePrompt.mutateAsync({
				id: prompt.id,
				data: updateData
			});
		} catch (error) {
			console.error('Failed to update prompt status:', error);
		} finally {
			setTogglingPromptId(null);
		}
	}, [updatePrompt]);


	const columns = useMemo<MRT_ColumnDef<Prompt>[]>(
		() => [
			{
				accessorKey: 'is_active',
				header: 'Active',
				size: 80,
				Cell: ({ row }) => (
					<Tooltip title={row.original.is_active ? 'Deactivate prompt' : 'Activate prompt'}>
						<Switch
							checked={!!row.original.is_active}
							onChange={() => handleTogglePromptStatus(row.original)}
							disabled={togglingPromptId === row.original.id}
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
						CATEGORY_ICONS[row.original.category as keyof typeof CATEGORY_ICONS] || 'lucide:message-square';
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
				accessorKey: 'category',
				header: 'Category',
				size: 150,
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
				accessorKey: 'prompt_template',
				header: 'Template',
				size: 300,
				Cell: ({ cell }) => {
					const template = cell.getValue<string>();
					const truncated = template.length > 100 ? template.substring(0, 100) + '...' : template;
					return (
						<Typography
							variant="body2"
							color="textSecondary"
							title={template}
						>
							{truncated}
						</Typography>
					);
				}
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
		[handleTogglePromptStatus, togglingPromptId]
	);

	return (
		<Root
			header={
				<div className="p-6">
					<div className="flex items-center justify-between">
						<div>
							<Typography variant="h4">Prompts</Typography>
							<Typography
								variant="body1"
								color="textSecondary"
								className="mt-1"
							>
								Manage your prompt templates for MCP interactions
							</Typography>
						</div>
						<Button
							variant="contained"
							color="primary"
							startIcon={<SvgIcon>lucide:plus</SvgIcon>}
							onClick={handleCreatePrompt}
						>
							Create Prompt
						</Button>
					</div>
				</div>
			}
			content={
				<div className="p-6">
					<LazyDataTable
						columns={columns}
						data={prompts}
						state={{ isLoading: isLoading }}
						enableRowActions
						renderRowActions={({ row }) => (
							<Box className="flex items-center space-x-1">
								<Tooltip title="View Details">
									<IconButton
										size="small"
										onClick={() => handleViewPrompt(row.original)}
									>
										<SvgIcon size={18}>lucide:eye</SvgIcon>
									</IconButton>
								</Tooltip>
								<Tooltip title="Edit Prompt">
									<IconButton
										size="small"
										onClick={() => handleEditPrompt(row.original)}
									>
										<SvgIcon size={18}>lucide:edit</SvgIcon>
									</IconButton>
								</Tooltip>
								<Tooltip title="Delete Prompt">
									<IconButton
										size="small"
										color="error"
										onClick={() => handleDeletePrompt(row.original)}
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

					{/* Create/Edit Prompt Dialog */}
					<Dialog
						open={createModalOpen}
						onClose={() => setCreateModalOpen(false)}
						maxWidth="md"
						fullWidth
					>
						<DialogTitle>{editingPrompt ? 'Edit Prompt' : 'Create Prompt'}</DialogTitle>
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
									label="Category"
									value={formData.category}
									onChange={(e) => setFormData((prev) => ({ ...prev, category: e.target.value }))}
									select
									fullWidth
									required
								>
									<MenuItem value="general">General</MenuItem>
									<MenuItem value="coding">Coding</MenuItem>
									<MenuItem value="analysis">Analysis</MenuItem>
									<MenuItem value="creative">Creative</MenuItem>
									<MenuItem value="educational">Educational</MenuItem>
									<MenuItem value="business">Business</MenuItem>
									<MenuItem value="custom">Custom</MenuItem>
								</TextField>
								<TextField
									label="Description"
									value={formData.description}
									onChange={(e) => setFormData((prev) => ({ ...prev, description: e.target.value }))}
									fullWidth
									multiline
									rows={2}
								/>
								<TextField
									label="Prompt Template"
									value={formData.prompt_template}
									onChange={(e) =>
										setFormData((prev) => ({ ...prev, prompt_template: e.target.value }))
									}
									fullWidth
									multiline
									rows={4}
									required
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
								onClick={handleSavePrompt}
								disabled={!formData.name.trim() || !formData.prompt_template.trim()}
							>
								{editingPrompt ? 'Update' : 'Create'}
							</Button>
						</DialogActions>
					</Dialog>

					{/* View Prompt Dialog */}
					<Dialog
						open={viewModalOpen}
						onClose={() => setViewModalOpen(false)}
						maxWidth="md"
						fullWidth
					>
						<DialogTitle>Prompt Details</DialogTitle>
						<DialogContent>
							{viewingPrompt && (
								<Stack spacing={3} sx={{ mt: 1 }}>
									<Card variant="outlined">
										<CardContent>
											<Stack spacing={2}>
												<Box>
													<Typography variant="subtitle2" color="textSecondary">
														Name
													</Typography>
													<Typography variant="body1">{viewingPrompt.name}</Typography>
												</Box>

												<Box>
													<Typography variant="subtitle2" color="textSecondary">
														Category
													</Typography>
													<Chip
														size="small"
														label={viewingPrompt.category}
														color={CATEGORY_COLORS[viewingPrompt.category as keyof typeof CATEGORY_COLORS] || 'default'}
													/>
												</Box>

												{viewingPrompt.description && (
													<Box>
														<Typography variant="subtitle2" color="textSecondary">
															Description
														</Typography>
														<Typography variant="body1">
															{viewingPrompt.description}
														</Typography>
													</Box>
												)}

												<Box>
													<Typography variant="subtitle2" color="textSecondary">
														Prompt Template
													</Typography>
													<Card variant="outlined" sx={{ bgcolor: 'grey.50' }}>
														<CardContent>
															<Typography variant="body2" component="pre" sx={{ whiteSpace: 'pre-wrap' }}>
																{viewingPrompt.prompt_template}
															</Typography>
														</CardContent>
													</Card>
												</Box>

												<Box>
													<Typography variant="subtitle2" color="textSecondary">
														Status
													</Typography>
													<Chip
														size="small"
														label={viewingPrompt.is_active ? 'Active' : 'Inactive'}
														color={viewingPrompt.is_active ? 'success' : 'default'}
													/>
												</Box>

												<Divider />

												<Box>
													<Typography variant="subtitle2" color="textSecondary">
														Created
													</Typography>
													<Typography variant="body2">
														{new Date(viewingPrompt.created_at).toLocaleString()}
													</Typography>
												</Box>

												<Box>
													<Typography variant="subtitle2" color="textSecondary">
														Last Updated
													</Typography>
													<Typography variant="body2">
														{new Date(viewingPrompt.updated_at).toLocaleString()}
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
							{viewingPrompt && (
								<Button
									variant="contained"
									color="primary"
									onClick={() => {
										setViewModalOpen(false);
										handleEditPrompt(viewingPrompt);
									}}
								>
									Edit Prompt
								</Button>
							)}
						</DialogActions>
					</Dialog>
				</div>
			}
		/>
	);
}

export default PromptsView;
