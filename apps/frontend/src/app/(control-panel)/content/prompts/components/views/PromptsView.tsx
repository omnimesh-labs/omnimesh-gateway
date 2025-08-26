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

interface Prompt {
	id: string;
	name: string;
	category: string;
	description?: string;
	prompt_template: string;
	usage_count: number;
	is_active: boolean;
	created_at: string;
	updated_at: string;
}

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
	const [editingPrompt, setEditingPrompt] = useState<Prompt | null>(null);
	const [formData, setFormData] = useState({
		name: '',
		category: 'general',
		description: '',
		prompt_template: '',
		is_active: true
	});
	const { enqueueSnackbar } = useSnackbar();

	// Mock data
	const mockPrompts: Prompt[] = [
		{
			id: '1',
			name: 'Code Review',
			category: 'coding',
			description: 'Review code for best practices',
			prompt_template: 'Please review the following code for best practices...',
			usage_count: 42,
			is_active: true,
			created_at: '2024-01-15T10:00:00Z',
			updated_at: '2024-01-20T15:30:00Z'
		},
		{
			id: '2',
			name: 'Data Analysis',
			category: 'analysis',
			description: 'Analyze data patterns and insights',
			prompt_template: 'Analyze the following data and provide insights...',
			usage_count: 28,
			is_active: true,
			created_at: '2024-01-10T09:00:00Z',
			updated_at: '2024-01-18T14:00:00Z'
		},
		{
			id: '3',
			name: 'Creative Writing',
			category: 'creative',
			description: 'Generate creative content',
			prompt_template: 'Write a creative story about...',
			usage_count: 15,
			is_active: false,
			created_at: '2024-01-12T11:00:00Z',
			updated_at: '2024-01-22T16:00:00Z'
		}
	];

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
		const action = editingPrompt ? 'updated' : 'created';
		enqueueSnackbar(`Prompt ${action} successfully (demo)`, { variant: 'success' });
		setCreateModalOpen(false);
		setEditingPrompt(null);
	};

	const handleDeletePrompt = (prompt: Prompt) => {
		enqueueSnackbar(`Delete functionality coming soon for ${prompt.name}`, { variant: 'info' });
	};

	const handleTestPrompt = (prompt: Prompt) => {
		enqueueSnackbar(`Test functionality coming soon for ${prompt.name}`, { variant: 'info' });
	};

	const columns = useMemo<MRT_ColumnDef<Prompt>[]>(
		() => [
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
						data={mockPrompts}
						enableRowActions
						renderRowActions={({ row }) => (
							<Box className="flex items-center space-x-1">
								<Tooltip title="Test Prompt">
									<IconButton
										size="small"
										onClick={() => handleTestPrompt(row.original)}
									>
										<SvgIcon size={18}>lucide:play</SvgIcon>
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
				</div>
			}
		/>
	);
}

export default PromptsView;
