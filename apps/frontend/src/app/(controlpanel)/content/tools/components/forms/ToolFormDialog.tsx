'use client';

import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle
} from '@/components/ui/dialog';
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Button } from '@/components/ui/button';
import { Switch } from '@/components/ui/switch';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { X, CheckCircle, AlertCircle, Info } from 'lucide-react';
import { useCreateTool, useUpdateTool } from '../../api/hooks/useTools';
import { Tool } from '@/lib/types';
import {
  toolSchemaWithImplementationValidation,
  validateToolSchema,
  validateToolExamples,
  suggestToolCategory,
  generateToolSchemaTemplate,
  type ToolFormData,
  type CreateToolData,
  type UpdateToolData
} from '@/lib/validation/tool';
import { enums } from '@/lib/validation/common';

interface ToolFormDialogProps {
	open: boolean;
	onClose: () => void;
	tool?: Tool | null;
}

export default function ToolFormDialog({ open, onClose, tool }: ToolFormDialogProps) {
	const isEdit = !!tool;
	const createMutation = useCreateTool();
	const updateMutation = useUpdateTool();

	// Validation state
	const [schemaValidation, setSchemaValidation] = useState<{ valid: boolean; error?: string; suggestions?: string[] }>({ valid: true });
	const [examplesValidation, setExamplesValidation] = useState<{ valid: boolean; error?: string; suggestions?: string[] }>({ valid: true });

	const form = useForm<ToolFormData>({
		resolver: zodResolver(toolSchemaWithImplementationValidation),
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
				category: tool.category as
					| 'general'
					| 'data'
					| 'communication'
					| 'automation'
					| 'ai'
					| 'security'
					| 'devops'
					| 'custom',
				implementation_type: tool.implementation_type as
					| 'internal'
					| 'external'
					| 'webhook'
					| 'script'
					| 'plugin',
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

	// Handle schema validation
	const handleSchemaChange = (schemaString: string) => {
		if (schemaString.trim() === '') {
			setSchemaValidation({ valid: true });
			return;
		}

		const validation = validateToolSchema(schemaString);
		setSchemaValidation(validation);
	};

	// Handle examples validation
	const handleExamplesChange = (examplesString: string) => {
		if (examplesString.trim() === '') {
			setExamplesValidation({ valid: true });
			return;
		}

		const validation = validateToolExamples(examplesString);
		setExamplesValidation(validation);
	};

	// Handle category suggestion based on name/description
	const handleNameChange = (name: string) => {
		if (name && !isEdit) {
			const suggestedCategory = suggestToolCategory(name, form.getValues('description'));
			if (suggestedCategory !== form.getValues('category')) {
				// Only suggest if different from current
				form.setValue('category', suggestedCategory);
			}
		}
	};

	// Generate schema template
	const handleGenerateSchema = () => {
		const category = form.getValues('category');
		const template = generateToolSchemaTemplate(category);
		form.setValue('schema', template);
		setSchemaValidation({ valid: true });
	};

	const handleSubmit = async (data: ToolFormData) => {
		try {
			// Validate JSON fields before submission
			if (data.schema && !schemaValidation.valid) {
				form.setError('schema', { message: schemaValidation.error || 'Invalid schema' });
				return;
			}

			if (data.examples && !examplesValidation.valid) {
				form.setError('examples', { message: examplesValidation.error || 'Invalid examples' });
				return;
			}

			const schema = data.schema ? JSON.parse(data.schema) : undefined;
			const examples = data.examples ? JSON.parse(data.examples) : undefined;
			const metadata = data.metadata ? JSON.parse(data.metadata) : undefined;

			const payload = {
				...data,
				schema,
				examples,
				metadata,
				tags: data.tags?.filter((tag) => tag.trim() !== '')
			};

			if (isEdit && tool) {
				await updateMutation.mutateAsync({
					id: tool.id,
					data: payload as UpdateToolData
				});
			} else {
				await createMutation.mutateAsync(payload as CreateToolData);
			}

			onClose();
		} catch (_error) {
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
		form.setValue(
			'tags',
			currentTags.filter((tag) => tag !== tagToRemove)
		);
	};

	const implementationType = form.watch('implementation_type');

	return (
		<Dialog
			open={open}
			onOpenChange={onClose}
		>
			<DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-[800px]">
				<DialogHeader>
					<DialogTitle>{isEdit ? 'Edit Tool' : 'Create Tool'}</DialogTitle>
					<DialogDescription>
						{isEdit ? 'Update the tool configuration' : 'Add a new tool to your gateway'}
					</DialogDescription>
				</DialogHeader>

				<Form {...form}>
					<form
						onSubmit={form.handleSubmit(handleSubmit)}
						className="space-y-4"
					>
						<Tabs
							defaultValue="basic"
							className="w-full"
						>
							<TabsList className="grid w-full grid-cols-3">
								<TabsTrigger value="basic">Basic Info</TabsTrigger>
								<TabsTrigger value="implementation">Implementation</TabsTrigger>
								<TabsTrigger value="advanced">Advanced</TabsTrigger>
							</TabsList>

							<TabsContent
								value="basic"
								className="mt-4 space-y-4"
							>
								<FormField
									control={form.control}
									name="name"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Name</FormLabel>
											<FormControl>
												<Input
													placeholder="My Awesome Tool"
													{...field}
													onChange={(e) => {
														field.onChange(e);
														handleNameChange(e.target.value);
													}}
												/>
											</FormControl>
											<FormDescription>
												Tool name will be used for identification and documentation
											</FormDescription>
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
												<Input
													placeholder="my_awesome_tool"
													{...field}
												/>
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
											<Select
												onValueChange={field.onChange}
												value={field.value}
											>
												<FormControl>
													<SelectTrigger>
														<SelectValue placeholder="Select category" />
													</SelectTrigger>
												</FormControl>
												<SelectContent>
													{enums.toolCategory.map((category) => (
														<SelectItem key={category} value={category}>
															{category.charAt(0).toUpperCase() + category.slice(1)}
														</SelectItem>
													))}
												</SelectContent>
											</Select>
											<FormDescription>
												Category helps organize and discover tools
											</FormDescription>
											<FormMessage />
										</FormItem>
									)}
								/>
							</TabsContent>

							<TabsContent
								value="implementation"
								className="mt-4 space-y-4"
							>
								<FormField
									control={form.control}
									name="implementation_type"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Implementation Type</FormLabel>
											<Select
												onValueChange={field.onChange}
												value={field.value}
											>
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
													<Input
														placeholder="https://api.example.com/endpoint"
														{...field}
													/>
												</FormControl>
												<FormDescription>The URL to call for this tool</FormDescription>
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
											<div className="flex items-center justify-between">
												<FormLabel>Schema (JSON)</FormLabel>
												<Button
													type="button"
													variant="outline"
													size="sm"
													onClick={handleGenerateSchema}
													disabled={!form.getValues('category')}
												>
													Generate Template
												</Button>
											</div>
											<FormControl>
												<Textarea
													placeholder='{"type": "object", "properties": {...}}'
													{...field}
													rows={6}
													className="font-mono text-sm"
													onChange={(e) => {
														field.onChange(e);
														handleSchemaChange(e.target.value);
													}}
												/>
											</FormControl>
											<FormDescription>
												JSON Schema for tool parameters. Leave empty if no parameters required.
											</FormDescription>
											{!schemaValidation.valid && (
												<Alert className="mt-2">
													<AlertCircle className="h-4 w-4" />
													<AlertDescription>
														{schemaValidation.error}
														{schemaValidation.suggestions && (
															<ul className="mt-1 list-disc list-inside text-sm">
																{schemaValidation.suggestions.map((suggestion, index) => (
																	<li key={index}>{suggestion}</li>
																))}
															</ul>
														)}
													</AlertDescription>
												</Alert>
											)}
											{schemaValidation.valid && field.value && (
												<div className="flex items-center gap-1 text-green-600 text-sm mt-1">
													<CheckCircle className="h-4 w-4" />
													<span>Valid JSON Schema</span>
												</div>
											)}
											<FormMessage />
										</FormItem>
									)}
								/>
							</TabsContent>

							<TabsContent
								value="advanced"
								className="mt-4 space-y-4"
							>
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
											<FormDescription>Markdown supported</FormDescription>
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
													onChange={(e) => {
														field.onChange(e);
														handleExamplesChange(e.target.value);
													}}
												/>
											</FormControl>
											<FormDescription>
												Usage examples as JSON array. Each example should have "input" and "output" properties.
											</FormDescription>
											{!examplesValidation.valid && (
												<Alert className="mt-2">
													<AlertCircle className="h-4 w-4" />
													<AlertDescription>
														{examplesValidation.error}
														{examplesValidation.suggestions && (
															<ul className="mt-1 list-disc list-inside text-sm">
																{examplesValidation.suggestions.map((suggestion, index) => (
																	<li key={index}>{suggestion}</li>
																))}
															</ul>
														)}
													</AlertDescription>
												</Alert>
											)}
											{examplesValidation.valid && field.value && (
												<div className="flex items-center gap-1 text-green-600 text-sm mt-1">
													<CheckCircle className="h-4 w-4" />
													<span>Valid examples format</span>
												</div>
											)}
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
															<Badge
																key={tag}
																variant="secondary"
																className="gap-1"
															>
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
											<FormDescription>Additional metadata in JSON format</FormDescription>
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
													<FormDescription>Enable or disable this tool</FormDescription>
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
							<Button
								type="button"
								variant="outline"
								onClick={onClose}
							>
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
