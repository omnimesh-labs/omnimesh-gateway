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
	Play,
	Wrench,
	Database,
	FileText,
	Globe,
	Terminal,
	Brain,
	Code,
	Package,
	ChevronLeft,
	ChevronRight,
	Lock,
	Unlock
} from 'lucide-react';
import { useTools, useDeleteTool } from '../../api/hooks/useTools';
import ToolFormDialog from '../forms/ToolFormDialog';
import ToolExecuteDialog from '../forms/ToolExecuteDialog';
import { Tool } from '@/lib/api';
import { useDebounce } from '@/hooks/useDebounce';

const CATEGORY_ICONS = {
	general: Wrench,
	data: Database,
	file: FileText,
	web: Globe,
	system: Terminal,
	ai: Brain,
	dev: Code,
	custom: Package
};

const CATEGORY_COLORS = {
	general: 'bg-gray-100 text-gray-800',
	data: 'bg-blue-100 text-blue-800',
	file: 'bg-green-100 text-green-800',
	web: 'bg-purple-100 text-purple-800',
	system: 'bg-orange-100 text-orange-800',
	ai: 'bg-pink-100 text-pink-800',
	dev: 'bg-indigo-100 text-indigo-800',
	custom: 'bg-yellow-100 text-yellow-800'
};

const IMPLEMENTATION_COLORS = {
	internal: 'bg-blue-100 text-blue-800',
	external: 'bg-green-100 text-green-800',
	webhook: 'bg-purple-100 text-purple-800',
	script: 'bg-orange-100 text-orange-800'
};

