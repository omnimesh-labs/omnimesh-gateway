'use client';

import { useState, useMemo } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useOptimizedQuery } from '@/hooks/useOptimizedQuery';
import { MRT_ColumnDef } from 'material-react-table';
import PageSimple from '@fuse/core/PageSimple';
import { styled } from '@mui/material/styles';
import { Typography, Button, Chip, IconButton, Tooltip, Box, Tabs, Tab, Switch } from '@mui/material';
import LazyDataTable from '@/components/data-table/LazyDataTable';
import SvgIcon from '@fuse/core/SvgIcon';
import { useSnackbar } from 'notistack';
import { MCPServer, CreateServerRequest, serverApi, discoveryApi, MCPDiscoveryResponse } from '@/lib/api';
import RegisterServerModal from './components/RegisterServerModal';
import AvailableServersTable from './components/AvailableServersTable';
import ServerDetailsModal from './components/ServerDetailsModal';
import UnregisterConfirmDialog from './components/UnregisterConfirmDialog';

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

const getStatusColor = (status: string): 'success' | 'warning' | 'error' | 'default' => {
	switch (status) {
		case 'active':
			return 'success';
		case 'maintenance':
			return 'warning';
		case 'unhealthy':
			return 'error';
		default:
			return 'default';
	}
};

