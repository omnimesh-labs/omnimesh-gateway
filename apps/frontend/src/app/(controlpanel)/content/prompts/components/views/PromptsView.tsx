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
import { usePrompts, useCreatePrompt, useUpdatePrompt, useDeletePrompt } from '../../api/hooks/usePrompts';
import type { Prompt, CreatePromptRequest, UpdatePromptRequest } from '@/lib/types';

// Form data type that includes both create and edit fields
interface PromptFormData {
	name: string;
	category: string;
	description: string;
	prompt_template: string;
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
	const [formData, setFormData] = useState<PromptFormData>({
		name: '',
		category: 'general',
		description: '',
		prompt_template: '',
		is_active: true
	});

	const { enqueueSnackbar } = useSnackbar();

	const { data: prompts = [], isLoading, error } = usePrompts();
	const createPrompt = useCreatePrompt();
	const updatePrompt = useUpdatePrompt();
	const deletePrompt = useDeletePrompt();
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

	const handleSavePrompt = () => {
		if (editingPrompt) {
			// Update existing prompt
			const updateData: UpdatePromptRequest = {
				name: formData.name,
				category: formData.category,
				description: formData.description,
				prompt_template: formData.prompt_template,
				is_active: formData.is_active
			};
			updatePrompt.mutate({
				id: editingPrompt.id,
				data: updateData
			}, {
				onSuccess: () => {
					enqueueSnackbar('Prompt updated successfully', { variant: 'success' });
					setCreateModalOpen(false);
					setEditingPrompt(null);
				},
				onError: (error) => {
					enqueueSnackbar(`Failed to update prompt: ${error.message}`, { variant: 'error' });
				}
			});
		} else {
			// Create new prompt
			const createData: CreatePromptRequest = {
				name: formData.name,
				category: formData.category,
				description: formData.description,
				prompt_template: formData.prompt_template
			};
			createPrompt.mutate(createData, {
				onSuccess: () => {
					enqueueSnackbar('Prompt created successfully', { variant: 'success' });
					setCreateModalOpen(false);
					setEditingPrompt(null);
				},
				onError: (error) => {
					enqueueSnackbar(`Failed to create prompt: ${error.message}`, { variant: 'error' });
				}
			});
		}
	};

	const handleDeletePrompt = (prompt: Prompt) => {
		if (confirm(`Are you sure you want to delete "${prompt.name}"? This action cannot be undone.`)) {
			deletePrompt.mutate(prompt.id, {
				onSuccess: () => {
					enqueueSnackbar('Prompt deleted successfully', { variant: 'success' });
				},
				onError: (error) => {
					enqueueSnackbar(`Failed to delete prompt: ${error.message}`, { variant: 'error' });
				}
			});
		}
	};

	const handleTogglePromptStatus = useCallback((prompt: Prompt) => {
		setTogglingPromptId(prompt.id);
		const updateData: UpdatePromptRequest = {
			is_active: !prompt.is_active
		};
		updatePrompt.mutate({
			id: prompt.id,
			data: updateData
		}, {
			onSuccess: () => {
				enqueueSnackbar(`Prompt ${prompt.is_active ? 'deactivated' : 'activated'} successfully`, { variant: 'success' });
			},
			onError: (error) => {
				enqueueSnackbar(`Failed to update prompt status: ${error.message}`, { variant: 'error' });
			},
			onSettled: () => {
				setTogglingPromptId(null);
			}
		});
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
				accessorKey: 'prompt_template',
				header: 'Template',
				size: 300,
				Cell: ({ cell, row }) => {
					const template = cell.getValue<string>();
					const truncated = template.length > 100 ? template.substring(0, 100) + '...' : template;
					const isInactive = !row.original.is_active;
					return (
						<Typography
							variant="body2"
							color="textSecondary"
							title={template}
							sx={{ opacity: isInactive ? 0.6 : 1 }}
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
								disabled={
									!formData.name.trim() ||
									!formData.prompt_template.trim() ||
									createPrompt.isPending ||
									updatePrompt.isPending
								}
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
