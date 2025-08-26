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
	Tabs,
	Tab,
	Dialog,
	DialogTitle,
	DialogContent,
	DialogActions
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

interface ContentItem {
	id: string;
	name: string;
	description?: string;
	category: string;
	usage_count: number;
	created_at: string;
	tags?: string[];
}

function ContentView() {
	const [tabValue, setTabValue] = useState(0);
	const [selectedItem, setSelectedItem] = useState<ContentItem | null>(null);
	const [viewDialogOpen, setViewDialogOpen] = useState(false);
	const { enqueueSnackbar } = useSnackbar();

	// Mock data for demonstration - following the same pattern as servers
	const mockTools: ContentItem[] = [
		{
			id: '1',
			name: 'file_operations',
			description: 'Tools for file system operations',
			category: 'filesystem',
			usage_count: 245,
			created_at: '2024-01-15T10:00:00Z',
			tags: ['files', 'system']
		},
		{
			id: '2',
			name: 'api_caller',
			description: 'Make HTTP API calls',
			category: 'network',
			usage_count: 186,
			created_at: '2024-01-20T15:30:00Z',
			tags: ['api', 'http']
		}
	];

	const mockPrompts: ContentItem[] = [
		{
			id: '1',
			name: 'code_review',
			description: 'Review code for best practices',
			category: 'development',
			usage_count: 98,
			created_at: '2024-01-10T09:00:00Z',
			tags: ['code', 'review']
		},
		{
			id: '2',
			name: 'documentation',
			description: 'Generate technical documentation',
			category: 'writing',
			usage_count: 156,
			created_at: '2024-01-18T14:00:00Z',
			tags: ['docs', 'technical']
		}
	];

	const mockResources: ContentItem[] = [
		{
			id: '1',
			name: 'api_documentation',
			description: 'REST API documentation',
			category: 'documentation',
			usage_count: 234,
			created_at: '2024-01-12T11:00:00Z',
			tags: ['api', 'docs']
		},
		{
			id: '2',
			name: 'config_templates',
			description: 'Configuration file templates',
			category: 'templates',
			usage_count: 89,
			created_at: '2024-01-22T16:00:00Z',
			tags: ['config', 'templates']
		}
	];

	const getCurrentData = () => {
		switch (tabValue) {
			case 0:
				return mockTools;
			case 1:
				return mockPrompts;
			case 2:
				return mockResources;
			default:
				return [];
		}
	};

	const handleViewItem = (item: ContentItem) => {
		setSelectedItem(item);
		setViewDialogOpen(true);
	};

	const columns = useMemo<MRT_ColumnDef<ContentItem>[]>(
		() => [
			{
				accessorKey: 'name',
				header: 'Name',
				size: 200,
				Cell: ({ row }) => (
					<Box className="flex items-center space-x-2">
						<SvgIcon size={20}>
							{tabValue === 0
								? 'lucide:wrench'
								: tabValue === 1
									? 'lucide:message-square'
									: 'lucide:database'}
						</SvgIcon>
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
				)
			},
			{
				accessorKey: 'category',
				header: 'Category',
				size: 120,
				Cell: ({ cell }) => (
					<Chip
						size="small"
						label={cell.getValue<string>()}
						variant="outlined"
						sx={{ textTransform: 'capitalize' }}
					/>
				)
			},
			{
				accessorKey: 'usage_count',
				header: 'Usage Count',
				size: 120
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
		[tabValue]
	);

	const currentData = getCurrentData();
	const currentType = ['Tools', 'Prompts', 'Resources'][tabValue];

	return (
		<Root
			header={
				<div className="p-6">
					<div className="flex items-center justify-between">
						<div>
							<Typography variant="h4">Content Management</Typography>
							<Typography
								variant="body1"
								color="textSecondary"
								className="mt-1"
							>
								Manage your organization's tools, prompts, and resources
							</Typography>
						</div>
						<Button
							variant="contained"
							color="primary"
							startIcon={<SvgIcon>lucide:plus</SvgIcon>}
							onClick={() => enqueueSnackbar('Create functionality coming soon!', { variant: 'info' })}
						>
							Create {currentType.slice(0, -1)}
						</Button>
					</div>
				</div>
			}
			content={
				<div className="p-6">
					<Box className="mb-6">
						<Tabs
							value={tabValue}
							onChange={(_, newValue) => setTabValue(newValue)}
							variant="scrollable"
							scrollButtons="auto"
						>
							<Tab
								label={`Tools (${mockTools.length})`}
								icon={<SvgIcon size={20}>lucide:wrench</SvgIcon>}
								iconPosition="start"
							/>
							<Tab
								label={`Prompts (${mockPrompts.length})`}
								icon={<SvgIcon size={20}>lucide:message-square</SvgIcon>}
								iconPosition="start"
							/>
							<Tab
								label={`Resources (${mockResources.length})`}
								icon={<SvgIcon size={20}>lucide:database</SvgIcon>}
								iconPosition="start"
							/>
						</Tabs>
					</Box>

					<DataTable
						columns={columns}
						data={currentData}
						enableRowActions
						renderRowActions={({ row }) => (
							<Box className="flex items-center space-x-1">
								<Tooltip title="View Details">
									<IconButton
										size="small"
										onClick={() => handleViewItem(row.original)}
									>
										<SvgIcon size={18}>lucide:eye</SvgIcon>
									</IconButton>
								</Tooltip>
								<Tooltip title="Edit">
									<IconButton
										size="small"
										onClick={() =>
											enqueueSnackbar('Edit functionality coming soon!', { variant: 'info' })
										}
									>
										<SvgIcon size={18}>lucide:edit</SvgIcon>
									</IconButton>
								</Tooltip>
								<Tooltip title="Delete">
									<IconButton
										size="small"
										color="error"
										onClick={() =>
											enqueueSnackbar('Delete functionality coming soon!', { variant: 'info' })
										}
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

					{/* View Item Dialog */}
					<Dialog
						open={viewDialogOpen}
						onClose={() => setViewDialogOpen(false)}
						maxWidth="md"
						fullWidth
					>
						<DialogTitle>{selectedItem?.name}</DialogTitle>
						<DialogContent>
							{selectedItem && (
								<Box className="space-y-4">
									<Box>
										<Typography
											variant="subtitle2"
											color="textSecondary"
										>
											Description
										</Typography>
										<Typography variant="body2">
											{selectedItem.description || 'No description provided'}
										</Typography>
									</Box>
									<Box>
										<Typography
											variant="subtitle2"
											color="textSecondary"
										>
											Category
										</Typography>
										<Chip
											size="small"
											label={selectedItem.category}
										/>
									</Box>
									<Box>
										<Typography
											variant="subtitle2"
											color="textSecondary"
										>
											Usage Count
										</Typography>
										<Typography variant="body2">{selectedItem.usage_count}</Typography>
									</Box>
									<Box>
										<Typography
											variant="subtitle2"
											color="textSecondary"
										>
											Tags
										</Typography>
										<Box className="mt-1 flex gap-1">
											{selectedItem.tags?.map((tag) => (
												<Chip
													key={tag}
													size="small"
													label={tag}
													variant="outlined"
												/>
											))}
										</Box>
									</Box>
								</Box>
							)}
						</DialogContent>
						<DialogActions>
							<Button onClick={() => setViewDialogOpen(false)}>Close</Button>
						</DialogActions>
					</Dialog>
				</div>
			}
		/>
	);
}

export default ContentView;
