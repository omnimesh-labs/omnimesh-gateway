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
import { X, CheckCircle, AlertCircle, Info, Lightbulb } from 'lucide-react';
import { useCreateResource, useUpdateResource } from '../../api/hooks/useResources';
import { Resource } from '@/lib/api';
import {
  resourceSchemaWithTypeValidation,
  validateUriForType,
  suggestMimeType,
  validateAccessPermissions,
  type ResourceFormData,
  type CreateResourceData,
  type UpdateResourceData
} from '@/lib/validation/resource';
import { enums } from '@/lib/validation/common';

interface ResourceFormDialogProps {
	open: boolean;
	onClose: () => void;
	resource?: Resource | null;
}

export default function ResourceFormDialog({ open, onClose, resource }: ResourceFormDialogProps) {
	const isEdit = !!resource;
	const createMutation = useCreateResource();
	const updateMutation = useUpdateResource();

	// Validation state
	const [uriValidation, setUriValidation] = useState<{ valid: boolean; error?: string }>({ valid: true });
	const [suggestedMimeType, setSuggestedMimeType] = useState<string>('');

	const form = useForm<ResourceFormData>({
		resolver: zodResolver(resourceSchemaWithTypeValidation),
		defaultValues: {
			name: '',
			description: '',
			resource_type: 'file',
			uri: '',
			mime_type: '',
			is_active: true,
			tags: [],
			metadata: ''
		}
	});

	useEffect(() => {
		if (resource) {
			form.reset({
				name: resource.name,
				description: resource.description || '',
				resource_type: resource.resource_type as
					| 'file'
					| 'folder'
					| 'database'
					| 'api'
					| 'url'
					| 'document'
					| 'image'
					| 'video'
					| 'audio'
					| 'other',
				uri: resource.uri,
				mime_type: resource.mime_type || '',
				size_bytes: resource.size_bytes,
				is_active: resource.is_active,
				tags: resource.tags || [],
				metadata: resource.metadata ? JSON.stringify(resource.metadata, null, 2) : ''
			});
		} else {
			form.reset({
				name: '',
				description: '',
				resource_type: 'file',
				uri: '',
				mime_type: '',
				is_active: true,
				tags: [],
				metadata: ''
			});
		}
	}, [resource, form]);

	// Handle URI validation based on resource type
	const handleUriChange = (uri: string, resourceType: typeof enums.resourceType[number]) => {
		const validation = validateUriForType(uri, resourceType);
		setUriValidation(validation);

		// Suggest MIME type based on URI
		if (validation.valid) {
			const suggested = suggestMimeType(uri, resourceType);
			if (suggested && !form.getValues('mime_type')) {
				setSuggestedMimeType(suggested);
			}
		}
	};

	// Handle resource type change
	const handleResourceTypeChange = (resourceType: typeof enums.resourceType[number]) => {
		const currentUri = form.getValues('uri');
		if (currentUri) {
			handleUriChange(currentUri, resourceType);
		}

		// Clear suggested MIME type when resource type changes
		setSuggestedMimeType('');
	};

	// Apply suggested MIME type
	const applySuggestedMimeType = () => {
		if (suggestedMimeType) {
			form.setValue('mime_type', suggestedMimeType);
			setSuggestedMimeType('');
		}
	};

	const handleSubmit = async (data: ResourceFormData) => {
		try {
			const metadata = data.metadata ? JSON.parse(data.metadata) : undefined;

			const payload = {
				...data,
				metadata,
				tags: data.tags?.filter((tag) => tag.trim() !== '')
			};

			if (isEdit && resource) {
				await updateMutation.mutateAsync({
					id: resource.id,
					data: payload
				});
			} else {
				await createMutation.mutateAsync(payload);
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
			<DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-[600px]">
				<DialogHeader>
					<DialogTitle>{isEdit ? 'Edit Resource' : 'Create Resource'}</DialogTitle>
					<DialogDescription>
						{isEdit ? 'Update the resource details' : 'Add a new resource to your gateway'}
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
											placeholder="My Resource"
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
											placeholder="Describe this resource..."
											{...field}
											rows={3}
										/>
									</FormControl>
									<FormMessage />
								</FormItem>
							)}
						/>

						<div className="grid grid-cols-2 gap-4">
							<FormField
								control={form.control}
								name="resource_type"
								render={({ field }) => (
									<FormItem>
										<FormLabel>Resource Type</FormLabel>
										<Select
											onValueChange={(value) => {
												field.onChange(value);
												handleResourceTypeChange(value as typeof enums.resourceType[number]);
											}}
											value={field.value}
										>
											<FormControl>
												<SelectTrigger>
													<SelectValue placeholder="Select type" />
												</SelectTrigger>
											</FormControl>
											<SelectContent>
												{enums.resourceType.map((type) => (
													<SelectItem key={type} value={type}>
														{type.charAt(0).toUpperCase() + type.slice(1)}
													</SelectItem>
												))}
											</SelectContent>
										</Select>
										<FormDescription>
											Resource type determines validation rules for the URI
										</FormDescription>
										<FormMessage />
									</FormItem>
								)}
							/>

							<FormField
								control={form.control}
								name="mime_type"
								render={({ field }) => (
									<FormItem>
										<div className="flex items-center justify-between">
											<FormLabel>MIME Type</FormLabel>
											{suggestedMimeType && (
												<Button
													type="button"
													variant="ghost"
													size="sm"
													onClick={applySuggestedMimeType}
													className="text-blue-600 hover:text-blue-700"
												>
													<Lightbulb className="h-4 w-4 mr-1" />
													Use {suggestedMimeType}
												</Button>
											)}
										</div>
										<FormControl>
											<Input
												placeholder="application/json"
												{...field}
											/>
										</FormControl>
										<FormDescription>
											Optional MIME type. Will be suggested based on URI extension.
										</FormDescription>
										<FormMessage />
									</FormItem>
								)}
							/>
						</div>

						<FormField
							control={form.control}
							name="uri"
							render={({ field }) => (
								<FormItem>
									<FormLabel>URI</FormLabel>
									<FormControl>
										<Input
											placeholder="https://example.com/resource or /path/to/file"
											{...field}
											onChange={(e) => {
												field.onChange(e);
												const resourceType = form.getValues('resource_type');
												handleUriChange(e.target.value, resourceType);
											}}
										/>
									</FormControl>
									<FormDescription>
										The location of the resource. Format depends on resource type.
									</FormDescription>
									{!uriValidation.valid && (
										<Alert className="mt-2" variant="destructive">
											<AlertCircle className="h-4 w-4" />
											<AlertDescription>{uriValidation.error}</AlertDescription>
										</Alert>
									)}
									{uriValidation.valid && field.value && (
										<div className="flex items-center gap-1 text-green-600 text-sm mt-1">
											<CheckCircle className="h-4 w-4" />
											<span>Valid URI format</span>
										</div>
									)}
									<FormMessage />
								</FormItem>
							)}
						/>

						<FormField
							control={form.control}
							name="size_bytes"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Size (bytes)</FormLabel>
									<FormControl>
										<Input
											type="number"
											placeholder="1024"
											{...field}
											onChange={(e) =>
												field.onChange(e.target.value ? parseInt(e.target.value) : undefined)
											}
										/>
									</FormControl>
									<FormDescription>Optional file size in bytes</FormDescription>
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
											rows={4}
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
										<FormDescription>Enable or disable this resource</FormDescription>
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
