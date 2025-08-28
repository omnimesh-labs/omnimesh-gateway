'use client';

import { useState } from 'react';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import {
	Dialog,
	DialogTitle,
	DialogContent,
	DialogActions,
	Button,
	TextField,
	FormControl,
	InputLabel,
	Select,
	MenuItem,
	FormHelperText,
	Chip,
	Box,
	Typography,
	Divider
} from '@mui/material';
import SvgIcon from '@fuse/core/SvgIcon';
import { CreateServerRequest } from '@/lib/client-api'; import type { MCPServer, CreateServerRequest, Namespace } from '@/lib/types';

const schema = z.object({
	name: z.string().min(1, 'Server name is required'),
	description: z.string().optional(),
	protocol: z.enum(['stdio', 'http', 'websocket'], {
		message: 'Please select a protocol'
	}),
	url: z.string().url('Invalid URL').optional().or(z.literal('')),
	command: z.string().optional(),
	args: z.string().optional(),
	environment: z.string().optional(),
	version: z.string().optional(),
	timeout: z.number().int().min(1).max(300).optional(),
	max_retries: z.number().int().min(0).max(10).optional(),
	health_check_url: z.string().url('Invalid URL').optional().or(z.literal('')),
	working_dir: z.string().optional()
});

type FormType = z.infer<typeof schema>;

interface RegisterServerModalProps {
	open: boolean;
	onClose: () => void;
	onRegister: (data: CreateServerRequest) => void;
	loading?: boolean;
}

