'use client';

import { useState, useMemo, useEffect, useCallback } from 'react';
import { MRT_ColumnDef } from 'material-react-table';
import PageSimple from '@fuse/core/PageSimple';
import { styled } from '@mui/material/styles';
import {
	Typography,
	Button,
	Chip,
	Box,
	Tabs,
	Tab,
	Card,
	CardContent,
	Grid,
	TextField,
	MenuItem,
	Stack,
	CircularProgress
} from '@mui/material';
import DataTable from '@/components/data-table/DataTable';
import SvgIcon from '@fuse/core/SvgIcon';
import { useSnackbar } from 'notistack';
import { adminApi, LogEntry, AuditLogEntry, LogQueryParams, AuditQueryParams, SystemStats } from '@/lib/api';
import LogDetailsModal from './components/LogDetailsModal';
import AuditDetailsModal from './components/AuditDetailsModal';

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

const getLogLevelColor = (level: string): 'default' | 'info' | 'warning' | 'error' => {
	switch (level) {
		case 'error':
			return 'error';
		case 'warn':
			return 'warning';
		case 'info':
			return 'info';
		default:
			return 'default';
	}
};

function LogsView() {
	const [tabValue, setTabValue] = useState(0);
	const [logLevel, setLogLevel] = useState('');
	const [timeRange, setTimeRange] = useState('1h');
	const [searchTerm, setSearchTerm] = useState('');
	const [userFilter, setUserFilter] = useState('');
	const [methodFilter, setMethodFilter] = useState('');

	// Audit filters
	const [auditResourceType, setAuditResourceType] = useState('');
	const [auditAction, setAuditAction] = useState('');
	const [auditActorId, setAuditActorId] = useState('');

	// Data and loading states
	const [logs, setLogs] = useState<LogEntry[]>([]);
	const [auditLogs, setAuditLogs] = useState<AuditLogEntry[]>([]);
	const [stats, setStats] = useState<SystemStats | null>(null);
	const [isLoading, setIsLoading] = useState(false);
	const [pagination, setPagination] = useState({ pageIndex: 0, pageSize: 25 });

	// Modal states
	const [selectedLogEntry, setSelectedLogEntry] = useState<LogEntry | null>(null);
	const [selectedAuditEntry, setSelectedAuditEntry] = useState<AuditLogEntry | null>(null);
	const [logModalOpen, setLogModalOpen] = useState(false);
	const [auditModalOpen, setAuditModalOpen] = useState(false);

	const { enqueueSnackbar } = useSnackbar();

	// Calculate time range for API calls
	const getTimeRange = () => {
		const now = new Date();
		const startTime = new Date();

		switch (timeRange) {
			case '15m':
				startTime.setMinutes(now.getMinutes() - 15);
				break;
			case '1h':
				startTime.setHours(now.getHours() - 1);
				break;
			case '24h':
				startTime.setDate(now.getDate() - 1);
				break;
			case '7d':
				startTime.setDate(now.getDate() - 7);
				break;
			default:
				startTime.setHours(now.getHours() - 1);
		}

		return {
			start_time: startTime.toISOString(),
			end_time: now.toISOString()
		};
	};

	// Fetch logs data
	const fetchLogs = useCallback(async () => {
		setIsLoading(true);
		try {
			const timeParams = getTimeRange();
			const params: LogQueryParams = {
				...timeParams,
				limit: pagination.pageSize,
				offset: pagination.pageIndex * pagination.pageSize,
				...(logLevel && { level: logLevel }),
				...(searchTerm && { search: searchTerm }),
				...(userFilter && { user_id: userFilter }),
				...(methodFilter && { method: methodFilter })
			};

			const data = await adminApi.getLogs(params);
			setLogs(data || []);
		} catch (error) {
			enqueueSnackbar('Failed to fetch logs', { variant: 'error' });
			console.error('Error fetching logs:', error);
		} finally {
			setIsLoading(false);
		}
	}, [timeRange, pagination, logLevel, searchTerm, userFilter, methodFilter, enqueueSnackbar]);

	// Fetch audit logs data
	const fetchAuditLogs = useCallback(async () => {
		setIsLoading(true);
		try {
			const timeParams = getTimeRange();
			const params: AuditQueryParams = {
				start_date: timeParams.start_time,
				end_date: timeParams.end_time,
				limit: pagination.pageSize,
				offset: pagination.pageIndex * pagination.pageSize,
				...(auditResourceType && { resource_type: auditResourceType }),
				...(auditAction && { action: auditAction }),
				...(auditActorId && { actor_id: auditActorId })
			};

			const response = await adminApi.getAuditLogs(params);
			setAuditLogs(response?.data || []);
		} catch (error) {
			enqueueSnackbar('Failed to fetch audit logs', { variant: 'error' });
			console.error('Error fetching audit logs:', error);
		} finally {
			setIsLoading(false);
		}
	}, [timeRange, pagination, auditResourceType, auditAction, auditActorId, enqueueSnackbar]);

	// Fetch stats data
	const fetchStats = useCallback(async () => {
		try {
			const data = await adminApi.getStats();
			setStats(data);
		} catch (error) {
			enqueueSnackbar('Failed to fetch statistics', { variant: 'error' });
			console.error('Error fetching stats:', error);
		}
	}, [enqueueSnackbar]);

	// Debounced search effect
	useEffect(() => {
		const timeoutId = setTimeout(() => {
			if (tabValue === 0) {
				fetchLogs();
			} else if (tabValue === 1) {
				fetchAuditLogs();
			}
		}, 500); // 500ms debounce for search inputs

		return () => clearTimeout(timeoutId);
	}, [searchTerm, userFilter, methodFilter, auditResourceType, auditAction, auditActorId]);

	// Immediate effect for tab changes and other filters
	useEffect(() => {
		if (tabValue === 0) {
			fetchLogs();
		} else if (tabValue === 1) {
			fetchAuditLogs();
		} else if (tabValue === 2) {
			fetchStats();
		}
	}, [tabValue, logLevel, timeRange, pagination]);

	const handleExportLogs = (format: 'json' | 'csv') => {
		enqueueSnackbar(`Export as ${format.toUpperCase()} functionality coming soon!`, { variant: 'info' });
	};

	// Modal handlers
	const handleLogRowClick = (logEntry: LogEntry) => {
		setSelectedLogEntry(logEntry);
		setLogModalOpen(true);
	};

	const handleAuditRowClick = (auditEntry: AuditLogEntry) => {
		setSelectedAuditEntry(auditEntry);
		setAuditModalOpen(true);
	};

	const handleCloseLogModal = () => {
		setLogModalOpen(false);
		setSelectedLogEntry(null);
	};

	const handleCloseAuditModal = () => {
		setAuditModalOpen(false);
		setSelectedAuditEntry(null);
	};

	const logColumns = useMemo<MRT_ColumnDef<LogEntry>[]>(
		() => [
			{
				accessorKey: 'timestamp',
				header: 'Timestamp',
				size: 180,
				Cell: ({ cell }) => {
					const timestamp = cell.getValue<string>();
					if (!timestamp) return '-';
					const date = new Date(timestamp);
					return date.toLocaleString();
				}
			},
			{
				accessorKey: 'level',
				header: 'Level',
				size: 100,
				Cell: ({ cell }) => {
					const level = cell.getValue<string>() || 'info';
					return (
						<Chip
							size="small"
							label={level.toUpperCase()}
							color={getLogLevelColor(level)}
						/>
					);
				}
			},
			{
				accessorKey: 'message',
				header: 'Message',
				size: 200,
				Cell: ({ cell }) => {
					const message = cell.getValue<string>();
					return (
						<Box sx={{ 
							overflow: 'hidden', 
							textOverflow: 'ellipsis',
							whiteSpace: 'nowrap',
							maxWidth: '200px'
						}}>
							{message || '-'}
						</Box>
					);
				}
			},
			{
				accessorKey: 'logger',
				header: 'Logger',
				size: 120
			},
			{
				accessorKey: 'request_id',
				header: 'Request ID',
				size: 150,
				Cell: ({ cell }) => {
					const requestId = cell.getValue<string>();
					return (
						<Typography variant="body2" sx={{ fontFamily: 'monospace', fontSize: '0.75rem' }}>
							{requestId ? `${requestId.slice(0, 8)}...` : '-'}
						</Typography>
					);
				}
			},
			{
				accessorKey: 'user_id',
				header: 'User',
				size: 150,
				Cell: ({ cell }) => {
					const userId = cell.getValue<string>();
					return (
						<Typography variant="body2" sx={{ fontFamily: 'monospace', fontSize: '0.75rem' }}>
							{userId ? `${userId.slice(0, 8)}...` : '-'}
						</Typography>
					);
				}
			},
			// {
			// 	accessorKey: 'data.response_size',
			// 	header: 'Response Size',
			// 	size: 120,
			// 	Cell: ({ cell }) => {
			// 		const size = cell.getValue<number>();
			// 		return size ? `${size} bytes` : '-';
			// 	}
			// },
			{
				accessorKey: 'data.status_code',
				header: 'Status',
				size: 80,
				Cell: ({ cell }) => {
					const status = cell.getValue<number>();
					return status || '-';
				}
			},
			// {
			// 	accessorKey: 'data.duration_ms',
			// 	header: 'Duration (ms)',
			// 	size: 120,
			// 	Cell: ({ cell }) => {
			// 		const duration = cell.getValue<number>();
			// 		return duration ? `${duration}ms` : '-';
			// 	}
			// },
			{
				accessorKey: 'environment',
				header: 'Environment',
				size: 100,
				Cell: ({ cell }) => {
					const environment = cell.getValue<string>();
					return (
						<Chip
							size="small"
							label={environment || 'unknown'}
							variant="outlined"
						/>
					);
				}
			}
		],
		[]
	);

	const auditColumns = useMemo<MRT_ColumnDef<AuditLogEntry>[]>(
		() => [
			{
				accessorKey: 'created_at',
				header: 'Timestamp',
				size: 180,
				Cell: ({ cell }) => {
					const date = new Date(cell.getValue<string>());
					return date.toLocaleString();
				}
			},
			{
				accessorKey: 'action',
				header: 'Action',
				size: 150
			},
			{
				accessorKey: 'resource_type',
				header: 'Resource Type',
				size: 120
			},
			{
				accessorKey: 'actor_id',
				header: 'Actor',
				size: 180
			},
			{
				accessorKey: 'actor_ip',
				header: 'IP Address',
				size: 120
			},
			{
				accessorKey: 'resource_id',
				header: 'Resource ID',
				size: 120
			},
			{
				accessorKey: 'metadata',
				header: 'Metadata',
				size: 200,
				Cell: ({ cell }) => {
					const metadata = cell.getValue<Record<string, unknown>>();
					return metadata ? JSON.stringify(metadata) : '';
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
							<Typography variant="h4">Logging & Audit</Typography>
							<Typography
								variant="body1"
								color="textSecondary"
								className="mt-1"
							>
								Monitor system activity, debug issues, and track administrative actions
							</Typography>
						</div>
						<Box className="flex gap-2">
							<Button
								variant="outlined"
								startIcon={<SvgIcon>lucide:refresh-cw</SvgIcon>}
								onClick={() => {
									if (tabValue === 0) fetchLogs();
									else if (tabValue === 1) fetchAuditLogs();
									else fetchStats();
								}}
								disabled={isLoading}
							>
								Refresh
							</Button>
							{/* <Button
								variant="outlined"
								startIcon={<SvgIcon>lucide:download</SvgIcon>}
								onClick={() => handleExportLogs('json')}
							>
								Export JSON
							</Button>
							<Button
								variant="outlined"
								startIcon={<SvgIcon>lucide:download</SvgIcon>}
								onClick={() => handleExportLogs('csv')}
							>
								Export CSV
							</Button> */}
						</Box>
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
								label={`System Logs (${logs?.length || 0})`}
								icon={<SvgIcon size={20}>lucide:file-text</SvgIcon>}
								iconPosition="start"
							/>
							<Tab
								label={`Audit Trail (${auditLogs?.length || 0})`}
								icon={<SvgIcon size={20}>lucide:shield</SvgIcon>}
								iconPosition="start"
							/>
							<Tab
								label="Statistics"
								icon={<SvgIcon size={20}>lucide:bar-chart</SvgIcon>}
								iconPosition="start"
							/>
						</Tabs>
					</Box>

					{(tabValue === 0 || tabValue === 1) && (
						<Box className="mb-6">
							<Card>
								<CardContent>
									<Box className="mb-3">
										<Typography
											variant="h6"
											className="mb-3"
										>
											Filters
											{isLoading && (
												<CircularProgress
													size={16}
													sx={{ ml: 2 }}
												/>
											)}
										</Typography>
									</Box>

									{/* Common filters */}
									<Stack
										direction="row"
										spacing={3}
										alignItems="center"
										className="mb-3"
									>
										<TextField
											select
											label="Time Range"
											value={timeRange}
											onChange={(e) => setTimeRange(e.target.value)}
											size="small"
											sx={{ minWidth: 140 }}
										>
											<MenuItem value="15m">Last 15 minutes</MenuItem>
											<MenuItem value="1h">Last hour</MenuItem>
											<MenuItem value="24h">Last 24 hours</MenuItem>
											<MenuItem value="7d">Last 7 days</MenuItem>
										</TextField>

										{tabValue === 0 && (
											<>
												<TextField
													select
													label="Log Level"
													value={logLevel}
													onChange={(e) => setLogLevel(e.target.value)}
													size="small"
													sx={{ minWidth: 140 }}
												>
													<MenuItem value="">All Levels</MenuItem>
													<MenuItem value="error">Error</MenuItem>
													<MenuItem value="warn">Warning</MenuItem>
													<MenuItem value="info">Info</MenuItem>
													<MenuItem value="debug">Debug</MenuItem>
												</TextField>
												<TextField
													label="Search"
													value={searchTerm}
													onChange={(e) => setSearchTerm(e.target.value)}
													size="small"
													sx={{ minWidth: 180 }}
													placeholder="Search messages..."
												/>
											</>
										)}
									</Stack>

									{/* Log-specific filters */}
									{tabValue === 0 && (
										<Stack
											direction="row"
											spacing={3}
											alignItems="center"
										>
											<TextField
												label="User ID"
												value={userFilter}
												onChange={(e) => setUserFilter(e.target.value)}
												size="small"
												sx={{ minWidth: 140 }}
												placeholder="Filter by user"
											/>
											<TextField
												label="Method"
												value={methodFilter}
												onChange={(e) => setMethodFilter(e.target.value)}
												size="small"
												sx={{ minWidth: 140 }}
												placeholder="RPC method"
											/>
										</Stack>
									)}

									{/* Audit-specific filters */}
									{tabValue === 1 && (
										<Stack
											direction="row"
											spacing={3}
											alignItems="center"
										>
											<TextField
												label="Resource Type"
												value={auditResourceType}
												onChange={(e) => setAuditResourceType(e.target.value)}
												size="small"
												sx={{ minWidth: 140 }}
												placeholder="server, namespace, etc."
											/>
											<TextField
												label="Action"
												value={auditAction}
												onChange={(e) => setAuditAction(e.target.value)}
												size="small"
												sx={{ minWidth: 140 }}
												placeholder="create, update, delete"
											/>
											<TextField
												label="Actor ID"
												value={auditActorId}
												onChange={(e) => setAuditActorId(e.target.value)}
												size="small"
												sx={{ minWidth: 140 }}
												placeholder="Filter by actor"
											/>
										</Stack>
									)}
								</CardContent>
							</Card>
						</Box>
					)}

					{tabValue === 0 && (
						<DataTable
							columns={logColumns}
							data={logs}
							enableRowActions={false}
							enableRowSelection={false}
							state={{
								isLoading,
								pagination
							}}
							onPaginationChange={setPagination}
							manualPagination
							enableGlobalFilter={false}
							initialState={{
								pagination: { pageIndex: 0, pageSize: 25 }
							}}
							muiTableBodyRowProps={({ row }) => ({
								onClick: () => handleLogRowClick(row.original),
								sx: {
									cursor: 'pointer',
									'&:hover': {
										backgroundColor: 'action.hover'
									}
								}
							})}
						/>
					)}

					{tabValue === 1 && (
						<DataTable
							columns={auditColumns}
							data={auditLogs}
							enableRowActions={false}
							enableRowSelection={false}
							state={{
								isLoading,
								pagination
							}}
							onPaginationChange={setPagination}
							manualPagination
							enableGlobalFilter={false}
							initialState={{
								pagination: { pageIndex: 0, pageSize: 25 }
							}}
							muiTableBodyRowProps={({ row }) => ({
								onClick: () => handleAuditRowClick(row.original),
								sx: {
									cursor: 'pointer',
									'&:hover': {
										backgroundColor: 'action.hover'
									}
								}
							})}
						/>
					)}

					{tabValue === 2 && (
						<Grid
							container
							spacing={3}
						>
							<Grid size={{ xs: 12, md: 6, lg: 3 }}>
								<Card>
									<CardContent>
										<Typography
											variant="h6"
											gutterBottom
										>
											Users
										</Typography>
										<Box className="flex items-center justify-between">
											<Box>
												<Typography variant="h4">{stats?.users?.total || 0}</Typography>
												<Typography
													variant="body2"
													color="textSecondary"
												>
													Total
												</Typography>
											</Box>
											<Box className="text-right">
												<Typography variant="h4">{stats?.users?.active || 0}</Typography>
												<Typography
													variant="body2"
													color="textSecondary"
												>
													Active
												</Typography>
											</Box>
										</Box>
									</CardContent>
								</Card>
							</Grid>

							<Grid size={{ xs: 12, md: 6, lg: 3 }}>
								<Card>
									<CardContent>
										<Typography
											variant="h6"
											gutterBottom
										>
											Servers
										</Typography>
										<Box className="flex items-center justify-between">
											<Box>
												<Typography variant="h4">{stats?.servers?.total || 0}</Typography>
												<Typography
													variant="body2"
													color="textSecondary"
												>
													Total
												</Typography>
											</Box>
											<Box className="text-right">
												<Typography variant="h4">{stats?.servers?.healthy || 0}</Typography>
												<Typography
													variant="body2"
													color="textSecondary"
												>
													Healthy
												</Typography>
											</Box>
										</Box>
									</CardContent>
								</Card>
							</Grid>

							<Grid size={{ xs: 12, md: 6, lg: 3 }}>
								<Card>
									<CardContent>
										<Typography
											variant="h6"
											gutterBottom
										>
											Requests
										</Typography>
										<Box className="space-y-2">
											<Box className="flex justify-between">
												<Typography variant="body2">Total</Typography>
												<Typography variant="body2">{stats?.requests?.total || 0}</Typography>
											</Box>
											<Box className="flex justify-between">
												<Typography variant="body2">Success</Typography>
												<Typography variant="body2">
													{stats?.requests?.successful || 0}
												</Typography>
											</Box>
											<Box className="flex justify-between">
												<Typography variant="body2">Failed</Typography>
												<Typography variant="body2">{stats?.requests?.failed || 0}</Typography>
											</Box>
										</Box>
									</CardContent>
								</Card>
							</Grid>

							<Grid size={{ xs: 12, md: 6, lg: 3 }}>
								<Card>
									<CardContent>
										<Typography
											variant="h6"
											gutterBottom
										>
											Rate Limits
										</Typography>
										<Box className="flex items-center justify-between">
											<Box>
												<Typography variant="h4">
													{stats?.rate_limits?.total_limits || 0}
												</Typography>
												<Typography
													variant="body2"
													color="textSecondary"
												>
													Active Limits
												</Typography>
											</Box>
											<Box className="text-right">
												<Typography variant="h4">{stats?.rate_limits?.blocked || 0}</Typography>
												<Typography
													variant="body2"
													color="textSecondary"
												>
													Blocked
												</Typography>
											</Box>
										</Box>
									</CardContent>
								</Card>
							</Grid>
						</Grid>
					)}

					{/* Modals */}
					<LogDetailsModal
						open={logModalOpen}
						onClose={handleCloseLogModal}
						logEntry={selectedLogEntry}
					/>
					<AuditDetailsModal
						open={auditModalOpen}
						onClose={handleCloseAuditModal}
						auditEntry={selectedAuditEntry}
					/>
				</div>
			}
		/>
	);
}

export default LogsView;
