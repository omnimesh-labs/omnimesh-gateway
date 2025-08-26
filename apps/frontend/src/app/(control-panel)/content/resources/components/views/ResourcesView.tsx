'use client';

import { useState, useCallback, useMemo } from 'react';
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow
} from '@/components/ui/table';
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuLabel,
	DropdownMenuSeparator,
	DropdownMenuTrigger
} from '@/components/ui/dropdown-menu';
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue
} from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle
} from '@/components/ui/alert-dialog';
import {
	Plus,
	Search,
	MoreHorizontal,
	Edit,
	Trash2,
	FileText,
	Link,
	Database,
	Globe,
	HardDrive,
	Code,
	ChevronLeft,
	ChevronRight
} from 'lucide-react';
import { useResources, useDeleteResource } from '../../api/hooks/useResources';
import ResourceFormDialog from '../forms/ResourceFormDialog';
import { Resource } from '@/lib/api';
import { useDebounce } from '@/hooks/useDebounce';

const RESOURCE_TYPE_ICONS = {
	file: FileText,
	url: Link,
	database: Database,
	api: Globe,
	memory: HardDrive,
	custom: Code
};

const RESOURCE_TYPE_COLORS = {
	file: 'bg-blue-100 text-blue-800',
	url: 'bg-green-100 text-green-800',
	database: 'bg-purple-100 text-purple-800',
	api: 'bg-orange-100 text-orange-800',
	memory: 'bg-gray-100 text-gray-800',
	custom: 'bg-pink-100 text-pink-800'
};