export default function RegisterServerModal({ open, onClose, onRegister, loading = false }: RegisterServerModalProps) {
	const [tags, setTags] = useState<string[]>([]);
	const [tagInput, setTagInput] = useState('');

	const {
		control,
		handleSubmit,
		watch,
		reset,
		formState: { errors, isValid }
	} = useForm<FormType>({
		resolver: zodResolver(schema),
		defaultValues: {
			name: '',
			description: '',
			protocol: 'stdio',
			url: '',
			command: '',
			args: '',
			environment: '',
			version: '1.0.0',
			timeout: 30,
			max_retries: 3,
			health_check_url: '',
			working_dir: ''
		}
	});

	const protocol = watch('protocol');

	const handleAddTag = () => {
		if (tagInput.trim() && !tags.includes(tagInput.trim())) {
			setTags([...tags, tagInput.trim()]);
			setTagInput('');
		}
	};

	const handleRemoveTag = (tagToRemove: string) => {
		setTags(tags.filter((tag) => tag !== tagToRemove));
	};

	const handleClose = () => {
		reset();
		setTags([]);
		setTagInput('');
		onClose();
	};

	const onSubmit = (data: FormType) => {
		const serverData: CreateServerRequest = {
			name: data.name,
			description: data.description || '',
			protocol: data.protocol,
			url: data.url || '',
			command: data.command || '',
			version: data.version || '',
			timeout: data.timeout || 30,
			max_retries: data.max_retries || 3,
			health_check_url: data.health_check_url || '',
			working_dir: data.working_dir || '',
			args: data.args ? data.args.split(' ').filter((arg) => arg.trim()) : undefined,
			environment: data.environment ? data.environment.split('\n').filter((env) => env.trim()) : undefined,
			metadata: tags.length > 0 ? { tags: tags.join(',') } : undefined
		};

		// Clean up empty fields
		Object.keys(serverData).forEach((key) => {
			const value = serverData[key as keyof CreateServerRequest];

			if (value === '' || value === undefined) {
				delete serverData[key as keyof CreateServerRequest];
			}
		});

		onRegister(serverData);
	};

	return (
		<Dialog
			open={open}
			onClose={handleClose}
			maxWidth="md"
			fullWidth
		>
			<DialogTitle>
				<Box className="flex items-center space-x-2">
					<SvgIcon>lucide:server</SvgIcon>
					<Typography variant="h6">Register New Server</Typography>
				</Box>
			</DialogTitle>

			<form onSubmit={handleSubmit(onSubmit)}>
				<DialogContent>
					<Box className="space-y-4">
						{/* Basic Information */}
						<Typography
							variant="h6"
							className="mb-3"
						>
							Basic Information
						</Typography>

						<Controller
							name="name"
							control={control}
							render={({ field }) => (
								<TextField
									{...field}
									label="Server Name"
									fullWidth
									error={!!errors.name}
									helperText={errors.name?.message}
									required
								/>
							)}
						/>

						<Controller
							name="description"
							control={control}
							render={({ field }) => (
								<TextField
									{...field}
									label="Description"
									fullWidth
									multiline
									rows={2}
									error={!!errors.description}
									helperText={errors.description?.message}
								/>
							)}
						/>

						<Controller
							name="protocol"
							control={control}
							render={({ field }) => (
								<FormControl
									fullWidth
									error={!!errors.protocol}
								>
									<InputLabel>Protocol</InputLabel>
									<Select
										{...field}
										label="Protocol"
									>
										<MenuItem value="stdio">STDIO</MenuItem>
										<MenuItem value="http">HTTP</MenuItem>
										<MenuItem value="websocket">WebSocket</MenuItem>
									</Select>
									{errors.protocol && <FormHelperText>{errors.protocol.message}</FormHelperText>}
								</FormControl>
							)}
						/>

						<Divider className="my-4" />

						{/* Protocol-specific fields */}
						<Typography
							variant="h6"
							className="mb-3"
						>
							{protocol === 'stdio' ? 'Command Configuration' : 'Connection Configuration'}
						</Typography>

						{protocol === 'stdio' ? (
							<>
								<Controller
									name="command"
									control={control}
									render={({ field }) => (
										<TextField
											{...field}
											label="Command"
											fullWidth
											placeholder="e.g., node server.js"
											// helperText="Command to start the MCP server"
										/>
									)}
								/>

								<Controller
									name="args"
									control={control}
									render={({ field }) => (
										<TextField
											{...field}
											label="Arguments"
											fullWidth
											placeholder="e.g., --port 3000 --verbose"
											// helperText="Space-separated command line arguments"
										/>
									)}
								/>

								<Controller
									name="working_dir"
									control={control}
									render={({ field }) => (
										<TextField
											{...field}
											label="Working Directory"
											fullWidth
											placeholder="e.g., /path/to/server"
											// helperText="Working directory for the server process"
										/>
									)}
								/>
							</>
						) : (
							<Controller
								name="url"
								control={control}
								render={({ field }) => (
									<TextField
										{...field}
										label="Server URL"
										fullWidth
										placeholder="e.g., http://localhost:3000 or ws://localhost:3000"
										error={!!errors.url}
										helperText={errors.url?.message}
									// helperText={errors.url?.message || 'URL to connect to the MCP server'}
									/>
								)}
							/>
						)}

						<Controller
							name="environment"
							control={control}
							render={({ field }) => (
								<TextField
									{...field}
									label="Environment Variables"
									fullWidth
									multiline
									rows={3}
									placeholder="KEY=value&#10;ANOTHER_KEY=another_value"
									// helperText="One environment variable per line (KEY=value format)"
								/>
							)}
						/>

						<Divider className="my-4" />

						{/* Advanced Configuration */}
						<Typography
							variant="h6"
							className="mb-3"
						>
							Advanced Configuration
						</Typography>

						<Box className="grid grid-cols-2 gap-4">
							<Controller
								name="version"
								control={control}
								render={({ field }) => (
									<TextField
										{...field}
										label="Version"
										placeholder="1.0.0"
									/>
								)}
							/>

							<Controller
								name="timeout"
								control={control}
								render={({ field }) => (
									<TextField
										{...field}
										label="Timeout (seconds)"
										type="number"
										inputProps={{ min: 1, max: 300 }}
										error={!!errors.timeout}
										helperText={errors.timeout?.message}
									/>
								)}
							/>
						</Box>

						<Box className="grid grid-cols-2 gap-4">
							<Controller
								name="max_retries"
								control={control}
								render={({ field }) => (
									<TextField
										{...field}
										label="Max Retries"
										type="number"
										inputProps={{ min: 0, max: 10 }}
										error={!!errors.max_retries}
										helperText={errors.max_retries?.message}
									/>
								)}
							/>

							<Controller
								name="health_check_url"
								control={control}
								render={({ field }) => (
									<TextField
										{...field}
										label="Health Check URL"
										placeholder="http://localhost:3000/health"
										error={!!errors.health_check_url}
										helperText={errors.health_check_url?.message}
									/>
								)}
							/>
						</Box>

						{/* Tags */}
						<Box>
							<Typography
								variant="body2"
								className="mb-2"
							>
								Tags
							</Typography>
							<Box className="mb-2 flex flex-wrap gap-1">
								{tags.map((tag) => (
									<Chip
										key={tag}
										label={tag}
										size="small"
										onDelete={() => handleRemoveTag(tag)}
									/>
								))}
							</Box>
							<Box className="flex gap-2">
								<TextField
									size="small"
									placeholder="Add tag"
									value={tagInput}
									onChange={(e) => setTagInput(e.target.value)}
									onKeyPress={(e) => e.key === 'Enter' && (e.preventDefault(), handleAddTag())}
								/>
								<Button
									size="small"
									onClick={handleAddTag}
									disabled={!tagInput.trim()}
								>
									Add
								</Button>
							</Box>
						</Box>
					</Box>
				</DialogContent>

				<DialogActions>
					<Button onClick={handleClose}>Cancel</Button>
					<Button
						type="submit"
						variant="contained"
						disabled={!isValid || loading}
						startIcon={loading ? <SvgIcon>lucide:loader-2</SvgIcon> : <SvgIcon>lucide:plus</SvgIcon>}
					>
						{loading ? 'Registering...' : 'Register Server'}
					</Button>
				</DialogActions>
			</form>
		</Dialog>
	);
}
