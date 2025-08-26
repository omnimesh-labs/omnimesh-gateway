'use client';

import { useState, useMemo } from 'react';
import { MRT_ColumnDef } from 'material-react-table';
import { Typography, Button, Chip, Box, IconButton, Tooltip } from '@mui/material';
import DataTable from '@/components/data-table/DataTable';
import SvgIcon from '@fuse/core/SvgIcon';
import { MCPPackage, CreateServerRequest } from '@/lib/api';

interface AvailableServersTableProps {
	servers: MCPPackage[];
	loading: boolean;
	onRegisterServer: (serverData: CreateServerRequest) => void;
	registering: boolean;
}

export default function AvailableServersTable({
	servers,
	loading,
	onRegisterServer,
	registering
}: AvailableServersTableProps) {
	const [selectedPackage, setSelectedPackage] = useState<string | null>(null);

	const handleRegisterFromPackage = (pkg: MCPPackage) => {
		setSelectedPackage(pkg.name);

		const serverData = {
			name: pkg.name.replace(/^@.*\//, ''), // Remove scope prefix
			description: pkg.description,
			protocol: 'stdio',
			command: pkg.command,
			args: pkg.args,
			environment: pkg.envs,
			version: '1.0.0',
			metadata: {
				github_url: pkg.githubUrl,
				package_registry: pkg.package_registry,
				package_name: pkg.package_name,
				source: 'discovery'
			}
		};

		onRegisterServer(serverData);
	};

	const columns = useMemo<MRT_ColumnDef<MCPPackage>[]>(
		() => [
			{
				accessorKey: 'name',
				header: 'Package',
				size: 250,
				Cell: ({ row }) => (
					<Box className="flex items-start space-x-3">
						<SvgIcon
							size={20}
							className="mt-1"
						>
							lucide:package
						</SvgIcon>
						<Box>
							<Typography
								variant="body2"
								className="font-medium"
							>
								{row.original.name}
							</Typography>
							<Typography
								variant="caption"
								color="textSecondary"
								className="line-clamp-2"
							>
								{row.original.description}
							</Typography>
						</Box>
					</Box>
				)
			},
			{
				accessorKey: 'package_registry',
				header: 'Registry',
				size: 100,
				Cell: ({ cell }) => (
					<Chip
						size="small"
						label={cell.getValue<string>()}
						variant="outlined"
					/>
				)
			},
			{
				accessorKey: 'github_stars',
				header: 'Stars',
				size: 80,
				Cell: ({ row }) => (
					<Box className="flex items-center space-x-1">
						<SvgIcon
							size={16}
							className="text-yellow-500"
						>
							lucide:star
						</SvgIcon>
						<Typography variant="body2">{row.original.github_stars.toLocaleString()}</Typography>
					</Box>
				)
			},
			{
				accessorKey: 'package_download_count',
				header: 'Downloads',
				size: 100,
				Cell: ({ cell }) => <Typography variant="body2">{cell.getValue<number>().toLocaleString()}</Typography>
			},
			{
				accessorKey: 'githubUrl',
				header: 'Links',
				size: 120,
				enableSorting: false,
				Cell: ({ row }) => (
					<Box className="flex items-center space-x-1">
						<Tooltip title="View on GitHub">
							<IconButton
								size="small"
								component="a"
								href={row.original.githubUrl}
								target="_blank"
								rel="noopener noreferrer"
							>
								<SvgIcon size={16}>lucide:github</SvgIcon>
							</IconButton>
						</Tooltip>
						<Tooltip title="View Package">
							<IconButton
								size="small"
								component="a"
								href={`https://www.npmjs.com/package/${row.original.package_name}`}
								target="_blank"
								rel="noopener noreferrer"
							>
								<SvgIcon size={16}>lucide:external-link</SvgIcon>
							</IconButton>
						</Tooltip>
					</Box>
				)
			}
		],
		[]
	);

	if (loading && servers.length === 0) {
		return (
			<Box className="flex items-center justify-center py-12">
				<Box className="text-center">
					<SvgIcon
						size={48}
						className="mb-4 text-gray-300"
					>
						lucide:loader-2
					</SvgIcon>
					<Typography
						variant="body1"
						color="textSecondary"
					>
						Loading available servers...
					</Typography>
				</Box>
			</Box>
		);
	}

	if (servers.length === 0) {
		return (
			<Box className="flex items-center justify-center py-12">
				<Box className="text-center">
					<SvgIcon
						size={64}
						className="mb-4 text-gray-300"
					>
						lucide:search-x
					</SvgIcon>
					<Typography
						variant="h6"
						className="mb-2"
					>
						No servers found
					</Typography>
					<Typography
						variant="body2"
						color="textSecondary"
					>
						No MCP servers are currently available in the discovery registry.
					</Typography>
				</Box>
			</Box>
		);
	}

	return (
		<DataTable
			columns={columns}
			data={servers}
			state={{
				isLoading: loading
			}}
			enableRowActions
			renderRowActions={({ row }) => (
				<Button
					size="small"
					variant="contained"
					color="primary"
					startIcon={
						selectedPackage === row.original.name && registering ? (
							<SvgIcon size={16}>lucide:loader-2</SvgIcon>
						) : (
							<SvgIcon size={16}>lucide:download</SvgIcon>
						)
					}
					onClick={() => handleRegisterFromPackage(row.original)}
					disabled={registering}
				>
					{selectedPackage === row.original.name && registering ? 'Adding...' : 'Add Server'}
				</Button>
			)}
			initialState={{
				pagination: {
					pageIndex: 0,
					pageSize: 15
				},
				sorting: [
					{
						id: 'github_stars',
						desc: true
					}
				]
			}}
			enableGlobalFilter
			enableColumnFilters
		/>
	);
}
