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
import { Alert, AlertDescription } from '@/components/ui/alert';
import { X, CheckCircle, AlertCircle, Info, RefreshCw } from 'lucide-react';
import { useCreatePrompt, useUpdatePrompt } from '../../api/hooks/usePrompts';
import { Prompt } from '@/lib/types';
import {
  promptSchemaWithTemplateValidation,
  validatePromptParameters,
  validatePromptTemplate,
  extractTemplateParameters,
  suggestPromptCategory,
  generateParameterTemplate,
  type PromptFormData,
  type CreatePromptData,
  type UpdatePromptData
} from '@/lib/validation/prompt';
import { enums } from '@/lib/validation/common';

interface PromptFormDialogProps {
	open: boolean;
	onClose: () => void;
	prompt?: Prompt | null;
}

export default function PromptFormDialog({ open, onClose, prompt }: PromptFormDialogProps) {
	const isEdit = !!prompt;
	const createMutation = useCreatePrompt();
	const updateMutation = useUpdatePrompt();

	// Validation state
	const [templateValidation, setTemplateValidation] = useState<{
		valid: boolean;
		error?: string;
		warnings?: string[];
		parameters?: string[];
	}>({ valid: true });
	const [parametersValidation, setParametersValidation] = useState<{
		valid: boolean;
		error?: string;
		suggestions?: string[];
		parameters?: Array<{ name: string; type?: string; description?: string; required?: boolean }>;
	}>({ valid: true });

	const form = useForm<PromptFormData>({
		resolver: zodResolver(promptSchemaWithTemplateValidation),
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
				category: prompt.category as 'general' | 'system' | 'knowledge_base' | 'custom',
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

	// Handle template validation
	const handleTemplateChange = (template: string) => {
		const validation = validatePromptTemplate(template);
		setTemplateValidation(validation);

		// Auto-generate parameters if template has parameters but no parameters defined
		if (validation.valid && validation.parameters && validation.parameters.length > 0) {
			const currentParameters = form.getValues('parameters');
			if (!currentParameters || currentParameters.trim() === '') {
				const parameterTemplate = generateParameterTemplate(validation.parameters);
				form.setValue('parameters', parameterTemplate);
				setParametersValidation({ valid: true });
			}
		}
	};

	// Handle parameters validation
	const handleParametersChange = (parametersString: string) => {
		if (parametersString.trim() === '') {
			setParametersValidation({ valid: true, parameters: [] });
			return;
		}

		const validation = validatePromptParameters(parametersString);
		setParametersValidation(validation);
	};

	// Handle category suggestion based on template content
	const handleTemplateBlur = (template: string) => {
		if (template && !isEdit) {
			const suggestedCategory = suggestPromptCategory(template, form.getValues('name'));
			if (suggestedCategory !== form.getValues('category')) {
				form.setValue('category', suggestedCategory);
			}
		}
	};

	// Generate parameters from template
	const handleGenerateParameters = () => {
		const template = form.getValues('prompt_template');
		if (template) {
			const templateParams = extractTemplateParameters(template);
			if (templateParams.length > 0) {
				const parameterTemplate = generateParameterTemplate(templateParams);
				form.setValue('parameters', parameterTemplate);
				setParametersValidation({ valid: true });
			}
		}
	};

	const handleSubmit = async (data: PromptFormData) => {
		try {
			// Validate template before submission
			if (!templateValidation.valid) {
				form.setError('prompt_template', { message: templateValidation.error || 'Invalid template' });
				return;
			}

			// Validate parameters before submission
			if (data.parameters && !parametersValidation.valid) {
				form.setError('parameters', { message: parametersValidation.error || 'Invalid parameters' });
				return;
			}

			const parameters = data.parameters ? JSON.parse(data.parameters) : undefined;
			const metadata = data.metadata ? JSON.parse(data.metadata) : undefined;

			const payload = {
				...data,
				parameters,
				metadata,
				tags: data.tags?.filter((tag) => tag.trim() !== '')
			};

			if (isEdit && prompt) {
				await updateMutation.mutateAsync({
					id: prompt.id,
					data: payload as UpdatePromptData
				});
			} else {
				await createMutation.mutateAsync(payload as CreatePromptData);
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

	return (
		<Dialog
			open={open}
			onOpenChange={onClose}
		>
			<DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-[700px]">
				<DialogHeader>
					<DialogTitle>{isEdit ? 'Edit Prompt' : 'Create Prompt'}</DialogTitle>
					<DialogDescription>
						{isEdit ? 'Update the prompt template' : 'Add a new prompt template'}
					</DialogDescription>
				</DialogHeader>

				<Form {...form}>
					<form
						onSubmit={form.handleSubmit(handleSubmit)}
						className="space-y-4"
					>
						<FormField
							control={form.control}
							name="name"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Name</FormLabel>
									<FormControl>
										<Input
											placeholder="Code Review Assistant"
											{...field}
										/>
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
											{enums.promptCategory.map((category) => (
												<SelectItem key={category} value={category}>
													{category.charAt(0).toUpperCase() + category.slice(1)}
												</SelectItem>
											))}
										</SelectContent>
									</Select>
									<FormDescription>
										Category helps organize and discover prompts
									</FormDescription>
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
											onChange={(e) => {
												field.onChange(e);
												handleTemplateChange(e.target.value);
											}}
											onBlur={(e) => {
												field.onBlur();
												handleTemplateBlur(e.target.value);
											}}
										/>
									</FormControl>
									<FormDescription>
										Use {'{{variable}}'} syntax for template variables. Parameters will be auto-detected.
									</FormDescription>
									{templateValidation.warnings && templateValidation.warnings.length > 0 && (
										<Alert className="mt-2">
											<Info className="h-4 w-4" />
											<AlertDescription>
												<ul className="list-disc list-inside text-sm">
													{templateValidation.warnings.map((warning, index) => (
														<li key={index}>{warning}</li>
													))}
												</ul>
											</AlertDescription>
										</Alert>
									)}
									{templateValidation.valid && templateValidation.parameters && templateValidation.parameters.length > 0 && (
										<div className="flex items-center gap-2 text-sm text-blue-600 mt-1">
											<Info className="h-4 w-4" />
											<span>
												Template parameters detected: {templateValidation.parameters.join(', ')}
											</span>
										</div>
									)}
									{!templateValidation.valid && (
										<Alert className="mt-2" variant="destructive">
											<AlertCircle className="h-4 w-4" />
											<AlertDescription>{templateValidation.error}</AlertDescription>
										</Alert>
									)}
									<FormMessage />
								</FormItem>
							)}
						/>

						<FormField
							control={form.control}
							name="parameters"
							render={({ field }) => (
								<FormItem>
									<div className="flex items-center justify-between">
										<FormLabel>Parameters (JSON)</FormLabel>
										<Button
											type="button"
											variant="outline"
											size="sm"
											onClick={handleGenerateParameters}
											disabled={!form.getValues('prompt_template')}
										>
											<RefreshCw className="h-4 w-4 mr-1" />
											Generate from Template
										</Button>
									</div>
									<FormControl>
										<Textarea
											placeholder='[{"name": "variable", "type": "string", "description": "Description"}]'
											{...field}
											rows={3}
											className="font-mono text-sm"
											onChange={(e) => {
												field.onChange(e);
												handleParametersChange(e.target.value);
											}}
										/>
									</FormControl>
									<FormDescription>
										Define parameters as a JSON array. Leave empty if no parameters needed.
									</FormDescription>
									{!parametersValidation.valid && (
										<Alert className="mt-2" variant="destructive">
											<AlertCircle className="h-4 w-4" />
											<AlertDescription>
												{parametersValidation.error}
												{parametersValidation.suggestions && (
													<ul className="mt-1 list-disc list-inside text-sm">
														{parametersValidation.suggestions.map((suggestion, index) => (
															<li key={index}>{suggestion}</li>
														))}
													</ul>
												)}
											</AlertDescription>
										</Alert>
									)}
									{parametersValidation.valid && parametersValidation.parameters && parametersValidation.parameters.length > 0 && (
										<div className="flex items-center gap-1 text-green-600 text-sm mt-1">
											<CheckCircle className="h-4 w-4" />
											<span>
												{parametersValidation.parameters.length} parameter(s) defined
											</span>
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

						<FormField
							control={form.control}
							name="is_active"
							render={({ field }) => (
								<FormItem className="flex items-center justify-between rounded-lg border p-3">
									<div className="space-y-0.5">
										<FormLabel>Active</FormLabel>
										<FormDescription>Enable or disable this prompt</FormDescription>
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
