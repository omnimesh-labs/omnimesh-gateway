'use client';

import { useEffect, useState } from 'react';
import { useForm, Controller } from 'react-hook-form';
import {
	Dialog,
	DialogContent,
	DialogActions,
	DialogTitle,
	Button,
	TextField,
	Tab,
	Box,
	Typography,
	Alert,
	IconButton,
	CircularProgress
} from '@mui/material';
import { TabContext, TabList, TabPanel } from '@mui/lab';
import { Copy, CheckCircle } from 'lucide-react';
import { useTool, useExecuteTool } from '../../api/hooks/useTools';
import { enqueueSnackbar } from 'notistack';
import { JSONSchema } from '@/lib/types';

interface ToolExecuteDialogProps {
	toolId: string | null;
	onClose: () => void;
}

export default function ToolExecuteDialog({ toolId, onClose }: ToolExecuteDialogProps) {
	const [result, setResult] = useState<unknown>(null);
	const [error, setError] = useState<string>('');
	const [copied, setCopied] = useState(false);
	const [tabValue, setTabValue] = useState('execute');

	const { data: tool } = useTool(toolId);
	const executeMutation = useExecuteTool();

	const { control, handleSubmit, reset } = useForm({
		defaultValues: {
			parameters: ''
		}
	});

	useEffect(() => {
		if (tool?.schema) {
			// Generate example parameters based on schema
			const exampleParams = generateExampleFromSchema(tool.schema);
			reset({
				parameters: JSON.stringify(exampleParams, null, 2)
			});
		} else {
			reset({
				parameters: '{}'
			});
		}

		setResult(null);
		setError('');
	}, [tool, reset]);

	const generateExampleFromSchema = (schema: JSONSchema): Record<string, unknown> => {
		if (!schema || !schema.properties) return {};

		const example: Record<string, unknown> = {};
		Object.entries(schema.properties).forEach(([key, prop]) => {
			switch (prop.type) {
				case 'string':
					example[key] = prop.example || prop.default || '';
					break;
				case 'number':
				case 'integer':
					example[key] = prop.example || prop.default || 0;
					break;
				case 'boolean':
					example[key] = prop.example || prop.default || false;
					break;
				case 'array':
					example[key] = prop.example || prop.default || [];
					break;
				case 'object':
					example[key] = prop.example || prop.default || {};
					break;
				default:
					example[key] = null;
			}
		});
		return example;
	};

	const handleExecute = (data: { parameters: string }) => {
		if (!toolId) return;

		setError('');
		setResult(null);

		try {
			const parameters = JSON.parse(data.parameters);
			executeMutation.mutate({
				id: toolId,
				data: { parameters }
			});
		} catch (error) {
			const message = error instanceof Error ? error.message : 'Failed to parse parameters';
			setError(message);
		}
	};

	const handleCopy = (text: string) => {
		navigator.clipboard.writeText(text);
		setCopied(true);
		enqueueSnackbar('Copied to clipboard', { variant: 'success' });
		setTimeout(() => setCopied(false), 2000);
	};

	if (!toolId || !tool) return null;

	return (
		<Dialog
			open={!!toolId}
			onClose={onClose}
			maxWidth="lg"
			fullWidth
		>
			<DialogTitle>
				Execute Tool: {tool.name}
				<Typography
					variant="body2"
					color="text.secondary"
				>
					Test the tool with custom parameters
				</Typography>
			</DialogTitle>
			<DialogContent sx={{ height: '80vh', display: 'flex', flexDirection: 'column' }}>
				<TabContext value={tabValue}>
					<Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
						<TabList onChange={(_, newValue) => setTabValue(newValue)}>
							<Tab
								label="Execute"
								value="execute"
							/>
							<Tab
								label="Schema"
								value="schema"
							/>
							<Tab
								label="Examples"
								value="examples"
							/>
						</TabList>
					</Box>

					<TabPanel
						value="execute"
						sx={{ flex: 1, display: 'flex', flexDirection: 'column', p: 2 }}
					>
						<Box sx={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 2, flex: 1 }}>
							{/* Input */}
							<Box sx={{ display: 'flex', flexDirection: 'column' }}>
								<Typography
									variant="subtitle2"
									sx={{ mb: 1 }}
								>
									Parameters
								</Typography>
								<form
									onSubmit={handleSubmit(handleExecute)}
									style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 16 }}
								>
									<Controller
										name="parameters"
										control={control}
										render={({ field }) => (
											<TextField
												{...field}
												placeholder='{"key": "value"}'
												multiline
												rows={10}
												variant="outlined"
												fullWidth
												sx={{
													flex: 1,
													'& .MuiInputBase-input': {
														fontFamily: 'monospace',
														fontSize: '0.875rem'
													}
												}}
												helperText="Enter parameters as JSON"
											/>
										)}
									/>

									<Button
										type="submit"
										variant="contained"
										fullWidth
										disabled={executeMutation.isPending}
										startIcon={executeMutation.isPending ? <CircularProgress size={16} /> : null}
									>
										{executeMutation.isPending ? 'Executing...' : 'Execute Tool'}
									</Button>
								</form>
							</Box>

							{/* Output */}
							<Box sx={{ display: 'flex', flexDirection: 'column' }}>
								<Box
									sx={{
										display: 'flex',
										alignItems: 'center',
										justifyContent: 'space-between',
										mb: 1
									}}
								>
									<Typography variant="subtitle2">Result</Typography>
									{result && (
										<IconButton
											size="small"
											onClick={() => handleCopy(JSON.stringify(result, null, 2))}
										>
											{copied ? <CheckCircle style={{ color: 'green' }} /> : <Copy />}
										</IconButton>
									)}
								</Box>

								<Box
									sx={{
										flex: 1,
										border: 1,
										borderColor: 'divider',
										borderRadius: 1,
										p: 2,
										backgroundColor: 'grey.50',
										overflow: 'auto'
									}}
								>
									{error && (
										<Alert
											severity="error"
											sx={{ mb: 2 }}
										>
											<Typography variant="body2">Error: {error}</Typography>
										</Alert>
									)}

									{result && (
										<Box
											component="pre"
											sx={{
												fontSize: '0.875rem',
												whiteSpace: 'pre-wrap',
												fontFamily: 'monospace'
											}}
										>
											{JSON.stringify(result, null, 2)}
										</Box>
									)}

									{!error && !result && (
										<Typography
											variant="body2"
											color="text.secondary"
										>
											Execute the tool to see results here
										</Typography>
									)}
								</Box>
							</Box>
						</Box>
					</TabPanel>

					<TabPanel
						value="schema"
						sx={{ flex: 1, p: 2 }}
					>
						<Box
							sx={{
								height: '100%',
								border: 1,
								borderColor: 'divider',
								borderRadius: 1,
								p: 2,
								backgroundColor: 'grey.50',
								overflow: 'auto'
							}}
						>
							{tool.schema ? (
								<Box
									component="pre"
									sx={{
										fontSize: '0.875rem',
										whiteSpace: 'pre-wrap',
										fontFamily: 'monospace'
									}}
								>
									{JSON.stringify(tool.schema, null, 2)}
								</Box>
							) : (
								<Typography
									variant="body2"
									color="text.secondary"
								>
									No schema defined for this tool
								</Typography>
							)}
						</Box>
					</TabPanel>

					<TabPanel
						value="examples"
						sx={{ flex: 1, p: 2 }}
					>
						<Box
							sx={{
								height: '100%',
								border: 1,
								borderColor: 'divider',
								borderRadius: 1,
								p: 2,
								backgroundColor: 'grey.50',
								overflow: 'auto'
							}}
						>
							{tool.examples && tool.examples.length > 0 ? (
								<Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
									{(tool.examples as { input?: unknown; output?: unknown }[]).map(
										(example, index) => (
											<Box
												key={index}
												sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}
											>
												<Typography variant="subtitle2">Example {index + 1}</Typography>
												<Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
													{example.input && (
														<Box>
															<Typography
																variant="caption"
																color="text.secondary"
															>
																Input:
															</Typography>
															<Box
																component="pre"
																sx={{
																	fontSize: '0.875rem',
																	backgroundColor: 'background.paper',
																	borderRadius: 1,
																	p: 1,
																	fontFamily: 'monospace'
																}}
															>
																{JSON.stringify(example.input, null, 2)}
															</Box>
														</Box>
													)}
													{example.output && (
														<Box>
															<Typography
																variant="caption"
																color="text.secondary"
															>
																Output:
															</Typography>
															<Box
																component="pre"
																sx={{
																	fontSize: '0.875rem',
																	backgroundColor: 'background.paper',
																	borderRadius: 1,
																	p: 1,
																	fontFamily: 'monospace'
																}}
															>
																{JSON.stringify(example.output, null, 2)}
															</Box>
														</Box>
													)}
												</Box>
											</Box>
										)
									)}
								</Box>
							) : (
								<Typography
									variant="body2"
									color="text.secondary"
								>
									No examples defined for this tool
								</Typography>
							)}
						</Box>
					</TabPanel>
				</TabContext>
			</DialogContent>
			<DialogActions>
				<Button onClick={onClose}>Close</Button>
			</DialogActions>
		</Dialog>
	);
}