function ServersView() {
	const [tabValue, setTabValue] = useState(0);
	const [registerModalOpen, setRegisterModalOpen] = useState(false);
	const [selectedServer, setSelectedServer] = useState<MCPServer | null>(null);
	const [detailsModalOpen, setDetailsModalOpen] = useState(false);
	const [serverToUnregister, setServerToUnregister] = useState<MCPServer | null>(null);
	const [unregisterDialogOpen, setUnregisterDialogOpen] = useState(false);
	const { enqueueSnackbar } = useSnackbar();
	const queryClient = useQueryClient();

	// Fetch registered servers with optimized caching
	const { data: servers = [], isLoading, refetch } = useOptimizedQuery<MCPServer[]>(
		['servers'],
		() => serverApi.listServers(),
		{
			refetchInterval: false, // Disable auto-refetch to prevent conflicts with manual updates
			cacheKey: 'servers-list',
			staleTime: 5 * 60 * 1000, // Consider data stale after 5 minutes
		}
	);

	// Fetch available servers for discovery with optimized caching
	const { data: discoveryData, isLoading: discoveryLoading } = useOptimizedQuery<MCPDiscoveryResponse>(
		['discovery', 'packages'],
		() => discoveryApi.listPackages(0, 50),
		{
			enabled: tabValue === 1,
			cacheKey: 'discovery-packages'
		}
	);

	// Unregister server mutation
	const unregisterMutation = useMutation({
		mutationFn: (serverId: string) => serverApi.unregisterServer(serverId),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['servers'] });
			enqueueSnackbar('Server unregistered successfully', { variant: 'success' });
		},
		onError: (error: Error | unknown) => {
			const message = error instanceof Error ? error.message : 'Failed to unregister server';
			enqueueSnackbar(message, { variant: 'error' });
		}
	});

	// Register server mutation
	const registerMutation = useMutation({
		mutationFn: (serverData: CreateServerRequest) => serverApi.registerServer(serverData),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['servers'] });
			enqueueSnackbar('Server registered successfully', { variant: 'success' });
			setRegisterModalOpen(false);
		},
		onError: (error: Error | unknown) => {
			const message = error instanceof Error ? error.message : 'Failed to register server';
			enqueueSnackbar(message, { variant: 'error' });
		}
	});

	// Toggle server active status mutation
	const toggleServerMutation = useMutation({
		mutationFn: ({ id, isActive }: { id: string; isActive: boolean }) =>
			serverApi.updateServer(id, { is_active: isActive }),
		onSuccess: (updatedServer, variables) => {
			// Update the cache with the actual server data returned from API
			queryClient.setQueryData<MCPServer[]>(['servers'], (oldServers) => {
				if (!oldServers) return oldServers;
				return oldServers.map(server =>
					server.id === variables.id ? updatedServer : server
				);
			});
			// Show specific toast based on activation/deactivation
			if (variables.isActive) {
				enqueueSnackbar('Server activated successfully', { variant: 'success' });
			} else {
				enqueueSnackbar('Server deactivated successfully', { variant: 'success' });
			}
		},
		onError: (error: Error | unknown) => {
			const message = error instanceof Error ? error.message : 'Failed to update server status';
			enqueueSnackbar(message, { variant: 'error' });
			// Refetch to ensure UI is in sync after error
			refetch();
		}
	});

	const handleUnregisterServer = (server: MCPServer) => {
		setServerToUnregister(server);
		setUnregisterDialogOpen(true);
	};

	const handleConfirmUnregister = (serverId: string) => {
		unregisterMutation.mutate(serverId);
		setUnregisterDialogOpen(false);
		setServerToUnregister(null);
	};

	const handleCloseUnregisterDialog = () => {
		setUnregisterDialogOpen(false);
		setServerToUnregister(null);
	};

	const handleRegisterServer = (serverData: CreateServerRequest) => {
		registerMutation.mutate(serverData);
	};

	const handleViewDetails = (server: MCPServer) => {
		setSelectedServer(server);
		setDetailsModalOpen(true);
	};

	const handleCloseDetails = () => {
		setDetailsModalOpen(false);
		setSelectedServer(null);
	};

	const handleToggleServerStatus = (server: MCPServer) => {
		toggleServerMutation.mutate({ id: server.id, isActive: !server.is_active });
	};

	const columns = useMemo<MRT_ColumnDef<MCPServer>[]>(
		() => [
			{
				accessorKey: 'is_active',
				header: 'Active',
				size: 80,
				Cell: ({ row }) => (
					<Tooltip title={row.original.is_active ? 'Deactivate server' : 'Activate server'}>
						<Switch
							checked={!!row.original.is_active}
							onChange={() => handleToggleServerStatus(row.original)}
							disabled={toggleServerMutation.isPending}
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
				maxSize: 250,
				Cell: ({ row }) => (
					<Box className="flex items-center space-x-2" sx={{ maxWidth: 250 }}>
						<SvgIcon size={20}>lucide:server</SvgIcon>
						<Box sx={{ minWidth: 0, flex: 1 }}>
							<Typography
								variant="body2"
								className="font-medium"
								sx={{
									overflow: 'hidden',
									textOverflow: 'ellipsis',
									whiteSpace: 'nowrap'
								}}
							>
								{row.original.name}
							</Typography>
							{row.original.description && (
								<Typography
									variant="caption"
									color="textSecondary"
									sx={{
										overflow: 'hidden',
										textOverflow: 'ellipsis',
										whiteSpace: 'nowrap',
										display: 'block'
									}}
								>
									{row.original.description}
								</Typography>
							)}
						</Box>
					</Box>
				)
			},
			{
				accessorKey: 'protocol',
				header: 'Protocol',
				size: 100,
				Cell: ({ cell }) => (
					<Chip
						size="small"
						label={cell.getValue<string>().toUpperCase()}
						variant="outlined"
					/>
				)
			},
			{
				accessorKey: 'version',
				header: 'Version',
				size: 100
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

	const availableServers = discoveryData ? Object.values(discoveryData.results) : [];

	return (
		<Root
			header={
				<div className="p-6">
					<div className="flex items-center justify-between">
						<div>
							<Typography variant="h4">Server Management</Typography>
							<Typography
								variant="body1"
								color="textSecondary"
								className="mt-1"
							>
								Manage your MCP servers and discover new ones
							</Typography>
						</div>
						<Button
							variant="contained"
							color="primary"
							startIcon={<SvgIcon>lucide:plus</SvgIcon>}
							onClick={() => setRegisterModalOpen(true)}
							disabled={registerMutation.isPending}
						>
							Register Server
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
								label={`Registered Servers (${servers.length})`}
								icon={<SvgIcon size={20}>lucide:server</SvgIcon>}
								iconPosition="start"
							/>
							<Tab
								label="Catalog"
								icon={<SvgIcon size={20}>lucide:search</SvgIcon>}
								iconPosition="start"
							/>
						</Tabs>
					</Box>

					{tabValue === 0 && (
						<LazyDataTable
							columns={columns}
							data={servers}
							state={{
								isLoading
							}}
							enableRowActions
							renderRowActions={({ row }) => (
								<Box className="flex items-center space-x-1">
									<Tooltip title="View Details">
										<IconButton
											size="small"
											onClick={() => handleViewDetails(row.original)}
										>
											<SvgIcon size={18}>lucide:eye</SvgIcon>
										</IconButton>
									</Tooltip>
									<Tooltip title="Unregister Server">
										<IconButton
											size="small"
											onClick={() => handleUnregisterServer(row.original)}
											disabled={unregisterMutation.isPending}
											color="error"
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

					{tabValue === 1 && (
						<AvailableServersTable
							servers={availableServers}
							loading={discoveryLoading}
							onRegisterServer={handleRegisterServer}
							registering={registerMutation.isPending}
							registeredServers={servers}
						/>
					)}

					<RegisterServerModal
						open={registerModalOpen}
						onClose={() => setRegisterModalOpen(false)}
						onRegister={handleRegisterServer}
						loading={registerMutation.isPending}
					/>

					<ServerDetailsModal
						server={selectedServer}
						open={detailsModalOpen}
						onClose={handleCloseDetails}
					/>

					<UnregisterConfirmDialog
						server={serverToUnregister}
						open={unregisterDialogOpen}
						onClose={handleCloseUnregisterDialog}
						onConfirm={handleConfirmUnregister}
						loading={unregisterMutation.isPending}
					/>
				</div>
			}
		/>
	);
}

export default ServersView;
