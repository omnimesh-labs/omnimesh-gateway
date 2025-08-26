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
import { X, Plus } from 'lucide-react';
import { useCreatePrompt, useUpdatePrompt } from '../../api/hooks/usePrompts';
import { Prompt } from '@/lib/api';

const promptSchema = z.object({
	name: z.string().min(1, 'Name is required').max(255),
	description: z.string().optional(),
	prompt_template: z.string().min(1, 'Template is required'),
	category: z.enum(['general', 'coding', 'analysis', 'creative', 'educational', 'business', 'custom']),
	is_active: z.boolean().default(true),
	tags: z.array(z.string()).optional(),
	parameters: z.string().optional(),
	metadata: z.string().optional()
});

type PromptFormData = z.infer<typeof promptSchema>;

interface PromptFormDialogProps {
	open: boolean;
	onClose: () => void;
	prompt?: Prompt | null;
}

export default function PromptFormDialog({ open, onClose, prompt }: PromptFormDialogProps) {
	const isEdit = !!prompt;
	const createMutation = useCreatePrompt();
	const updateMutation = useUpdatePrompt();
	
	const form = useForm<PromptFormData>({
		resolver: zodResolver(promptSchema),
		defaultValues: {
			name: '',
			description: '',
			prompt_template: '',
			category: 'general',
			is_active: true,
			tags: [],
			parameters: '',
			metadata: ''
		}
	});
	
	useEffect(() => {
		if (prompt) {
			form.reset({
				name: prompt.name,
				description: prompt.description || '',
				prompt_template: prompt.prompt_template,
				category: prompt.category as any,
				is_active: prompt.is_active,
				tags: prompt.tags || [],
				parameters: prompt.parameters ? JSON.stringify(prompt.parameters, null, 2) : '',
				metadata: prompt.metadata ? JSON.stringify(prompt.metadata, null, 2) : ''
			});
		} else {
			form.reset({
				name: '',
				description: '',
				prompt_template: '',
				category: 'general',
				is_active: true,
				tags: [],
				parameters: '',
				metadata: ''
			});
		}
	}, [prompt, form]);
	
	const handleSubmit = async (data: PromptFormData) => {
		try {
			const parameters = data.parameters ? JSON.parse(data.parameters) : undefined;
			const metadata = data.metadata ? JSON.parse(data.metadata) : undefined;
			
			const payload = {
				...data,
				parameters,
				metadata,
				tags: data.tags?.filter(tag => tag.trim() !== '')
			};
			
			if (isEdit && prompt) {
				await updateMutation.mutateAsync({
					id: prompt.id,
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
	
	return (
		<Dialog open={open} onOpenChange={onClose}>
			<DialogContent className="sm:max-w-[700px] max-h-[90vh] overflow-y-auto">
				<DialogHeader>
					<DialogTitle>{isEdit ? 'Edit Prompt' : 'Create Prompt'}</DialogTitle>
					<DialogDescription>
						{isEdit ? 'Update the prompt template' : 'Add a new prompt template'}
					</DialogDescription>
				</DialogHeader>
				
				<Form {...form}>
					<form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-4">
						<FormField
							control={form.control}
							name="name"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Name</FormLabel>
									<FormControl>
										<Input placeholder="Code Review Assistant" {...field} />
									</FormControl>
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
											placeholder="Describe what this prompt does..."
											{...field}
											rows={2}
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
											<SelectItem value="coding">Coding</SelectItem>
											<SelectItem value="analysis">Analysis</SelectItem>
											<SelectItem value="creative">Creative</SelectItem>
											<SelectItem value="educational">Educational</SelectItem>
											<SelectItem value="business">Business</SelectItem>
											<SelectItem value="custom">Custom</SelectItem>
										</SelectContent>
									</Select>
									<FormMessage />
								</FormItem>
							)}
						/>
						
						<FormField
							control={form.control}
							name="prompt_template"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Prompt Template</FormLabel>
									<FormControl>
										<Textarea
											placeholder="Enter your prompt template here. Use {{variable}} for parameters..."
											{...field}
											rows={6}
											className="font-mono text-sm"
										/>
									</FormControl>
									<FormDescription>
										Use {'{{variable}}'} syntax for template variables
									</FormDescription>
									<FormMessage />
								</FormItem>
							)}
						/>
						
						<FormField
							control={form.control}
							name="parameters"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Parameters (JSON)</FormLabel>
									<FormControl>
										<Textarea
											placeholder='[{"name": "variable", "type": "string", "description": "Description"}]'
											{...field}
											rows={3}
											className="font-mono text-sm"
										/>
									</FormControl>
									<FormDescription>
										Define parameters as a JSON array
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
						
						<FormField
							control={form.control}
							name="is_active"
							render={({ field }) => (
								<FormItem className="flex items-center justify-between rounded-lg border p-3">
									<div className="space-y-0.5">
										<FormLabel>Active</FormLabel>
										<FormDescription>
											Enable or disable this prompt
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