export default function ResourcesView() {
	const [searchQuery, setSearchQuery] = useState('');
	const [resourceType, setResourceType] = useState<string>('all');
	const [statusFilter, setStatusFilter] = useState<string>('all');
	const [currentPage, setCurrentPage] = useState(0);
	const [selectedResource, setSelectedResource] = useState<Resource | null>(null);
	const [isFormOpen, setIsFormOpen] = useState(false);
	const [deleteResourceId, setDeleteResourceId] = useState<string | null>(null);
	
	const pageSize = 20;
	const debouncedSearch = useDebounce(searchQuery, 300);
	
	const { data, isLoading, error } = useResources({
		search: debouncedSearch,
		limit: pageSize,
		offset: currentPage * pageSize,
		resourceType: resourceType === 'all' ? undefined : resourceType,
		isActive: statusFilter === 'all' ? undefined : statusFilter === 'active'
	});
	
	const deleteResourceMutation = useDeleteResource();
	
	const resources = data?.data || [];
	const totalPages = Math.ceil((data?.pagination?.total || 0) / pageSize);
	
	const handleEdit = useCallback((resource: Resource) => {
		setSelectedResource(resource);
		setIsFormOpen(true);
	}, []);
	
	const handleDelete = useCallback((id: string) => {
		setDeleteResourceId(id);
	}, []);
	
	const confirmDelete = useCallback(() => {
		if (deleteResourceId) {
			deleteResourceMutation.mutate(deleteResourceId, {
				onSettled: () => setDeleteResourceId(null)
			});
		}
	}, [deleteResourceId, deleteResourceMutation]);
	
	const handleCloseForm = useCallback(() => {
		setIsFormOpen(false);
		setSelectedResource(null);
	}, []);
	
	const formatFileSize = useCallback((bytes?: number) => {
		if (!bytes) return '-';
		const sizes = ['B', 'KB', 'MB', 'GB'];
		const i = Math.floor(Math.log(bytes) / Math.log(1024));
		return `${(bytes / Math.pow(1024, i)).toFixed(2)} ${sizes[i]}`;
	}, []);
	
	const renderTableContent = useMemo(() => {
		if (isLoading) {
			return (
				<>
					{Array.from({ length: 5 }).map((_, i) => (
						<TableRow key={i}>
							<TableCell colSpan={6}>
								<Skeleton className="h-12 w-full" />
							</TableCell>
						</TableRow>
					))}
				</>
			);
		}
		
		if (error) {
			return (
				<TableRow>
					<TableCell colSpan={6} className="text-center text-muted-foreground py-8">
						Error loading resources: {(error as any)?.message}
					</TableCell>
				</TableRow>
			);
		}
		
		if (resources.length === 0) {
			return (
				<TableRow>
					<TableCell colSpan={6} className="text-center text-muted-foreground py-8">
						No resources found
					</TableCell>
				</TableRow>
			);
		}
		
		return resources.map((resource) => {
			const Icon = RESOURCE_TYPE_ICONS[resource.resource_type as keyof typeof RESOURCE_TYPE_ICONS] || Code;
			const colorClass = RESOURCE_TYPE_COLORS[resource.resource_type as keyof typeof RESOURCE_TYPE_COLORS] || 'bg-gray-100 text-gray-800';
			
			return (
				<TableRow key={resource.id}>
					<TableCell className="font-medium">
						<div className="flex items-center gap-2">
							<Icon className="h-4 w-4 text-muted-foreground" />
							<span>{resource.name}</span>
						</div>
					</TableCell>
					<TableCell>
						<Badge variant="secondary" className={colorClass}>
							{resource.resource_type}
						</Badge>
					</TableCell>
					<TableCell className="max-w-xs truncate">
						<span className="text-sm text-muted-foreground" title={resource.uri}>
							{resource.uri}
						</span>
					</TableCell>
					<TableCell>{formatFileSize(resource.size_bytes)}</TableCell>
					<TableCell>
						<Badge variant={resource.is_active ? 'default' : 'secondary'}>
							{resource.is_active ? 'Active' : 'Inactive'}
						</Badge>
					</TableCell>
					<TableCell>
						<DropdownMenu>
							<DropdownMenuTrigger asChild>
								<Button variant="ghost" className="h-8 w-8 p-0">
									<span className="sr-only">Open menu</span>
									<MoreHorizontal className="h-4 w-4" />
								</Button>
							</DropdownMenuTrigger>
							<DropdownMenuContent align="end">
								<DropdownMenuLabel>Actions</DropdownMenuLabel>
								<DropdownMenuSeparator />
								<DropdownMenuItem onClick={() => handleEdit(resource)}>
									<Edit className="mr-2 h-4 w-4" />
									Edit
								</DropdownMenuItem>
								<DropdownMenuItem 
									onClick={() => handleDelete(resource.id)}
									className="text-red-600"
								>
									<Trash2 className="mr-2 h-4 w-4" />
									Delete
								</DropdownMenuItem>
							</DropdownMenuContent>
						</DropdownMenu>
					</TableCell>
				</TableRow>
			);
		});
	}, [isLoading, error, resources, formatFileSize, handleEdit, handleDelete]);
	
	return (
		<div className="container mx-auto py-6 space-y-6">
			<Card>
				<CardHeader>
					<div className="flex items-center justify-between">
						<div>
							<CardTitle>Resources</CardTitle>
							<CardDescription>
								Manage your MCP gateway resources
							</CardDescription>
						</div>
						<Button onClick={() => setIsFormOpen(true)}>
							<Plus className="mr-2 h-4 w-4" />
							Add Resource
						</Button>
					</div>
				</CardHeader>
				<CardContent className="space-y-4">
					{/* Filters */}
					<div className="flex flex-col sm:flex-row gap-4">
						<div className="relative flex-1">
							<Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
							<Input
								placeholder="Search resources..."
								value={searchQuery}
								onChange={(e) => setSearchQuery(e.target.value)}
								className="pl-8"
							/>
						</div>
						<Select value={resourceType} onValueChange={setResourceType}>
							<SelectTrigger className="w-[180px]">
								<SelectValue placeholder="Resource Type" />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="all">All Types</SelectItem>
								<SelectItem value="file">File</SelectItem>
								<SelectItem value="url">URL</SelectItem>
								<SelectItem value="database">Database</SelectItem>
								<SelectItem value="api">API</SelectItem>
								<SelectItem value="memory">Memory</SelectItem>
								<SelectItem value="custom">Custom</SelectItem>
							</SelectContent>
						</Select>
						<Select value={statusFilter} onValueChange={setStatusFilter}>
							<SelectTrigger className="w-[150px]">
								<SelectValue placeholder="Status" />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="all">All Status</SelectItem>
								<SelectItem value="active">Active</SelectItem>
								<SelectItem value="inactive">Inactive</SelectItem>
							</SelectContent>
						</Select>
					</div>
					
					{/* Table */}
					<div className="rounded-md border">
						<Table>
							<TableHeader>
								<TableRow>
									<TableHead>Name</TableHead>
									<TableHead>Type</TableHead>
									<TableHead>URI</TableHead>
									<TableHead>Size</TableHead>
									<TableHead>Status</TableHead>
									<TableHead className="w-[70px]"></TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{renderTableContent}
							</TableBody>
						</Table>
					</div>
					
					{/* Pagination */}
					{totalPages > 1 && (
						<div className="flex items-center justify-between">
							<p className="text-sm text-muted-foreground">
								Showing {currentPage * pageSize + 1} to{' '}
								{Math.min((currentPage + 1) * pageSize, data?.pagination?.total || 0)} of{' '}
								{data?.pagination?.total || 0} resources
							</p>
							<div className="flex items-center space-x-2">
								<Button
									variant="outline"
									size="sm"
									onClick={() => setCurrentPage(p => Math.max(0, p - 1))}
									disabled={currentPage === 0}
								>
									<ChevronLeft className="h-4 w-4" />
									Previous
								</Button>
								<div className="text-sm">
									Page {currentPage + 1} of {totalPages}
								</div>
								<Button
									variant="outline"
									size="sm"
									onClick={() => setCurrentPage(p => Math.min(totalPages - 1, p + 1))}
									disabled={currentPage === totalPages - 1}
								>
									Next
									<ChevronRight className="h-4 w-4" />
								</Button>
							</div>
						</div>
					)}
				</CardContent>
			</Card>
			
			{/* Form Dialog */}
			<ResourceFormDialog
				open={isFormOpen}
				onClose={handleCloseForm}
				resource={selectedResource}
			/>
			
			{/* Delete Confirmation */}
			<AlertDialog open={!!deleteResourceId} onOpenChange={() => setDeleteResourceId(null)}>
				<AlertDialogContent>
					<AlertDialogHeader>
						<AlertDialogTitle>Are you sure?</AlertDialogTitle>
						<AlertDialogDescription>
							This action cannot be undone. This will permanently delete the resource.
						</AlertDialogDescription>
					</AlertDialogHeader>
					<AlertDialogFooter>
						<AlertDialogCancel>Cancel</AlertDialogCancel>
						<AlertDialogAction onClick={confirmDelete}>Delete</AlertDialogAction>
					</AlertDialogFooter>
				</AlertDialogContent>
			</AlertDialog>
		</div>
	);
}