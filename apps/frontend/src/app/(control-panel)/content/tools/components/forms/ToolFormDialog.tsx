'use client';

import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle
} from '@/components/ui/dialog';
import {
	Form,
	FormControl,
	FormDescription,
	FormField,
	FormItem,
	FormLabel,
	FormMessage
} from '@/components/ui/form';
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue
} from '@/components/ui/select';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Button } from '@/components/ui/button';
import { Switch } from '@/components/ui/switch';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { X } from 'lucide-react';
import { useCreateTool, useUpdateTool } from '../../api/hooks/useTools';
import { Tool } from '@/lib/api';

const toolSchema = z.object({
	name: z.string().min(1, 'Name is required').max(255),
	description: z.string().optional(),
	function_name: z.string().min(1, 'Function name is required').regex(/^[a-zA-Z_][a-zA-Z0-9_]*$/, 'Must be a valid function name'),
	category: z.enum(['general', 'data', 'file', 'web', 'system', 'ai', 'dev', 'custom']),
	implementation_type: z.enum(['internal', 'external', 'webhook', 'script']).optional(),
	endpoint_url: z.string().optional(),
	timeout_seconds: z.number().min(1).max(300).optional(),
	max_retries: z.number().min(0).max(10).optional(),
	is_active: z.boolean().default(true),
	is_public: z.boolean().default(false),
	schema: z.string().optional(),
	examples: z.string().optional(),
	documentation: z.string().optional(),
	tags: z.array(z.string()).optional(),
	metadata: z.string().optional()
});

type ToolFormData = z.infer<typeof toolSchema>;

interface ToolFormDialogProps {
	open: boolean;
	onClose: () => void;
	tool?: Tool | null;
}

