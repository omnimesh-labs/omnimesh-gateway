'use client';

import { useState, useEffect, useCallback, useMemo } from 'react';
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
	FormControlLabel,
	Switch,
	Select,
	MenuItem,
	FormControl,
	InputLabel,
	Card,
	CardContent,
	Alert,
	CircularProgress,
	Stack,
	GridLegacy as Grid
} from '@mui/material';
import DataTable from '@/components/data-table/DataTable';
import SvgIcon from '@fuse/core/SvgIcon';
import { useSnackbar } from 'notistack';
import { 
	a2aApi, 
	A2AAgent, 
	A2AAgentSpec, 
	A2AStats,
	A2ATestRequest
} from '@/lib/api';

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

function A2AView() {
	const [agents, setAgents] = useState<A2AAgent[]>([]);
	const [stats, setStats] = useState<A2AStats | null>(null);
	const [loading, setLoading] = useState(true);
	const [createDialogOpen, setCreateDialogOpen] = useState(false);
	const [testDialogOpen, setTestDialogOpen] = useState(false);
	const [selectedAgent, setSelectedAgent] = useState<A2AAgent | null>(null);
	const [editingAgent, setEditingAgent] = useState<A2AAgent | null>(null);
	const [showInactive, setShowInactive] = useState(false);
	const [tagFilter, setTagFilter] = useState('');
	const [availableTags, setAvailableTags] = useState<string[]>([]);
	const { enqueueSnackbar } = useSnackbar();

	// Form state for create/edit
	const [formData, setFormData] = useState<Partial<A2AAgentSpec>>({
		name: '',
		description: '',
		agent_type: 'general',
		capabilities: [],
		tags: [],
		is_active: true
	});

	// Test dialog state
	const [testMessage, setTestMessage] = useState('');
	const [testContext, setTestContext] = useState('');
	const [testLoading, setTestLoading] = useState(false);
	const [testResult, setTestResult] = useState<any>(null);

	// Mock data for demonstration
	const mockAgents: A2AAgent[] = [
		{
			id: '1',
			organization_id: 'org-1',
			name: 'Customer Support Agent',
			description: 'Handles customer inquiries and support tickets',
			agent_type: 'support',
			is_active: true,
			capabilities: ['chat', 'ticket_management', 'knowledge_base'],
			tags: ['support', 'customer-facing'],
			created_at: '2024-01-15T10:00:00Z',
			updated_at: '2024-01-20T15:30:00Z',
			metrics: {
				request_count: 1234,
				error_count: 12,
				avg_response_time: 450
			}
		},
		{
			id: '2',
			organization_id: 'org-1',
			name: 'Code Review Agent',
			description: 'Automated code review and suggestions',
			agent_type: 'development',
			is_active: true,
			capabilities: ['code_analysis', 'best_practices', 'security_scan'],
			tags: ['development', 'automation'],
			created_at: '2024-01-10T09:00:00Z',
			updated_at: '2024-01-18T14:00:00Z',
			metrics: {
				request_count: 856,
				error_count: 5,
				avg_response_time: 650
			}
		},
		{
			id: '3',
			organization_id: 'org-1',
			name: 'Data Analysis Agent',
			description: 'Performs data analysis and generates insights',
			agent_type: 'analytics',
			is_active: false,
			capabilities: ['data_processing', 'visualization', 'reporting'],
			tags: ['analytics', 'data'],
			created_at: '2024-01-12T11:00:00Z',
			updated_at: '2024-01-22T16:00:00Z',
			metrics: {
				request_count: 423,
				error_count: 2,
				avg_response_time: 1200
			}
		}
	];

	const mockStats: A2AStats = {
		total: 3,
		active: 2,
		inactive: 1,
		by_type: {
			support: 1,
			development: 1,
			analytics: 1
		}
	};

	const loadAgents = useCallback(async () => {
		try {
			setLoading(true);
			// TODO: Replace with actual API calls when backend is ready
			// const [agentsData, statsData] = await Promise.all([
			//     a2aApi.listAgents({ is_active: showInactive ? undefined : true, tags: tagFilter || undefined }),
			//     a2aApi.getStats()
			// ]);
			
			// Mock implementation
			await new Promise(resolve => setTimeout(resolve, 500));
			let filteredAgents = [...mockAgents];
			
			if (!showInactive) {
				filteredAgents = filteredAgents.filter(a => a.is_active);
			}
			
			if (tagFilter) {
				filteredAgents = filteredAgents.filter(a => 
					a.tags?.some(tag => tag.includes(tagFilter))
				);
			}
			
			setAgents(filteredAgents);
			setStats(mockStats);
			
			// Extract unique tags
			const tags = new Set<string>();
			mockAgents.forEach(agent => {
				agent.tags?.forEach(tag => tags.add(tag));
			});
			setAvailableTags(Array.from(tags).sort());
		} catch (error) {
			console.error('Failed to load agents:', error);
			enqueueSnackbar('Failed to load A2A agents', { variant: 'error' });
		} finally {
			setLoading(false);
		}
	}, [showInactive, tagFilter, enqueueSnackbar]);

	useEffect(() => {
		loadAgents();
	}, [loadAgents]);

	const handleCreateAgent = () => {
		setEditingAgent(null);
		setFormData({
			name: '',
			description: '',
			agent_type: 'general',
			capabilities: [],
			tags: [],
			is_active: true
		});
		setCreateDialogOpen(true);
	};

	const handleEditAgent = (agent: A2AAgent) => {
		setEditingAgent(agent);
		setFormData({
			name: agent.name,
			description: agent.description,
			agent_type: agent.agent_type,
			capabilities: agent.capabilities,
			tags: agent.tags,
			is_active: agent.is_active
		});
		setCreateDialogOpen(true);
	};

	const handleSaveAgent = async () => {
		try {
			if (editingAgent) {
				// TODO: Replace with actual API call
				// await a2aApi.updateAgent(editingAgent.id, formData);
				enqueueSnackbar('Agent updated successfully', { variant: 'success' });
			} else {
				// TODO: Replace with actual API call
				// await a2aApi.createAgent(formData as A2AAgentSpec);
				enqueueSnackbar('Agent created successfully', { variant: 'success' });
			}
			setCreateDialogOpen(false);
			loadAgents();
		} catch (error) {
			enqueueSnackbar('Failed to save agent', { variant: 'error' });
		}
	};

	const handleDeleteAgent = async (id: string) => {
		if (confirm('Are you sure you want to delete this agent?')) {
			try {
				// TODO: Replace with actual API call
				// await a2aApi.deleteAgent(id);
				enqueueSnackbar('Agent deleted successfully', { variant: 'success' });
				loadAgents();
			} catch (error) {
				enqueueSnackbar('Failed to delete agent', { variant: 'error' });
			}
		}
	};

	const handleToggleAgent = async (agent: A2AAgent) => {
		try {
			// TODO: Replace with actual API call
			// await a2aApi.toggleAgent(agent.id, !agent.is_active);
			enqueueSnackbar(
				`Agent ${!agent.is_active ? 'activated' : 'deactivated'} successfully`, 
				{ variant: 'success' }
			);
			loadAgents();
		} catch (error) {
			enqueueSnackbar('Failed to toggle agent status', { variant: 'error' });
		}
	};

	const handleTestAgent = (agent: A2AAgent) => {
		setSelectedAgent(agent);
		setTestMessage('');
		setTestContext('');
		setTestResult(null);
		setTestDialogOpen(true);
	};

	const handleRunTest = async () => {
		if (!selectedAgent || !testMessage) return;
		
		try {
			setTestLoading(true);
			// TODO: Replace with actual API call
			// const result = await a2aApi.testAgent(selectedAgent.id, { message: testMessage, context: testContext });
			
			// Mock result
			await new Promise(resolve => setTimeout(resolve, 1500));
			setTestResult({
				success: true,
				content: `This is a mock response from ${selectedAgent.name}. In production, this would be the actual agent response to: "${testMessage}"`,
				execution_time_ms: 450,
				tokens_used: {
					prompt: 25,
					completion: 50,
					total: 75
				}
			});
		} catch (error) {
			setTestResult({
				success: false,
				error: 'Failed to test agent'
			});
		} finally {
			setTestLoading(false);
		}
	};

	const columns = useMemo<MRT_ColumnDef<A2AAgent>[]>(() => [
		{
			accessorKey: 'name',
			header: 'Agent Name',
			size: 200,
			Cell: ({ row }) => (
				<Box className="flex items-center space-x-2">
					<SvgIcon size={20}>lucide:bot</SvgIcon>
					<Box>
						<Typography variant="body2" className="font-medium">
							{row.original.name}
						</Typography>
						{row.original.description && (
							<Typography variant="caption" color="textSecondary">
								{row.original.description}
							</Typography>
						)}
					</Box>
				</Box>
			)
		},
		{
			accessorKey: 'agent_type',
			header: 'Type',
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
			accessorKey: 'is_active',
			header: 'Status',
			size: 100,
			Cell: ({ cell }) => (
				<Chip
					size="small"
					label={cell.getValue<boolean>() ? 'Active' : 'Inactive'}
					color={cell.getValue<boolean>() ? 'success' : 'default'}
				/>
			)
		},
		{
			accessorKey: 'capabilities',
			header: 'Capabilities',
			size: 200,
			Cell: ({ cell }) => (
				<Box className="flex flex-wrap gap-1">
					{cell.getValue<string[]>()?.slice(0, 2).map(cap => (
						<Chip key={cap} size="small" label={cap} variant="outlined" />
					))}
					{cell.getValue<string[]>()?.length > 2 && (
						<Chip size="small" label={`+${cell.getValue<string[]>().length - 2}`} variant="outlined" />
					)}
				</Box>
			)
		},
		{
			id: 'metrics',
			header: 'Usage',
			size: 150,
			Cell: ({ row }) => (
				<Box>
					<Typography variant="body2">
						{row.original.metrics?.request_count || 0} requests
					</Typography>
					<Typography variant="caption" color="textSecondary">
						{row.original.metrics?.avg_response_time || 0}ms avg
					</Typography>
				</Box>
			)
		},
		{
			accessorKey: 'updated_at',
			header: 'Last Updated',
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
	], []);

	return (
		<Root
			header={
				<div className="p-6">
					<div className="flex items-center justify-between">
						<div>
							<Typography variant="h4">A2A Agent Management</Typography>
							<Typography variant="body1" color="textSecondary" className="mt-1">
								Manage app-to-app authentication agents
							</Typography>
						</div>
						<Button
							variant="contained"
							color="primary"
							startIcon={<SvgIcon>lucide:plus</SvgIcon>}
							onClick={handleCreateAgent}
						>
							Create Agent
						</Button>
					</div>
				</div>
			}
			content={
				<div className="p-6">
					{/* Stats Cards */}
					{stats && (
						<Box sx={{ flexGrow: 1, mb: 3 }}>
							<Grid container spacing={3}>
								<Grid item xs={12} sm={6} md={3}>
								<Card>
									<CardContent>
										<Typography variant="h6">{stats.total}</Typography>
										<Typography variant="body2" color="textSecondary">
											Total Agents
										</Typography>
									</CardContent>
								</Card>
							</Grid>
							<Grid xs={12} sm={6} md={3}>
								<Card>
									<CardContent>
										<Typography variant="h6" color="success.main">
											{stats.active}
										</Typography>
										<Typography variant="body2" color="textSecondary">
											Active Agents
										</Typography>
									</CardContent>
								</Card>
							</Grid>
							<Grid xs={12} sm={6} md={3}>
								<Card>
									<CardContent>
										<Typography variant="h6" color="text.secondary">
											{stats.inactive}
										</Typography>
										<Typography variant="body2" color="textSecondary">
											Inactive Agents
										</Typography>
									</CardContent>
								</Card>
							</Grid>
							<Grid xs={12} sm={6} md={3}>
								<Card>
									<CardContent>
										<Typography variant="h6">
											{Object.keys(stats.by_type).length}
										</Typography>
										<Typography variant="body2" color="textSecondary">
											Agent Types
										</Typography>
									</CardContent>
								</Card>
							</Grid>
							</Grid>
						</Box>
					)}

					{/* Filters */}
					<Box className="mb-4 flex items-center gap-4">
						<FormControlLabel
							control={
								<Switch
									checked={showInactive}
									onChange={(e) => setShowInactive(e.target.checked)}
								/>
							}
							label="Show Inactive"
						/>
						<FormControl size="small" sx={{ minWidth: 200 }}>
							<InputLabel>Filter by Tag</InputLabel>
							<Select
								value={tagFilter}
								label="Filter by Tag"
								onChange={(e) => setTagFilter(e.target.value)}
							>
								<MenuItem value="">All Tags</MenuItem>
								{availableTags.map(tag => (
									<MenuItem key={tag} value={tag}>{tag}</MenuItem>
								))}
							</Select>
						</FormControl>
					</Box>

					{/* Agents Table */}
					{loading ? (
						<Box className="flex justify-center p-8">
							<CircularProgress />
						</Box>
					) : (
						<DataTable
							columns={columns}
							data={agents}
							enableRowActions
							renderRowActions={({ row }) => (
								<Box className="flex items-center space-x-1">
									<Tooltip title="Test Agent">
										<IconButton size="small" onClick={() => handleTestAgent(row.original)}>
											<SvgIcon size={18}>lucide:play</SvgIcon>
										</IconButton>
									</Tooltip>
									<Tooltip title="Edit">
										<IconButton size="small" onClick={() => handleEditAgent(row.original)}>
											<SvgIcon size={18}>lucide:edit</SvgIcon>
										</IconButton>
									</Tooltip>
									<Tooltip title={row.original.is_active ? 'Deactivate' : 'Activate'}>
										<IconButton 
											size="small"
											onClick={() => handleToggleAgent(row.original)}
										>
											<SvgIcon size={18}>
												{row.original.is_active ? 'lucide:pause' : 'lucide:play-circle'}
											</SvgIcon>
										</IconButton>
									</Tooltip>
									<Tooltip title="Delete">
										<IconButton
											size="small"
											color="error"
											onClick={() => handleDeleteAgent(row.original.id)}
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
					)}

					{/* Create/Edit Dialog */}
					<Dialog 
						open={createDialogOpen} 
						onClose={() => setCreateDialogOpen(false)}
						maxWidth="md"
						fullWidth
					>
						<DialogTitle>
							{editingAgent ? 'Edit Agent' : 'Create New Agent'}
						</DialogTitle>
						<DialogContent>
							<Stack spacing={3} sx={{ mt: 1 }}>
								<TextField
									label="Agent Name"
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
								<FormControl fullWidth>
									<InputLabel>Agent Type</InputLabel>
									<Select
										value={formData.agent_type}
										label="Agent Type"
										onChange={(e) => setFormData({ ...formData, agent_type: e.target.value })}
									>
										<MenuItem value="general">General</MenuItem>
										<MenuItem value="support">Support</MenuItem>
										<MenuItem value="development">Development</MenuItem>
										<MenuItem value="analytics">Analytics</MenuItem>
										<MenuItem value="security">Security</MenuItem>
									</Select>
								</FormControl>
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
							<Button onClick={() => setCreateDialogOpen(false)}>Cancel</Button>
							<Button variant="contained" onClick={handleSaveAgent}>
								{editingAgent ? 'Update' : 'Create'}
							</Button>
						</DialogActions>
					</Dialog>

					{/* Test Agent Dialog */}
					<Dialog 
						open={testDialogOpen} 
						onClose={() => setTestDialogOpen(false)}
						maxWidth="md"
						fullWidth
					>
						<DialogTitle>
							Test Agent: {selectedAgent?.name}
						</DialogTitle>
						<DialogContent>
							<Stack spacing={3} sx={{ mt: 1 }}>
								<TextField
									label="Test Message"
									value={testMessage}
									onChange={(e) => setTestMessage(e.target.value)}
									fullWidth
									required
									multiline
									rows={3}
									placeholder="Enter a message to test the agent..."
								/>
								<TextField
									label="Context (Optional)"
									value={testContext}
									onChange={(e) => setTestContext(e.target.value)}
									fullWidth
									multiline
									rows={2}
									placeholder="Provide additional context if needed..."
								/>
								{testResult && (
									<Alert severity={testResult.success ? 'success' : 'error'}>
										{testResult.content || testResult.error}
										{testResult.execution_time_ms && (
											<Typography variant="caption" display="block" sx={{ mt: 1 }}>
												Execution time: {testResult.execution_time_ms}ms
											</Typography>
										)}
										{testResult.tokens_used && (
											<Typography variant="caption" display="block">
												Tokens used: {testResult.tokens_used.total}
											</Typography>
										)}
									</Alert>
								)}
							</Stack>
						</DialogContent>
						<DialogActions>
							<Button onClick={() => setTestDialogOpen(false)}>Close</Button>
							<Button 
								variant="contained" 
								onClick={handleRunTest}
								disabled={!testMessage || testLoading}
								startIcon={testLoading ? <CircularProgress size={20} /> : <SvgIcon>lucide:play</SvgIcon>}
							>
								Run Test
							</Button>
						</DialogActions>
					</Dialog>
				</div>
			}
		/>
	);
}

export default A2AView;