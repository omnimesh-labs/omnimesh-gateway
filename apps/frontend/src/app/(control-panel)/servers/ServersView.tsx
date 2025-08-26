'use client';

import { useState, useMemo, useEffect } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useOptimizedQuery } from '@/hooks/useOptimizedQuery';
import { measurePageLoad } from '@/lib/performance';
import { MRT_ColumnDef } from 'material-react-table';
import PageSimple from '@fuse/core/PageSimple';
import { styled } from '@mui/material/styles';
import { Typography, Button, Chip, IconButton, Tooltip, Box, Tabs, Tab } from '@mui/material';
import LazyDataTable from '@/components/data-table/LazyDataTable';
import SvgIcon from '@fuse/core/SvgIcon';
import { useSnackbar } from 'notistack';
import { MCPServer, serverApi, discoveryApi, MCPDiscoveryResponse } from '@/lib/api';
import RegisterServerModal from './components/RegisterServerModal';
import AvailableServersTable from './components/AvailableServersTable';

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
	const { enqueueSnackbar } = useSnackbar();
	const queryClient = useQueryClient();

	// Track page performance
	useEffect(() => {
		measurePageLoad('ServersView');
	}, []);

	// Fetch registered servers with optimized caching
	const { data: servers = [], isLoading } = useOptimizedQuery<MCPServer[]>(
		['servers'],
		() => serverApi.listServers(),
		{
			refetchInterval: 15000,
			cacheKey: 'servers-list'
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
		onError: (error: any) => {
			enqueueSnackbar(error.message || 'Failed to unregister server', { variant: 'error' });
		}
	});

	// Register server mutation
	const registerMutation = useMutation({
		mutationFn: (serverData: any) => serverApi.registerServer(serverData),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['servers'] });
			enqueueSnackbar('Server registered successfully', { variant: 'success' });
			setRegisterModalOpen(false);
		},
		onError: (error: any) => {
			enqueueSnackbar(error.message || 'Failed to register server', { variant: 'error' });
		}
	});

	const handleUnregisterServer = async (serverId: string, serverName: string) => {
		if (window.confirm(`Are you sure you want to unregister "${serverName}"?`)) {
			unregisterMutation.mutate(serverId);
		}
	};

	const handleRegisterServer = (serverData: any) => {
		registerMutation.mutate(serverData);
	};

	const columns = useMemo<MRT_ColumnDef<MCPServer>[]>(
		() => [
			{
				accessorKey: 'name',
				header: 'Name',
				size: 200,
				Cell: ({ row }) => (
					<Box className="flex items-center space-x-2">
						<SvgIcon size={20}>lucide:server</SvgIcon>
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
				accessorKey: 'status',
				header: 'Status',
				size: 120,
				Cell: ({ cell }) => (
					<Chip
						size="small"
						label={cell.getValue<string>()}
						color={getStatusColor(cell.getValue<string>())}
						sx={{ textTransform: 'capitalize' }}
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
										<IconButton size="small">
											<SvgIcon size={18}>lucide:eye</SvgIcon>
										</IconButton>
									</Tooltip>
									<Tooltip title="Remove Server">
										<IconButton
											size="small"
											onClick={() => handleUnregisterServer(row.original.id, row.original.name)}
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
						/>
					)}

					<RegisterServerModal
						open={registerModalOpen}
						onClose={() => setRegisterModalOpen(false)}
						onRegister={handleRegisterServer}
						loading={registerMutation.isPending}
					/>
				</div>
			}
		/>
	);
}

export default ServersView;
