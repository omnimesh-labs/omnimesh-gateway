'use client';

import { useState, useMemo, useCallback, useEffect } from 'react';
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
import { policyApi } from '@/lib/client-api';
import type { Policy, CreatePolicyRequest, UpdatePolicyRequest } from '@/lib/types';

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


const getPolicyTypeColor = (type: string): 'primary' | 'warning' | 'error' => {
	switch (type) {
		case 'rate_limit':
			return 'warning';
		case 'access':
			return 'primary';
		case 'security':
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
		type: 'access',
		priority: 100,
		conditions: {} as Record<string, any>,
		actions: {} as Record<string, any>
	});
	const [policies, setPolicies] = useState<Policy[]>([]);
	const [isLoading, setIsLoading] = useState(false);
	const [togglingPolicyId, setTogglingPolicyId] = useState<string | null>(null);
	const { enqueueSnackbar } = useSnackbar();

	// Fetch policies on mount
	useEffect(() => {
		fetchPolicies();
	}, []);

	const fetchPolicies = async () => {
		setIsLoading(true);
		try {
			const data = await policyApi.listPolicies();
			setPolicies(data);
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : 'Failed to fetch policies';
			enqueueSnackbar(errorMessage, { variant: 'error' });
			console.error('Error fetching policies:', error);
			setPolicies([]); // Set empty array on error
		} finally {
			setIsLoading(false);
		}
	};

	const handleCreatePolicy = useCallback(() => {
		setFormData({
			name: '',
			description: '',
			type: 'access',
			priority: 100,
			conditions: {},
			actions: {}
		});
		setEditingPolicy(null);
		setCreateModalOpen(true);
	}, []);

	const handleEditPolicy = useCallback((policy: Policy) => {
		setFormData({
			name: policy.name,
			description: policy.description || '',
			type: policy.type,
			priority: policy.priority,
			conditions: policy.conditions,
			actions: policy.actions
		});
		setEditingPolicy(policy);
		setCreateModalOpen(true);
	}, []);

	const handleSavePolicy = useCallback(async () => {
		try {
			if (editingPolicy) {
				const updateData: UpdatePolicyRequest = {
					name: formData.name,
					description: formData.description,
					priority: formData.priority,
					conditions: formData.conditions,
					actions: formData.actions
				};
				await policyApi.updatePolicy(editingPolicy.id, updateData);
				enqueueSnackbar('Policy updated successfully', { variant: 'success' });
			} else {
				const createData: CreatePolicyRequest = {
					name: formData.name,
					description: formData.description,
					type: formData.type,
					priority: formData.priority,
					conditions: formData.conditions,
					actions: formData.actions
				};
				await policyApi.createPolicy(createData);
				enqueueSnackbar('Policy created successfully', { variant: 'success' });
			}

			setCreateModalOpen(false);
			setEditingPolicy(null);
			fetchPolicies(); // Refresh the list
		} catch (error) {
			const action = editingPolicy ? 'update' : 'create';
			const errorMessage = error instanceof Error ? error.message : `Failed to ${action} policy`;
			enqueueSnackbar(errorMessage, { variant: 'error' });
			console.error(`Error ${action}ing policy:`, error);
		}
	}, [editingPolicy, formData, enqueueSnackbar]);

	const handleDeletePolicy = useCallback(
		async (policy: Policy) => {
			try {
				await policyApi.deletePolicy(policy.id);
				enqueueSnackbar(`Policy ${policy.name} deleted successfully`, { variant: 'success' });
				fetchPolicies(); // Refresh the list
			} catch (error) {
				const errorMessage = error instanceof Error ? error.message : `Failed to delete policy ${policy.name}`;
				enqueueSnackbar(errorMessage, { variant: 'error' });
				console.error('Error deleting policy:', error);
			}
		},
		[enqueueSnackbar]
	);

	const handleTogglePolicy = useCallback(
		async (policy: Policy) => {
			setTogglingPolicyId(policy.id);
			try {
				const newStatus = !policy.is_active;
				await policyApi.togglePolicy(policy.id, newStatus);
				enqueueSnackbar(
					`Policy ${policy.name} ${newStatus ? 'activated' : 'deactivated'} successfully`,
					{ variant: 'success' }
				);
				fetchPolicies(); // Refresh the list
			} catch (error) {
				const errorMessage = error instanceof Error ? error.message : `Failed to toggle policy ${policy.name}`;
				enqueueSnackbar(errorMessage, { variant: 'error' });
				console.error('Error toggling policy:', error);
			} finally {
				setTogglingPolicyId(null);
			}
		},
		[enqueueSnackbar]
	);

	const columns = useMemo<MRT_ColumnDef<Policy>[]>(
		() => [
			{
				accessorKey: 'is_active',
				header: 'Active',
				size: 80,
				Cell: ({ row }) => (
					<Tooltip title={row.original.is_active ? 'Deactivate policy' : 'Activate policy'}>
						<Switch
							checked={!!row.original.is_active}
							onChange={() => handleTogglePolicy(row.original)}
							disabled={togglingPolicyId === row.original.id}
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
				accessorKey: 'priority',
				header: 'Priority',
				size: 100
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
		[handleTogglePolicy, togglingPolicyId]
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
						data={policies}
						state={{ isLoading }}
						enableRowActions
						renderRowActionMenuItems={({ row, closeMenu }) => [
							<MenuItem
								key="view"
								onClick={() => {
									// TODO: Implement view rules functionality
									closeMenu();
								}}
							>
								<SvgIcon size={16}>lucide:eye</SvgIcon>
								<span className="ml-2">View Rules</span>
							</MenuItem>,
							<MenuItem
								key="edit"
								onClick={() => {
									handleEditPolicy(row.original);
									closeMenu();
								}}
							>
								<SvgIcon size={16}>lucide:edit</SvgIcon>
								<span className="ml-2">Edit Policy</span>
							</MenuItem>,
							<MenuItem
								key="delete"
								onClick={() => {
									handleDeletePolicy(row.original);
									closeMenu();
								}}
								sx={{ color: 'error.main' }}
							>
								<SvgIcon size={16}>lucide:trash-2</SvgIcon>
								<span className="ml-2">Delete</span>
							</MenuItem>
						]}
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

								<TextField
									label="Policy Type"
									value={formData.type}
									onChange={(e) =>
										setFormData((prev) => ({
											...prev,
											type: e.target.value
										}))
									}
									select
									fullWidth
									required
									disabled={editingPolicy !== null} // Don't allow changing type when editing
								>
									<MenuItem value="access">Access Control</MenuItem>
									<MenuItem value="rate_limit">Rate Limiting</MenuItem>
									<MenuItem value="security">Security & Content Filter</MenuItem>
								</TextField>

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

								{editingPolicy && (
									<Box className="p-4 bg-gray-50 dark:bg-gray-800 rounded-md">
										<Typography variant="subtitle2" className="mb-2">
											Current Status
										</Typography>
										<FormControlLabel
											control={
												<Switch
													checked={editingPolicy.is_active}
													readOnly
													color={editingPolicy.is_active ? 'success' : 'default'}
												/>
											}
											label={`Policy is currently ${editingPolicy.is_active ? 'active' : 'inactive'}`}
											disabled
										/>
										<Typography variant="body2" color="textSecondary" className="mt-2">
											Use the toggle button in the table or dropdown menu to activate/deactivate this policy.
										</Typography>
									</Box>
								)}

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

							</Stack>
						</DialogContent>
						<DialogActions>
							<Button onClick={() => setCreateModalOpen(false)}>Cancel</Button>
							<Button
								variant="contained"
								onClick={handleSavePolicy}
								disabled={!formData.name.trim() || formData.priority < 0}
							>
								{editingPolicy ? 'Update Policy' : 'Create Policy'}
							</Button>
						</DialogActions>
					</Dialog>
				</div>
			}
		/>
	);
}

export default PoliciesView;