export default function ToolFormDialog({ open, onClose, tool }: ToolFormDialogProps) {
	const isEdit = !!tool;
	const createMutation = useCreateTool();
	const updateMutation = useUpdateTool();
	
	const form = useForm<ToolFormData>({
		resolver: zodResolver(toolSchema),
		defaultValues: {
			name: '',
			description: '',
			function_name: '',
			category: 'general',
			implementation_type: 'internal',
			endpoint_url: '',
			timeout_seconds: 30,
			max_retries: 3,
			is_active: true,
			is_public: false,
			schema: '',
			examples: '',
			documentation: '',
			tags: [],
			metadata: ''
		}
	});
	
	useEffect(() => {
		if (tool) {
			form.reset({
				name: tool.name,
				description: tool.description || '',
				function_name: tool.function_name,
				category: tool.category as any,
				implementation_type: tool.implementation_type as any,
				endpoint_url: tool.endpoint_url || '',
				timeout_seconds: tool.timeout_seconds,
				max_retries: tool.max_retries,
				is_active: tool.is_active,
				is_public: tool.is_public,
				schema: tool.schema ? JSON.stringify(tool.schema, null, 2) : '',
				examples: tool.examples ? JSON.stringify(tool.examples, null, 2) : '',
				documentation: tool.documentation || '',
				tags: tool.tags || [],
				metadata: tool.metadata ? JSON.stringify(tool.metadata, null, 2) : ''
			});
		} else {
			form.reset({
				name: '',
				description: '',
				function_name: '',
				category: 'general',
				implementation_type: 'internal',
				endpoint_url: '',
				timeout_seconds: 30,
				max_retries: 3,
				is_active: true,
				is_public: false,
				schema: '',
				examples: '',
				documentation: '',
				tags: [],
				metadata: ''
			});
		}
	}, [tool, form]);
	
	const handleSubmit = async (data: ToolFormData) => {
		try {
			const schema = data.schema ? JSON.parse(data.schema) : undefined;
			const examples = data.examples ? JSON.parse(data.examples) : undefined;
			const metadata = data.metadata ? JSON.parse(data.metadata) : undefined;
			
			const payload = {
				...data,
				schema,
				examples,
				metadata,
				tags: data.tags?.filter(tag => tag.trim() !== '')
			};
			
			if (isEdit && tool) {
				await updateMutation.mutateAsync({
					id: tool.id,
					data: payload
				});
			} else {
				await createMutation.mutateAsync(payload);
			}
			
			onClose();
		} catch (error) {
			// Error is handled by the mutation hooks
		}
	};
	
	const handleAddTag = (tag: string) => {
		const currentTags = form.getValues('tags') || [];
		if (tag && !currentTags.includes(tag)) {
			form.setValue('tags', [...currentTags, tag]);
		}
	};
	
	const handleRemoveTag = (tagToRemove: string) => {
		const currentTags = form.getValues('tags') || [];
		form.setValue('tags', currentTags.filter(tag => tag !== tagToRemove));
	};
	
	const implementationType = form.watch('implementation_type');
	
	return (
		<Dialog open={open} onOpenChange={onClose}>
			<DialogContent className="sm:max-w-[800px] max-h-[90vh] overflow-y-auto">
				<DialogHeader>
					<DialogTitle>{isEdit ? 'Edit Tool' : 'Create Tool'}</DialogTitle>
					<DialogDescription>
						{isEdit ? 'Update the tool configuration' : 'Add a new tool to your gateway'}
					</DialogDescription>
				</DialogHeader>
				
				<Form {...form}>
					<form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-4">
						<Tabs defaultValue="basic" className="w-full">
							<TabsList className="grid w-full grid-cols-3">
								<TabsTrigger value="basic">Basic Info</TabsTrigger>
								<TabsTrigger value="implementation">Implementation</TabsTrigger>
								<TabsTrigger value="advanced">Advanced</TabsTrigger>
							</TabsList>
							
							<TabsContent value="basic" className="space-y-4 mt-4">
								<FormField
									control={form.control}
									name="name"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Name</FormLabel>
											<FormControl>
												<Input placeholder="My Awesome Tool" {...field} />
											</FormControl>
											<FormMessage />
										</FormItem>
									)}
								/>
								
								<FormField
									control={form.control}
									name="function_name"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Function Name</FormLabel>
											<FormControl>
												<Input placeholder="my_awesome_tool" {...field} />
											</FormControl>
											<FormDescription>
												Must be a valid function identifier (letters, numbers, underscores)
											</FormDescription>
											<FormMessage />
										</FormItem>
									)}
								/>
								
								<FormField
									control={form.control}
									name="description"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Description</FormLabel>
											<FormControl>
												<Textarea
													placeholder="Describe what this tool does..."
													{...field}
													rows={3}
												/>
											</FormControl>
											<FormMessage />
										</FormItem>
									)}
								/>
								
								<FormField
									control={form.control}
									name="category"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Category</FormLabel>
											<Select onValueChange={field.onChange} value={field.value}>
												<FormControl>
													<SelectTrigger>
														<SelectValue placeholder="Select category" />
													</SelectTrigger>
												</FormControl>
												<SelectContent>
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
											<FormMessage />
										</FormItem>
									)}
								/>
							</TabsContent>
							
							<TabsContent value="implementation" className="space-y-4 mt-4">
								<FormField
									control={form.control}
									name="implementation_type"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Implementation Type</FormLabel>
											<Select onValueChange={field.onChange} value={field.value}>
												<FormControl>
													<SelectTrigger>
														<SelectValue placeholder="Select type" />
													</SelectTrigger>
												</FormControl>
												<SelectContent>
													<SelectItem value="internal">Internal</SelectItem>
													<SelectItem value="external">External API</SelectItem>
													<SelectItem value="webhook">Webhook</SelectItem>
													<SelectItem value="script">Script</SelectItem>
												</SelectContent>
											</Select>
											<FormMessage />
										</FormItem>
									)}
								/>
								
								{(implementationType === 'external' || implementationType === 'webhook') && (
									<FormField
										control={form.control}
										name="endpoint_url"
										render={({ field }) => (
											<FormItem>
												<FormLabel>Endpoint URL</FormLabel>
												<FormControl>
													<Input placeholder="https://api.example.com/endpoint" {...field} />
												</FormControl>
												<FormDescription>
													The URL to call for this tool
												</FormDescription>
												<FormMessage />
											</FormItem>
										)}
									/>
								)}
								
								<div className="grid grid-cols-2 gap-4">
									<FormField
										control={form.control}
										name="timeout_seconds"
										render={({ field }) => (
											<FormItem>
												<FormLabel>Timeout (seconds)</FormLabel>
												<FormControl>
													<Input
														type="number"
														placeholder="30"
														{...field}
														onChange={(e) => field.onChange(parseInt(e.target.value) || 30)}
													/>
												</FormControl>
												<FormDescription>Max execution time</FormDescription>
												<FormMessage />
											</FormItem>
										)}
									/>
									
									<FormField
										control={form.control}
										name="max_retries"
										render={({ field }) => (
											<FormItem>
												<FormLabel>Max Retries</FormLabel>
												<FormControl>
													<Input
														type="number"
														placeholder="3"
														{...field}
														onChange={(e) => field.onChange(parseInt(e.target.value) || 3)}
													/>
												</FormControl>
												<FormDescription>Retry on failure</FormDescription>
												<FormMessage />
											</FormItem>
										)}
									/>
								</div>
								
								<FormField
									control={form.control}
									name="schema"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Schema (JSON)</FormLabel>
											<FormControl>
												<Textarea
													placeholder='{"type": "object", "properties": {...}}'
													{...field}
													rows={6}
													className="font-mono text-sm"
												/>
											</FormControl>
											<FormDescription>
												JSON Schema for tool parameters
											</FormDescription>
											<FormMessage />
										</FormItem>
									)}
								/>
							</TabsContent>
							
							<TabsContent value="advanced" className="space-y-4 mt-4">
								<FormField
									control={form.control}
									name="documentation"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Documentation</FormLabel>
											<FormControl>
												<Textarea
													placeholder="Detailed documentation for this tool..."
													{...field}
													rows={4}
												/>
											</FormControl>
											<FormDescription>
												Markdown supported
											</FormDescription>
											<FormMessage />
										</FormItem>
									)}
								/>
								
								<FormField
									control={form.control}
									name="examples"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Examples (JSON)</FormLabel>
											<FormControl>
												<Textarea
													placeholder='[{"input": {...}, "output": {...}}]'
													{...field}
													rows={4}
													className="font-mono text-sm"
												/>
											</FormControl>
											<FormDescription>
												Usage examples as JSON array
											</FormDescription>
											<FormMessage />
										</FormItem>
									)}
								/>
								
								<FormField
									control={form.control}
									name="tags"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Tags</FormLabel>
											<FormControl>
												<div className="space-y-2">
													<Input
														placeholder="Add tag and press Enter"
														onKeyDown={(e) => {
															if (e.key === 'Enter') {
																e.preventDefault();
																handleAddTag((e.target as HTMLInputElement).value);
																(e.target as HTMLInputElement).value = '';
															}
														}}
													/>
													<div className="flex flex-wrap gap-2">
														{field.value?.map((tag) => (
															<Badge key={tag} variant="secondary" className="gap-1">
																{tag}
																<X
																	className="h-3 w-3 cursor-pointer"
																	onClick={() => handleRemoveTag(tag)}
																/>
															</Badge>
														))}
													</div>
												</div>
											</FormControl>
											<FormDescription>Press Enter to add tags</FormDescription>
											<FormMessage />
										</FormItem>
									)}
								/>
								
								<FormField
									control={form.control}
									name="metadata"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Metadata (JSON)</FormLabel>
											<FormControl>
												<Textarea
													placeholder='{"key": "value"}'
													{...field}
													rows={3}
													className="font-mono text-sm"
												/>
											</FormControl>
											<FormDescription>
												Additional metadata in JSON format
											</FormDescription>
											<FormMessage />
										</FormItem>
									)}
								/>
								
								<div className="space-y-4">
									<FormField
										control={form.control}
										name="is_public"
										render={({ field }) => (
											<FormItem className="flex items-center justify-between rounded-lg border p-3">
												<div className="space-y-0.5">
													<FormLabel>Public</FormLabel>
													<FormDescription>
														Make this tool available to all organizations
													</FormDescription>
												</div>
												<FormControl>
													<Switch
														checked={field.value}
														onCheckedChange={field.onChange}
													/>
												</FormControl>
											</FormItem>
										)}
									/>
									
									<FormField
										control={form.control}
										name="is_active"
										render={({ field }) => (
											<FormItem className="flex items-center justify-between rounded-lg border p-3">
												<div className="space-y-0.5">
													<FormLabel>Active</FormLabel>
													<FormDescription>
														Enable or disable this tool
													</FormDescription>
												</div>
												<FormControl>
													<Switch
														checked={field.value}
														onCheckedChange={field.onChange}
													/>
												</FormControl>
											</FormItem>
										)}
									/>
								</div>
							</TabsContent>
						</Tabs>
						
						<DialogFooter>
							<Button type="button" variant="outline" onClick={onClose}>
								Cancel
							</Button>
							<Button
								type="submit"
								disabled={createMutation.isPending || updateMutation.isPending}
							>
								{createMutation.isPending || updateMutation.isPending
									? 'Saving...'
									: isEdit
									? 'Update'
									: 'Create'}
							</Button>
						</DialogFooter>
					</form>
				</Form>
			</DialogContent>
		</Dialog>
	);
}