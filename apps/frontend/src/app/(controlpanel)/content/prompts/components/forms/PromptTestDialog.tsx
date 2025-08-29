'use client';

import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle
} from '@/components/ui/dialog';
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Button } from '@/components/ui/button';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Loader2, Copy, CheckCircle } from 'lucide-react';
import { usePrompt, usePromptTest } from '../../api/hooks/usePrompts';
import { PromptParameter } from '@/lib/types';
import { toast } from 'sonner';

interface PromptTestDialogProps {
	promptId: string | null;
	onClose: () => void;
}

export default function PromptTestDialog({ promptId, onClose }: PromptTestDialogProps) {
	const [result, setResult] = useState<string>('');
	const [copied, setCopied] = useState(false);

	const { data: prompt } = usePrompt(promptId);
	const testMutation = usePromptTest();

	const form = useForm({
		defaultValues: {
			parameters: {} as Record<string, unknown>
		}
	});

	useEffect(() => {
		if (prompt?.parameters) {
			const defaultParams: Record<string, string> = {};
			(prompt.parameters as PromptParameter[])?.forEach((param) => {
				defaultParams[param.name] = '';
			});
			form.reset({ parameters: defaultParams });
		}
	}, [prompt, form]);

	const handleTest = (data: { parameters: Record<string, unknown> }) => {
		if (!promptId) return;

		testMutation.mutate({
			id: promptId,
			data: { parameters: data.parameters }
		});
	};

	const handleCopy = () => {
		navigator.clipboard.writeText(result);
		setCopied(true);
		toast.success('Copied to clipboard');
		setTimeout(() => setCopied(false), 2000);
	};

	if (!promptId || !prompt) return null;

	const parameters = (prompt.parameters as PromptParameter[]) || [];

	return (
		<Dialog
			open={!!promptId}
			onOpenChange={onClose}
		>
			<DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-[700px]">
				<DialogHeader>
					<DialogTitle>Test Prompt: {prompt.name}</DialogTitle>
					<DialogDescription>Fill in the parameters to test this prompt template</DialogDescription>
				</DialogHeader>

				<div className="space-y-4">
					{/* Template Preview */}
					<div className="bg-muted rounded-lg p-4">
						<h4 className="mb-2 text-sm font-medium">Template:</h4>
						<pre className="whitespace-pre-wrap text-sm">{prompt.prompt_template}</pre>
					</div>

					{/* Parameters Form */}
					{parameters.length > 0 && (
						<Form {...form}>
							<form
								onSubmit={form.handleSubmit(handleTest)}
								className="space-y-4"
							>
								<div className="space-y-3">
									{parameters.map((param: PromptParameter) => (
										<FormField
											key={param.name}
											control={form.control}
											name={`parameters.${param.name}`}
											render={({ field }) => (
												<FormItem>
													<FormLabel>{param.name}</FormLabel>
													{param.type === 'text' ? (
														<FormControl>
															<Textarea
																placeholder={
																	param.description || `Enter ${param.name}...`
																}
																{...field}
																rows={3}
															/>
														</FormControl>
													) : (
														<FormControl>
															<Input
																placeholder={
																	param.description || `Enter ${param.name}...`
																}
																{...field}
															/>
														</FormControl>
													)}
													{param.description && (
														<FormDescription>{param.description}</FormDescription>
													)}
													<FormMessage />
												</FormItem>
											)}
										/>
									))}
								</div>

								<Button
									type="submit"
									className="w-full"
									disabled={testMutation.isPending}
								>
									{testMutation.isPending ? (
										<>
											<Loader2 className="mr-2 h-4 w-4 animate-spin" />
											Testing...
										</>
									) : (
										'Test Prompt'
									)}
								</Button>
							</form>
						</Form>
					)}

					{/* No parameters */}
					{parameters.length === 0 && (
						<div className="space-y-4">
							<Alert>
								<AlertDescription>
									This prompt has no parameters. Click test to see the rendered output.
								</AlertDescription>
							</Alert>
							<Button
								onClick={() => handleTest({ parameters: {} })}
								className="w-full"
								disabled={testMutation.isPending}
							>
								{testMutation.isPending ? (
									<>
										<Loader2 className="mr-2 h-4 w-4 animate-spin" />
										Testing...
									</>
								) : (
									'Test Prompt'
								)}
							</Button>
						</div>
					)}

					{/* Result */}
					{result && (
						<div className="space-y-2">
							<div className="flex items-center justify-between">
								<h4 className="text-sm font-medium">Result:</h4>
								<Button
									variant="ghost"
									size="sm"
									onClick={handleCopy}
								>
									{copied ? (
										<CheckCircle className="h-4 w-4 text-green-500" />
									) : (
										<Copy className="h-4 w-4" />
									)}
								</Button>
							</div>
							<div className="bg-background rounded-lg border p-4">
								<pre className="whitespace-pre-wrap text-sm">{result}</pre>
							</div>
						</div>
					)}
				</div>

				<DialogFooter>
					<Button
						variant="outline"
						onClick={onClose}
					>
						Close
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}