export default function ToolsView() {
	const [searchQuery, setSearchQuery] = useState('');
	const [category, setCategory] = useState<string>('all');
	const [statusFilter, setStatusFilter] = useState<string>('all');
	const [publicFilter, setPublicFilter] = useState<string>('all');
	const [currentPage, setCurrentPage] = useState(0);
	const [selectedTool, setSelectedTool] = useState<Tool | null>(null);
	const [isFormOpen, setIsFormOpen] = useState(false);
	const [executeToolId, setExecuteToolId] = useState<string | null>(null);
	const [deleteToolId, setDeleteToolId] = useState<string | null>(null);
	
	const pageSize = 20;
	const debouncedSearch = useDebounce(searchQuery, 300);
	
	const { data, isLoading, error } = useTools({
		search: debouncedSearch,
		limit: pageSize,
		offset: currentPage * pageSize,
		category: category === 'all' ? undefined : category,
		isActive: statusFilter === 'all' ? undefined : statusFilter === 'active',
		isPublic: publicFilter === 'all' ? undefined : publicFilter === 'public'
	});
	
	const deleteToolMutation = useDeleteTool();
	
	const tools = data?.data || [];
	const totalPages = Math.ceil((data?.pagination?.total || 0) / pageSize);
	
	const handleEdit = useCallback((tool: Tool) => {
		setSelectedTool(tool);
		setIsFormOpen(true);
	}, []);
	
	const handleExecute = useCallback((id: string) => {
		setExecuteToolId(id);
	}, []);
	
	const handleDelete = useCallback((id: string) => {
		setDeleteToolId(id);
	}, []);
	
	const confirmDelete = useCallback(() => {
		if (deleteToolId) {
			deleteToolMutation.mutate(deleteToolId, {
				onSettled: () => setDeleteToolId(null)
			});
		}
	}, [deleteToolId, deleteToolMutation]);
	
	const handleCloseForm = useCallback(() => {
		setIsFormOpen(false);
		setSelectedTool(null);
	}, []);
	
	const renderTableContent = useMemo(() => {
		if (isLoading) {
			return (
				<>
					{Array.from({ length: 5 }).map((_, i) => (
						<TableRow key={i}>
							<TableCell colSpan={8}>
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
					<TableCell colSpan={8} className="text-center text-muted-foreground py-8">
						Error loading tools: {(error as any)?.message}
					</TableCell>
				</TableRow>
			);
		}
		
		if (tools.length === 0) {
			return (
				<TableRow>
					<TableCell colSpan={8} className="text-center text-muted-foreground py-8">
						No tools found
					</TableCell>
				</TableRow>
			);
		}
		
		return tools.map((tool) => {
			const Icon = CATEGORY_ICONS[tool.category as keyof typeof CATEGORY_ICONS] || Package;
			const categoryColor = CATEGORY_COLORS[tool.category as keyof typeof CATEGORY_COLORS] || 'bg-gray-100 text-gray-800';
			const implColor = IMPLEMENTATION_COLORS[tool.implementation_type as keyof typeof IMPLEMENTATION_COLORS] || 'bg-gray-100 text-gray-800';
			
			return (
				<TableRow key={tool.id}>
					<TableCell className="font-medium">
						<div className="flex items-center gap-2">
							<Icon className="h-4 w-4 text-muted-foreground" />
							<span>{tool.name}</span>
						</div>
					</TableCell>
					<TableCell className="font-mono text-sm text-muted-foreground">
						{tool.function_name}
					</TableCell>
					<TableCell>
						<Badge variant="secondary" className={categoryColor}>
							{tool.category}
						</Badge>
					</TableCell>
					<TableCell>
						<Badge variant="secondary" className={implColor}>
							{tool.implementation_type}
						</Badge>
					</TableCell>
					<TableCell>{tool.usage_count || 0}</TableCell>
					<TableCell>
						{tool.is_public ? (
							<Unlock className="h-4 w-4 text-green-600" />
						) : (
							<Lock className="h-4 w-4 text-gray-400" />
						)}
					</TableCell>
					<TableCell>
						<Badge variant={tool.is_active ? 'default' : 'secondary'}>
							{tool.is_active ? 'Active' : 'Inactive'}
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
								<DropdownMenuItem onClick={() => handleExecute(tool.id)}>
									<Play className="mr-2 h-4 w-4" />
									Execute
								</DropdownMenuItem>
								<DropdownMenuItem onClick={() => handleEdit(tool)}>
									<Edit className="mr-2 h-4 w-4" />
									Edit
								</DropdownMenuItem>
								<DropdownMenuItem 
									onClick={() => handleDelete(tool.id)}
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
	}, [isLoading, error, tools, handleExecute, handleEdit, handleDelete]);
	
	return (
		<div className="container mx-auto py-6 space-y-6">
			<Card>
				<CardHeader>
					<div className="flex items-center justify-between">
						<div>
							<CardTitle>Tools</CardTitle>
							<CardDescription>
								Manage your MCP tools and functions
							</CardDescription>
						</div>
						<Button onClick={() => setIsFormOpen(true)}>
							<Plus className="mr-2 h-4 w-4" />
							Add Tool
						</Button>
					</div>
				</CardHeader>
				<CardContent className="space-y-4">
					{/* Filters */}
					<div className="flex flex-col sm:flex-row gap-4">
						<div className="relative flex-1">
							<Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
							<Input
								placeholder="Search tools by name or function..."
								value={searchQuery}
								onChange={(e) => setSearchQuery(e.target.value)}
								className="pl-8"
							/>
						</div>
						<Select value={category} onValueChange={setCategory}>
							<SelectTrigger className="w-[150px]">
								<SelectValue placeholder="Category" />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="all">All Categories</SelectItem>
								<SelectItem value="general">General</SelectItem>
								<SelectItem value="data">Data</SelectItem>
								<SelectItem value="file">File</SelectItem>
								<SelectItem value="web">Web</SelectItem>
								<SelectItem value="system">System</SelectItem>
								<SelectItem value="ai">AI</SelectItem>
								<SelectItem value="dev">Dev</SelectItem>
								<SelectItem value="custom">Custom</SelectItem>
							</SelectContent>
						</Select>
						<Select value={publicFilter} onValueChange={setPublicFilter}>
							<SelectTrigger className="w-[130px]">
								<SelectValue placeholder="Access" />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="all">All Access</SelectItem>
								<SelectItem value="public">Public</SelectItem>
								<SelectItem value="private">Private</SelectItem>
							</SelectContent>
						</Select>
						<Select value={statusFilter} onValueChange={setStatusFilter}>
							<SelectTrigger className="w-[130px]">
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
									<TableHead>Function</TableHead>
									<TableHead>Category</TableHead>
									<TableHead>Type</TableHead>
									<TableHead>Usage</TableHead>
									<TableHead>Access</TableHead>
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
								{data?.pagination?.total || 0} tools
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
			<ToolFormDialog
				open={isFormOpen}
				onClose={handleCloseForm}
				tool={selectedTool}
			/>
			
			{/* Execute Dialog */}
			<ToolExecuteDialog
				toolId={executeToolId}
				onClose={() => setExecuteToolId(null)}
			/>
			
			{/* Delete Confirmation */}
			<AlertDialog open={!!deleteToolId} onOpenChange={() => setDeleteToolId(null)}>
				<AlertDialogContent>
					<AlertDialogHeader>
						<AlertDialogTitle>Are you sure?</AlertDialogTitle>
						<AlertDialogDescription>
							This action cannot be undone. This will permanently delete the tool.
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