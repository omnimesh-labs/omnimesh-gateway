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
	MessageSquare,
	Code,
	Palette,
	BookOpen,
	Briefcase,
	Sparkles,
	ChevronLeft,
	ChevronRight
} from 'lucide-react';
import { usePrompts, useDeletePrompt } from '../../api/hooks/usePrompts';
import PromptFormDialog from '../forms/PromptFormDialog';
import PromptTestDialog from '../forms/PromptTestDialog';
import { Prompt } from '@/lib/api';
import { useDebounce } from '@/hooks/useDebounce';

const CATEGORY_ICONS = {
	general: MessageSquare,
	coding: Code,
	analysis: Sparkles,
	creative: Palette,
	educational: BookOpen,
	business: Briefcase,
	custom: MessageSquare
};

const CATEGORY_COLORS = {
	general: 'bg-gray-100 text-gray-800',
	coding: 'bg-blue-100 text-blue-800',
	analysis: 'bg-purple-100 text-purple-800',
	creative: 'bg-pink-100 text-pink-800',
	educational: 'bg-green-100 text-green-800',
	business: 'bg-orange-100 text-orange-800',
	custom: 'bg-indigo-100 text-indigo-800'
};

export default function PromptsView() {
	const [searchQuery, setSearchQuery] = useState('');
	const [category, setCategory] = useState<string>('all');
	const [statusFilter, setStatusFilter] = useState<string>('all');
	const [currentPage, setCurrentPage] = useState(0);
	const [selectedPrompt, setSelectedPrompt] = useState<Prompt | null>(null);
	const [isFormOpen, setIsFormOpen] = useState(false);
	const [testPromptId, setTestPromptId] = useState<string | null>(null);
	const [deletePromptId, setDeletePromptId] = useState<string | null>(null);
	
	const pageSize = 20;
	const debouncedSearch = useDebounce(searchQuery, 300);
	
	const { data, isLoading, error } = usePrompts({
		search: debouncedSearch,
		limit: pageSize,
		offset: currentPage * pageSize,
		category: category === 'all' ? undefined : category,
		isActive: statusFilter === 'all' ? undefined : statusFilter === 'active'
	});
	
	const deletePromptMutation = useDeletePrompt();
	
	const prompts = data?.data || [];
	const totalPages = Math.ceil((data?.pagination?.total || 0) / pageSize);
	
	const handleEdit = useCallback((prompt: Prompt) => {
		setSelectedPrompt(prompt);
		setIsFormOpen(true);
	}, []);
	
	const handleTest = useCallback((id: string) => {
		setTestPromptId(id);
	}, []);
	
	const handleDelete = useCallback((id: string) => {
		setDeletePromptId(id);
	}, []);
	
	const confirmDelete = useCallback(() => {
		if (deletePromptId) {
			deletePromptMutation.mutate(deletePromptId, {
				onSettled: () => setDeletePromptId(null)
			});
		}
	}, [deletePromptId, deletePromptMutation]);
	
	const handleCloseForm = useCallback(() => {
		setIsFormOpen(false);
		setSelectedPrompt(null);
	}, []);
	
	const truncateText = useCallback((text: string, maxLength: number = 100) => {
		if (text.length <= maxLength) return text;
		return text.substring(0, maxLength) + '...';
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
						Error loading prompts: {(error as any)?.message}
					</TableCell>
				</TableRow>
			);
		}
		
		if (prompts.length === 0) {
			return (
				<TableRow>
					<TableCell colSpan={6} className="text-center text-muted-foreground py-8">
						No prompts found
					</TableCell>
				</TableRow>
			);
		}
		
		return prompts.map((prompt) => {
			const Icon = CATEGORY_ICONS[prompt.category as keyof typeof CATEGORY_ICONS] || MessageSquare;
			const colorClass = CATEGORY_COLORS[prompt.category as keyof typeof CATEGORY_COLORS] || 'bg-gray-100 text-gray-800';
			
			return (
				<TableRow key={prompt.id}>
					<TableCell className="font-medium">
						<div className="flex items-center gap-2">
							<Icon className="h-4 w-4 text-muted-foreground" />
							<span>{prompt.name}</span>
						</div>
					</TableCell>
					<TableCell>
						<Badge variant="secondary" className={colorClass}>
							{prompt.category}
						</Badge>
					</TableCell>
					<TableCell className="max-w-xs">
						<span className="text-sm text-muted-foreground" title={prompt.prompt_template}>
							{truncateText(prompt.prompt_template)}
						</span>
					</TableCell>
					<TableCell>{prompt.usage_count || 0}</TableCell>
					<TableCell>
						<Badge variant={prompt.is_active ? 'default' : 'secondary'}>
							{prompt.is_active ? 'Active' : 'Inactive'}
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
								<DropdownMenuItem onClick={() => handleTest(prompt.id)}>
									<Play className="mr-2 h-4 w-4" />
									Test
								</DropdownMenuItem>
								<DropdownMenuItem onClick={() => handleEdit(prompt)}>
									<Edit className="mr-2 h-4 w-4" />
									Edit
								</DropdownMenuItem>
								<DropdownMenuItem 
									onClick={() => handleDelete(prompt.id)}
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
	}, [isLoading, error, prompts, truncateText, handleTest, handleEdit, handleDelete]);
	
	return (
		<div className="container mx-auto py-6 space-y-6">
			<Card>
				<CardHeader>
					<div className="flex items-center justify-between">
						<div>
							<CardTitle>Prompts</CardTitle>
							<CardDescription>
								Manage your prompt templates
							</CardDescription>
						</div>
						<Button onClick={() => setIsFormOpen(true)}>
							<Plus className="mr-2 h-4 w-4" />
							Add Prompt
						</Button>
					</div>
				</CardHeader>
				<CardContent className="space-y-4">
					{/* Filters */}
					<div className="flex flex-col sm:flex-row gap-4">
						<div className="relative flex-1">
							<Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
							<Input
								placeholder="Search prompts..."
								value={searchQuery}
								onChange={(e) => setSearchQuery(e.target.value)}
								className="pl-8"
							/>
						</div>
						<Select value={category} onValueChange={setCategory}>
							<SelectTrigger className="w-[180px]">
								<SelectValue placeholder="Category" />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="all">All Categories</SelectItem>
								<SelectItem value="general">General</SelectItem>
								<SelectItem value="coding">Coding</SelectItem>
								<SelectItem value="analysis">Analysis</SelectItem>
								<SelectItem value="creative">Creative</SelectItem>
								<SelectItem value="educational">Educational</SelectItem>
								<SelectItem value="business">Business</SelectItem>
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
									<TableHead>Category</TableHead>
									<TableHead>Template</TableHead>
									<TableHead>Usage</TableHead>
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
								{data?.pagination?.total || 0} prompts
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
			<PromptFormDialog
				open={isFormOpen}
				onClose={handleCloseForm}
				prompt={selectedPrompt}
			/>
			
			{/* Test Dialog */}
			<PromptTestDialog
				promptId={testPromptId}
				onClose={() => setTestPromptId(null)}
			/>
			
			{/* Delete Confirmation */}
			<AlertDialog open={!!deletePromptId} onOpenChange={() => setDeletePromptId(null)}>
				<AlertDialogContent>
					<AlertDialogHeader>
						<AlertDialogTitle>Are you sure?</AlertDialogTitle>
						<AlertDialogDescription>
							This action cannot be undone. This will permanently delete the prompt.
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