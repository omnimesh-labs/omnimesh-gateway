'use client';

import { useState, useMemo, useEffect } from 'react';
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
	Alert,
	Paper
} from '@mui/material';
import DataTable from '@/components/data-table/DataTable';
import SvgIcon from '@fuse/core/SvgIcon';
import { useSnackbar } from 'notistack';
import { authApi } from '@/lib/auth-api';
import type { ApiKey as ApiKeyType, CreateApiKeyRequest } from '@/lib/types';

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

type ApiKey = ApiKeyType;

const getRoleColor = (role: string): 'error' | 'primary' | 'default' => {
	switch (role) {
		case 'admin':
			return 'error';
		case 'user':
			return 'primary';
		default:
			return 'default';
	}
};

function ApiKeysView() {
	const [createModalOpen, setCreateModalOpen] = useState(false);
	const [createdKey, setCreatedKey] = useState<string | null>(null);
	const [keyCreatedModalOpen, setKeyCreatedModalOpen] = useState(false);
	const [formData, setFormData] = useState({
		name: '',
		role: 'user' as 'admin' | 'user' | 'viewer',
		expires_at: ''
	});
	const [isLoading, setIsLoading] = useState(false);
	const { enqueueSnackbar } = useSnackbar();
	const [apiKeys, setApiKeys] = useState<ApiKey[]>([]);
	const [loading, setLoading] = useState(true);

	// Fetch API keys on mount
	useEffect(() => {
		fetchApiKeys();
	}, []);

	const fetchApiKeys = async () => {
		try {
			setLoading(true);
			const keys = await authApi.getApiKeys();
			setApiKeys(keys);
		} catch (error) {
			console.error('Failed to fetch API keys:', error);
			enqueueSnackbar('Failed to load API keys', { variant: 'error' });
		} finally {
			setLoading(false);
		}
	};

	const handleCreateApiKey = async (e: React.FormEvent) => {
		e.preventDefault();
		setIsLoading(true);

		try {
			const request: CreateApiKeyRequest = {
				name: formData.name,
				role: formData.role as 'admin' | 'user' | 'viewer' | 'api_user',
				expires_at: formData.expires_at || undefined
			};

			const response = await authApi.createApiKey(request);
			setCreatedKey(response.data?.key || response.key); // The actual key is only returned on creation
			setCreateModalOpen(false);
			setKeyCreatedModalOpen(true);

			// Refresh the list
			fetchApiKeys();

			setFormData({ name: '', role: 'user', expires_at: '' });
			enqueueSnackbar('API key created successfully', { variant: 'success' });
		} catch (error) {
			console.error('Failed to create API key:', error);
			enqueueSnackbar('Failed to create API key', { variant: 'error' });
		} finally {
			setIsLoading(false);
		}
	};

	const handleDeleteApiKey = async (apiKey: ApiKey) => {
		try {
			await authApi.deleteApiKey(apiKey.id);
			enqueueSnackbar(`API key "${apiKey.name}" deleted successfully`, { variant: 'success' });
			fetchApiKeys();
		} catch (error) {
			console.error('Failed to delete API key:', error);
			enqueueSnackbar('Failed to delete API key', { variant: 'error' });
		}
	};

	const handleCopyKey = () => {
		if (createdKey) {
			navigator.clipboard.writeText(createdKey);
			enqueueSnackbar('API key copied to clipboard', { variant: 'success' });
		}
	};

	const columns = useMemo<MRT_ColumnDef<ApiKey>[]>(
		() => [
			{
				accessorKey: 'name',
				header: 'Name',
				size: 200,
				Cell: ({ row }) => (
					<Box className="flex items-center space-x-2">
						<SvgIcon size={20}>lucide:key</SvgIcon>
						<Box>
							<Typography
								variant="body2"
								className="font-medium"
							>
								{row.original.name}
							</Typography>
							{row.original.key_hash && (
								<Typography
									variant="caption"
									color="textSecondary"
									className="font-mono"
								>
									{row.original.key_hash.substring(0, 8)}...
								</Typography>
							)}
						</Box>
					</Box>
				)
			},
			{
				accessorKey: 'role',
				header: 'Role',
				size: 100,
				Cell: ({ cell }) => (
					<Chip
						size="small"
						label={cell.getValue<string>()}
						color={getRoleColor(cell.getValue<string>())}
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
				accessorKey: 'expires_at',
				header: 'Expires',
				size: 150,
				Cell: ({ cell }) => {
					const expires = cell.getValue<string>();

					if (!expires) return <Typography variant="body2">Never</Typography>;

					const date = new Date(expires);
					const isExpired = date < new Date();

					return (
						<Typography
							variant="body2"
							color={isExpired ? 'error.main' : 'textPrimary'}
						>
							{date.toLocaleDateString()}
						</Typography>
					);
				}
			},
			{
				accessorKey: 'last_used_at',
				header: 'Last Used',
				size: 150,
				Cell: ({ cell }) => {
					const lastUsed = cell.getValue<string>();

					if (!lastUsed)
						return (
							<Typography
								variant="body2"
								color="textSecondary"
							>
								Never
							</Typography>
						);

					return <Typography variant="body2">{new Date(lastUsed).toLocaleDateString()}</Typography>;
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
		[]
	);

	return (
		<Root
			header={
				<div className="p-6">
					<div className="flex items-center justify-between">
						<div>
							<Typography variant="h4">API Keys</Typography>
							<Typography
								variant="body1"
								color="textSecondary"
								className="mt-1"
							>
								Manage your API keys for programmatic access
							</Typography>
						</div>
						<Button
							variant="contained"
							color="primary"
							startIcon={<SvgIcon>lucide:plus</SvgIcon>}
							onClick={() => setCreateModalOpen(true)}
						>
							Create API Key
						</Button>
					</div>
				</div>
			}
			content={
				<div className="p-6">
					<Box className="mb-4">
						<Alert severity="warning">
							API keys provide full access to your account. Keep them secure and never share them
							publicly.
						</Alert>
					</Box>

					<DataTable
						columns={columns}
						data={apiKeys || []}
						state={{ isLoading: loading }}
						enableRowActions
						renderRowActions={({ row }) => (
							<Box className="flex items-center space-x-1">
								<Tooltip title="Delete API Key">
									<IconButton
										size="small"
										color="error"
										onClick={() => handleDeleteApiKey(row.original)}
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

					{/* Create API Key Dialog */}
					<Dialog
						open={createModalOpen}
						onClose={() => setCreateModalOpen(false)}
						maxWidth="sm"
						fullWidth
					>
						<form onSubmit={handleCreateApiKey}>
							<DialogTitle>Create New API Key</DialogTitle>
							<DialogContent>
								<Stack
									spacing={3}
									sx={{ mt: 1 }}
								>
									<TextField
										label="Key Name"
										value={formData.name}
										onChange={(e) => setFormData((prev) => ({ ...prev, name: e.target.value }))}
										fullWidth
										required
										placeholder="e.g., Production API, Mobile App"
									/>

									<TextField
										label="Role"
										value={formData.role}
										onChange={(e) =>
											setFormData((prev) => ({
												...prev,
												role: e.target.value as 'admin' | 'user' | 'viewer' | 'api_user'
											}))
										}
										select
										fullWidth
										required
									>
										<MenuItem value="viewer">Viewer (Read-only)</MenuItem>
										<MenuItem value="user">User (Read/Write)</MenuItem>
										<MenuItem value="admin">Admin (Full access)</MenuItem>
									</TextField>

									<TextField
										label="Expiration Date (Optional)"
										type="datetime-local"
										value={formData.expires_at}
										onChange={(e) =>
											setFormData((prev) => ({ ...prev, expires_at: e.target.value }))
										}
										fullWidth
										InputLabelProps={{ shrink: true }}
									/>
								</Stack>
							</DialogContent>
							<DialogActions>
								<Button onClick={() => setCreateModalOpen(false)}>Cancel</Button>
								<Button
									type="submit"
									variant="contained"
									disabled={isLoading || !formData.name.trim()}
								>
									{isLoading ? 'Creating...' : 'Create API Key'}
								</Button>
							</DialogActions>
						</form>
					</Dialog>

					{/* API Key Created Dialog */}
					<Dialog
						open={keyCreatedModalOpen}
						onClose={() => setKeyCreatedModalOpen(false)}
						maxWidth="md"
						fullWidth
					>
						<DialogTitle className="flex items-center gap-2">
							<SvgIcon color="warning">lucide:alert-triangle</SvgIcon>
							API Key Created - Save This Key!
						</DialogTitle>
						<DialogContent>
							<Stack spacing={3}>
								<Alert severity="warning">
									<Typography
										variant="body2"
										gutterBottom
									>
										<strong>This key will only be shown once.</strong> Please copy and save it
										securely.
									</Typography>
									<Typography variant="body2">
										If you lose this key, you'll need to create a new one.
									</Typography>
								</Alert>

								<Box>
									<Typography
										variant="subtitle2"
										gutterBottom
									>
										Your new API key:
									</Typography>
									<Paper
										variant="outlined"
										className="border-yellow-200 bg-yellow-50 p-3"
									>
										<Box className="flex items-center justify-between">
											<Typography
												variant="body2"
												className="break-all font-mono text-yellow-800"
											>
												{createdKey}
											</Typography>
											<IconButton
												size="small"
												onClick={handleCopyKey}
												className="ml-2"
											>
												<SvgIcon size={16}>lucide:copy</SvgIcon>
											</IconButton>
										</Box>
									</Paper>
								</Box>
							</Stack>
						</DialogContent>
						<DialogActions>
							<Button
								onClick={() => setKeyCreatedModalOpen(false)}
								variant="contained"
							>
								I've Saved the Key
							</Button>
						</DialogActions>
					</Dialog>
				</div>
			}
		/>
	);
}

export default ApiKeysView;
