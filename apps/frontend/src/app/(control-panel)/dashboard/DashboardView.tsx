'use client';

import { useEffect, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import PageSimple from '@fuse/core/PageSimple';
import { styled } from '@mui/material/styles';
import { Card, CardContent, Grid, Typography, Box, Chip, LinearProgress } from '@mui/material';
import SvgIcon from '@fuse/core/SvgIcon';
import { authApi, serverApi, adminApi, SystemStats, MCPServer } from '@/lib/api';

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

const StatCard = ({ title, value, subtitle, icon, color = 'primary' }: {
	title: string;
	value: string | number;
	subtitle?: string;
	icon: string;
	color?: 'primary' | 'secondary' | 'success' | 'warning' | 'error';
}) => (
	<Card className="relative overflow-hidden">
		<CardContent>
			<Box className="flex items-center justify-between">
				<Box>
					<Typography color="textSecondary" gutterBottom variant="overline">
						{title}
					</Typography>
					<Typography variant="h4" component="h2">
						{value}
					</Typography>
					{subtitle && (
						<Typography color="textSecondary" variant="body2">
							{subtitle}
						</Typography>
					)}
				</Box>
				<Box>
					<SvgIcon className={`text-${color}`} size={48}>
						{icon}
					</SvgIcon>
				</Box>
			</Box>
		</CardContent>
	</Card>
);

const ServerStatusCard = ({ server }: { server: MCPServer }) => (
	<Card className="mb-3">
		<CardContent className="py-3">
			<Box className="flex items-center justify-between">
				<Box className="flex items-center space-x-3">
					<SvgIcon size={24}>lucide:server</SvgIcon>
					<Box>
						<Typography variant="subtitle2">{server.name}</Typography>
						<Typography variant="caption" color="textSecondary">
							{server.protocol.toUpperCase()}
						</Typography>
					</Box>
				</Box>
				<Chip
					size="small"
					label={server.status}
					color={
						server.status === 'active' ? 'success' :
						server.status === 'unhealthy' ? 'error' :
						server.status === 'maintenance' ? 'warning' : 'default'
					}
				/>
			</Box>
		</CardContent>
	</Card>
);

function DashboardView() {
	const [user, setUser] = useState(null);

	// Fetch user profile
	useEffect(() => {
		const fetchUser = async () => {
			try {
				const userProfile = await authApi.getProfile();
				setUser(userProfile);
			} catch (error) {
				console.error('Failed to fetch user profile:', error);
			}
		};

		fetchUser();
	}, []);

	// Fetch system stats
	const { data: stats, isLoading: statsLoading } = useQuery<SystemStats>({
		queryKey: ['admin', 'stats'],
		queryFn: () => adminApi.getStats(),
		refetchInterval: 30000, // Refresh every 30 seconds
	});

	// Fetch servers
	const { data: servers, isLoading: serversLoading } = useQuery<MCPServer[]>({
		queryKey: ['servers'],
		queryFn: () => serverApi.listServers(),
		refetchInterval: 15000, // Refresh every 15 seconds
	});

	if (statsLoading) {
		return (
			<Root
				header={
					<div className="p-6">
						<Typography variant="h4">Dashboard</Typography>
						<LinearProgress className="mt-2" />
					</div>
				}
				content={
					<div className="p-6">
						<LinearProgress />
					</div>
				}
			/>
		);
	}

	const healthyServers = servers?.filter(s => s.status === 'active').length || 0;
	const totalServers = servers?.length || 0;
	const serverHealthRate = totalServers > 0 ? (healthyServers / totalServers) * 100 : 0;

	return (
		<Root
			header={
				<div className="p-6">
					<Typography variant="h4">Dashboard</Typography>
					<Typography variant="body1" color="textSecondary" className="mt-1">
						Welcome back{user ? `, ${user.email}` : ''}! Here's your MCP Gateway overview.
					</Typography>
				</div>
			}
			content={
				<div className="p-6">
					<Grid container spacing={3}>
						{/* Stats Cards */}
						<Grid size={{ xs: 12, sm: 6, md: 3 }}>
							<StatCard
								title="Total Servers"
								value={stats?.servers?.total || 0}
								subtitle={`${healthyServers} healthy`}
								icon="lucide:server"
								color="primary"
							/>
						</Grid>
						<Grid size={{ xs: 12, sm: 6, md: 3 }}>
							<StatCard
								title="Total Users"
								value={stats?.users?.total || 0}
								subtitle={`${stats?.users?.active || 0} active`}
								icon="lucide:users"
								color="secondary"
							/>
						</Grid>
						<Grid size={{ xs: 12, sm: 6, md: 3 }}>
							<StatCard
								title="Total Requests"
								value={stats?.requests?.total || 0}
								subtitle={`${stats?.requests?.successful || 0} successful`}
								icon="lucide:activity"
								color="success"
							/>
						</Grid>
						<Grid size={{ xs: 12, sm: 6, md: 3 }}>
							<StatCard
								title="Rate Limits"
								value={stats?.rate_limits?.blocked || 0}
								subtitle="blocked requests"
								icon="lucide:shield"
								color="warning"
							/>
						</Grid>

						{/* Server Health Overview */}
						<Grid size={{ xs: 12, md: 8 }}>
							<Card>
								<CardContent>
									<Typography variant="h6" gutterBottom>
										Server Health Overview
									</Typography>
									<Box className="mb-4">
										<Box className="flex justify-between items-center mb-2">
											<Typography variant="body2" color="textSecondary">
												Overall Health: {serverHealthRate.toFixed(1)}%
											</Typography>
											<Typography variant="body2" color="textSecondary">
												{healthyServers}/{totalServers} servers healthy
											</Typography>
										</Box>
										<LinearProgress
											variant="determinate"
											value={serverHealthRate}
											color={serverHealthRate > 80 ? 'success' : serverHealthRate > 60 ? 'warning' : 'error'}
											className="h-2 rounded"
										/>
									</Box>
									<Box className="max-h-64 overflow-y-auto">
										{serversLoading ? (
											<LinearProgress />
										) : servers && servers.length > 0 ? (
											servers.slice(0, 5).map((server) => (
												<ServerStatusCard key={server.id} server={server} />
											))
										) : (
											<Box className="text-center py-8">
												<SvgIcon size={64} className="text-gray-300 mb-4">lucide:server</SvgIcon>
												<Typography variant="body2" color="textSecondary">
													No servers registered yet
												</Typography>
											</Box>
										)}
									</Box>
								</CardContent>
							</Card>
						</Grid>

						{/* Quick Actions */}
						<Grid size={{ xs: 12, md: 4 }}>
							<Card>
								<CardContent>
									<Typography variant="h6" gutterBottom>
										Quick Actions
									</Typography>
									<Box className="space-y-3">
										<Card
											className="cursor-pointer hover:bg-gray-50 border border-gray-200"
											onClick={() => window.location.href = '/servers'}
										>
											<CardContent className="py-3">
												<Box className="flex items-center space-x-3">
													<SvgIcon className="text-primary">lucide:plus</SvgIcon>
													<Typography variant="body2">Register New Server</Typography>
												</Box>
											</CardContent>
										</Card>

										<Card
											className="cursor-pointer hover:bg-gray-50 border border-gray-200"
											onClick={() => window.location.href = '/namespaces'}
										>
											<CardContent className="py-3">
												<Box className="flex items-center space-x-3">
													<SvgIcon className="text-secondary">lucide:folder-plus</SvgIcon>
													<Typography variant="body2">Create Namespace</Typography>
												</Box>
											</CardContent>
										</Card>

										<Card
											className="cursor-pointer hover:bg-gray-50 border border-gray-200"
											onClick={() => window.location.href = '/logs'}
										>
											<CardContent className="py-3">
												<Box className="flex items-center space-x-3">
													<SvgIcon className="text-info">lucide:eye</SvgIcon>
													<Typography variant="body2">View System Logs</Typography>
												</Box>
											</CardContent>
										</Card>

										<Card
											className="cursor-pointer hover:bg-gray-50 border border-gray-200"
											onClick={() => window.location.href = '/configuration'}
										>
											<CardContent className="py-3">
												<Box className="flex items-center space-x-3">
													<SvgIcon className="text-warning">lucide:download</SvgIcon>
													<Typography variant="body2">Export Configuration</Typography>
												</Box>
											</CardContent>
										</Card>
									</Box>
								</CardContent>
							</Card>
						</Grid>
					</Grid>
				</div>
			}
		/>
	);
}

export default DashboardView;