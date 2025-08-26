'use client';

import { useState, useMemo, useCallback } from 'react';
import { MRT_ColumnDef, MRT_Row } from 'material-react-table';
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
	Accordion,
	AccordionSummary,
	AccordionDetails
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

interface Policy {
	id: string;
	name: string;
	description?: string;
	type: 'rate_limit' | 'access_control' | 'content_filter';
	scope: 'global' | 'namespace' | 'user';
	is_active: boolean;
	priority: number;
	created_at: string;
	updated_at: string;
	rules: any;
}

const getPolicyTypeColor = (type: string): 'primary' | 'warning' | 'error' => {
	switch (type) {
		case 'rate_limit':
			return 'warning';
		case 'access_control':
			return 'primary';
		case 'content_filter':
			return 'error';
		default:
			return 'primary';
	}
};

function PoliciesView() {
	const [createModalOpen, setCreateModalOpen] = useState(false);
	const [editingPolicy, setEditingPolicy] = useState<Policy | null>(null);
	const [formData, setFormData] = useState({
		name: '',
		description: '',
		type: 'rate_limit' as 'rate_limit' | 'access_control' | 'content_filter',
		scope: 'namespace' as 'global' | 'namespace' | 'user',
		is_active: true,
		priority: 100
	});
	const { enqueueSnackbar } = useSnackbar();

	// Mock data - memoized to prevent recreation on each render
	const mockPolicies = useMemo<Policy[]>(
		() => [
			{
				id: '1',
				name: 'Default Rate Limit',
				description: 'Standard rate limiting for all users',
				type: 'rate_limit',
				scope: 'global',
				is_active: true,
				priority: 100,
				created_at: '2024-01-15T10:00:00Z',
				updated_at: '2024-01-20T15:30:00Z',
				rules: {
					requests_per_minute: 60,
					burst_limit: 10
				}
			},
			{
				id: '2',
				name: 'Admin Access Control',
				description: 'Full access for administrators',
				type: 'access_control',
				scope: 'user',
				is_active: true,
				priority: 200,
				created_at: '2024-01-10T09:00:00Z',
				updated_at: '2024-01-18T14:00:00Z',
				rules: {
					allowed_actions: ['*'],
					denied_actions: []
				}
			},
			{
				id: '3',
				name: 'Production Content Filter',
				description: 'Content filtering for production namespace',
				type: 'content_filter',
				scope: 'namespace',
				is_active: false,
				priority: 50,
				created_at: '2024-01-12T11:00:00Z',
				updated_at: '2024-01-22T16:00:00Z',
				rules: {
					blocked_keywords: ['test', 'debug'],
					max_content_length: 10000
				}
			}
		],
		[]
	);

	const handleCreatePolicy = useCallback(() => {
		setFormData({
			name: '',
			description: '',
			type: 'rate_limit',
			scope: 'namespace',
			is_active: true,
			priority: 100
		});
		setEditingPolicy(null);
		setCreateModalOpen(true);
	}, []);

	const handleEditPolicy = useCallback((policy: Policy) => {
		setFormData({
			name: policy.name,
			description: policy.description || '',
			type: policy.type,
			scope: policy.scope,
			is_active: policy.is_active,
			priority: policy.priority
		});
		setEditingPolicy(policy);
		setCreateModalOpen(true);
	}, []);

	const handleSavePolicy = useCallback(() => {
		const action = editingPolicy ? 'updated' : 'created';
		enqueueSnackbar(`Policy ${action} successfully (demo)`, { variant: 'success' });
		setCreateModalOpen(false);
		setEditingPolicy(null);
	}, [editingPolicy, enqueueSnackbar]);

	const handleDeletePolicy = useCallback(
		(policy: Policy) => {
			enqueueSnackbar(`Delete functionality coming soon for ${policy.name}`, { variant: 'info' });
		},
		[enqueueSnackbar]
	);

	const columns = useMemo<MRT_ColumnDef<Policy>[]>(
		() => [
			{
				accessorKey: 'name',
				header: 'Name',
				size: 200,
				Cell: ({ row }) => (
					<Box className="flex items-center space-x-2">
						<SvgIcon size={20}>lucide:shield</SvgIcon>
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
				accessorKey: 'type',
				header: 'Type',
				size: 150,
				Cell: ({ cell }) => (
					<Chip
						size="small"
						label={cell.getValue<string>().replace('_', ' ')}
						color={getPolicyTypeColor(cell.getValue<string>())}
						sx={{ textTransform: 'capitalize' }}
					/>
				)
			},
			{
				accessorKey: 'scope',
				header: 'Scope',
				size: 100,
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
				accessorKey: 'priority',
				header: 'Priority',
				size: 100
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
				accessorKey: 'updated_at',
				header: 'Last Modified',
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
							<Typography variant="h4">Policy Management</Typography>
							<Typography
								variant="body1"
								color="textSecondary"
								className="mt-1"
							>
								Define and manage access control, rate limiting, and content filtering policies
							</Typography>
						</div>
						<Button
							variant="contained"
							color="primary"
							startIcon={<SvgIcon>lucide:plus</SvgIcon>}
							onClick={handleCreatePolicy}
						>
							Create Policy
						</Button>
					</div>
				</div>
			}
			content={
				<div className="p-6">
					<DataTable
						columns={columns}
						data={mockPolicies}
						enableRowActions
						renderRowActions={useCallback(
							({ row }: { row: MRT_Row<Policy> }) => (
								<Box className="flex items-center space-x-1">
									<Tooltip title="View Rules">
										<IconButton size="small">
											<SvgIcon size={18}>lucide:eye</SvgIcon>
										</IconButton>
									</Tooltip>
									<Tooltip title="Edit Policy">
										<IconButton
											size="small"
											onClick={() => handleEditPolicy(row.original)}
										>
											<SvgIcon size={18}>lucide:edit</SvgIcon>
										</IconButton>
									</Tooltip>
									<Tooltip title="Delete Policy">
										<IconButton
											size="small"
											color="error"
											onClick={() => handleDeletePolicy(row.original)}
										>
											<SvgIcon size={18}>lucide:trash-2</SvgIcon>
										</IconButton>
									</Tooltip>
								</Box>
							),
							[handleEditPolicy, handleDeletePolicy]
						)}
					/>

					{/* Create/Edit Policy Dialog */}
					<Dialog
						open={createModalOpen}
						onClose={() => setCreateModalOpen(false)}
						maxWidth="md"
						fullWidth
					>
						<DialogTitle>{editingPolicy ? 'Edit Policy' : 'Create Policy'}</DialogTitle>
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
									label="Description"
									value={formData.description}
									onChange={(e) => setFormData((prev) => ({ ...prev, description: e.target.value }))}
									fullWidth
									multiline
									rows={2}
								/>

								<Box className="grid grid-cols-2 gap-3">
									<TextField
										label="Policy Type"
										value={formData.type}
										onChange={(e) =>
											setFormData((prev) => ({ ...prev, type: e.target.value as any }))
										}
										select
										fullWidth
										required
									>
										<MenuItem value="rate_limit">Rate Limit</MenuItem>
										<MenuItem value="access_control">Access Control</MenuItem>
										<MenuItem value="content_filter">Content Filter</MenuItem>
									</TextField>

									<TextField
										label="Scope"
										value={formData.scope}
										onChange={(e) =>
											setFormData((prev) => ({ ...prev, scope: e.target.value as any }))
										}
										select
										fullWidth
										required
									>
										<MenuItem value="global">Global</MenuItem>
										<MenuItem value="namespace">Namespace</MenuItem>
										<MenuItem value="user">User</MenuItem>
									</TextField>
								</Box>

								<TextField
									label="Priority"
									type="number"
									value={formData.priority}
									onChange={(e) =>
										setFormData((prev) => ({ ...prev, priority: parseInt(e.target.value) || 0 }))
									}
									fullWidth
									helperText="Higher numbers = higher priority"
								/>

								<Accordion>
									<AccordionSummary expandIcon={<SvgIcon>lucide:chevron-down</SvgIcon>}>
										<Typography>Policy Rules</Typography>
									</AccordionSummary>
									<AccordionDetails>
										<Typography
											variant="body2"
											color="textSecondary"
										>
											Policy rules configuration will be available based on the selected policy
											type. This is a placeholder for the rule configuration interface.
										</Typography>
									</AccordionDetails>
								</Accordion>

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
								onClick={handleSavePolicy}
								disabled={!formData.name.trim()}
							>
								{editingPolicy ? 'Update' : 'Create'}
							</Button>
						</DialogActions>
					</Dialog>
				</div>
			}
		/>
	);
}

export default PoliciesView;
