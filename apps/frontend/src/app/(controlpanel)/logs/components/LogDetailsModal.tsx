'use client';

import { useState } from 'react';
import {
	Dialog,
	DialogTitle,
	DialogContent,
	DialogActions,
	Button,
	Typography,
	Box,
	Chip,
	Grid,
	Paper,
	Divider,
	IconButton,
	Tooltip,
	Accordion,
	AccordionSummary,
	AccordionDetails
} from '@mui/material';
import { useTheme } from '@mui/material/styles';
import SvgIcon from '@fuse/core/SvgIcon';
import { LogEntry } from '@/lib/types';

interface LogDetailsModalProps {
	open: boolean;
	onClose: () => void;
	logEntry: LogEntry | null;
}

const getLogLevelColor = (level: string): 'default' | 'info' | 'warning' | 'error' => {
	switch (level) {
		case 'error':
			return 'error';
		case 'warn':
			return 'warning';
		case 'info':
			return 'info';
		default:
			return 'default';
	}
};

const formatJsonString = (jsonStr: string | undefined): string => {
	if (!jsonStr) return '';
	try {
		return JSON.stringify(JSON.parse(jsonStr), null, 2);
	} catch {
		return jsonStr;
	}
};

const copyToClipboard = async (text: string) => {
	try {
		await navigator.clipboard.writeText(text);
	} catch (err) {
		console.error('Failed to copy text: ', err);
	}
};

export default function LogDetailsModal({ open, onClose, logEntry }: LogDetailsModalProps) {
	const theme = useTheme();
	const [expandedSection, setExpandedSection] = useState<string | false>('basic');

	if (!logEntry) return null;

	const handleAccordionChange = (panel: string) => (event: React.SyntheticEvent, isExpanded: boolean) => {
		setExpandedSection(isExpanded ? panel : false);
	};

	return (
		<Dialog open={open} onClose={onClose} maxWidth="lg" fullWidth>
			<DialogTitle>
				<Box display="flex" alignItems="center" justifyContent="space-between">
					<Typography variant="h6">Log Entry Details</Typography>
					<Box display="flex" alignItems="center" gap={1}>
						<Chip
							size="small"
							label={logEntry.level.toUpperCase()}
							color={getLogLevelColor(logEntry.level)}
						/>
						<Chip size="small" label={logEntry.environment} color="default" variant="outlined" />
					</Box>
				</Box>
			</DialogTitle>

			<DialogContent>
				<Box sx={{ width: '100%' }}>
					{/* Basic Information */}
					<Accordion expanded={expandedSection === 'basic'} onChange={handleAccordionChange('basic')}>
						<AccordionSummary expandIcon={<SvgIcon>lucide:chevron-down</SvgIcon>}>
							<Typography variant="h6">Basic Information</Typography>
						</AccordionSummary>
						<AccordionDetails>
							<Grid container spacing={2}>
								<Grid item xs={12} md={6}>
									<Paper sx={{ p: 2, bgcolor: theme.palette.background.default }}>
										<Typography variant="subtitle2" color="textSecondary">
											Timestamp
										</Typography>
										<Typography variant="body2">
											{logEntry.timestamp ? new Date(logEntry.timestamp).toLocaleString() : '-'}
										</Typography>
									</Paper>
								</Grid>
								<Grid item xs={12} md={6}>
									<Paper sx={{ p: 2, bgcolor: theme.palette.background.default }}>
										<Typography variant="subtitle2" color="textSecondary">
											Request ID
										</Typography>
										<Box display="flex" alignItems="center" gap={1}>
											<Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
												{logEntry.request_id || '-'}
											</Typography>
											{logEntry.request_id && (
												<Tooltip title="Copy to clipboard">
													<IconButton
														size="small"
														onClick={() => copyToClipboard(logEntry.request_id)}
													>
														<SvgIcon size={16}>lucide:copy</SvgIcon>
													</IconButton>
												</Tooltip>
											)}
										</Box>
									</Paper>
								</Grid>
								<Grid item xs={12} md={6}>
									<Paper sx={{ p: 2, bgcolor: theme.palette.background.default }}>
										<Typography variant="subtitle2" color="textSecondary">
											Logger
										</Typography>
										<Typography variant="body2">{logEntry.logger || '-'}</Typography>
									</Paper>
								</Grid>
								<Grid item xs={12} md={6}>
									<Paper sx={{ p: 2, bgcolor: theme.palette.background.default }}>
										<Typography variant="subtitle2" color="textSecondary">
											User ID
										</Typography>
										<Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
											{logEntry.user_id || '-'}
										</Typography>
									</Paper>
								</Grid>
								<Grid item xs={12}>
									<Paper sx={{ p: 2, bgcolor: theme.palette.background.default }}>
										<Typography variant="subtitle2" color="textSecondary">
											Message
										</Typography>
										<Typography variant="body2">{logEntry.message || '-'}</Typography>
									</Paper>
								</Grid>
							</Grid>
						</AccordionDetails>
					</Accordion>

					{/* Request Details */}
					{logEntry.data && (
						<Accordion
							expanded={expandedSection === 'request'}
							onChange={handleAccordionChange('request')}
						>
							<AccordionSummary expandIcon={<SvgIcon>lucide:chevron-down</SvgIcon>}>
								<Typography variant="h6">Request Details</Typography>
							</AccordionSummary>
							<AccordionDetails>
								<Grid container spacing={2}>
									{logEntry.data.method && (
										<Grid item xs={12} md={4}>
											<Paper sx={{ p: 2, bgcolor: theme.palette.background.default }}>
												<Typography variant="subtitle2" color="textSecondary">
													Method
												</Typography>
												<Typography variant="body2">{logEntry.data.method}</Typography>
											</Paper>
										</Grid>
									)}
									{logEntry.data.path && (
										<Grid item xs={12} md={4}>
											<Paper sx={{ p: 2, bgcolor: theme.palette.background.default }}>
												<Typography variant="subtitle2" color="textSecondary">
													Path
												</Typography>
												<Typography variant="body2">{logEntry.data.path}</Typography>
											</Paper>
										</Grid>
									)}
									{logEntry.data.remote_ip && (
										<Grid item xs={12} md={4}>
											<Paper sx={{ p: 2, bgcolor: theme.palette.background.default }}>
												<Typography variant="subtitle2" color="textSecondary">
													Remote IP
												</Typography>
												<Typography variant="body2">{logEntry.data.remote_ip}</Typography>
											</Paper>
										</Grid>
									)}
									{logEntry.data.query_params && (
										<Grid item xs={12}>
											<Paper sx={{ p: 2, bgcolor: theme.palette.background.default }}>
												<Box display="flex" alignItems="center" justifyContent="between" mb={1}>
													<Typography variant="subtitle2" color="textSecondary">
														Query Parameters
													</Typography>
													<Tooltip title="Copy to clipboard">
														<IconButton
															size="small"
															onClick={() => copyToClipboard(logEntry.data.query_params || '')}
														>
															<SvgIcon size={16}>lucide:copy</SvgIcon>
														</IconButton>
													</Tooltip>
												</Box>
												<Typography
													variant="body2"
													sx={{
														fontFamily: 'monospace',
														bgcolor: theme.palette.mode === 'dark' ? 'grey.900' : 'grey.100',
														p: 1,
														borderRadius: 1,
														whiteSpace: 'pre-wrap',
														wordBreak: 'break-all'
													}}
												>
													{decodeURIComponent(logEntry.data.query_params)}
												</Typography>
											</Paper>
										</Grid>
									)}
								</Grid>
							</AccordionDetails>
						</Accordion>
					)}

					{/* Response Details */}
					{logEntry.data && (logEntry.data.response_body || logEntry.data.response_size) && (
						<Accordion
							expanded={expandedSection === 'response'}
							onChange={handleAccordionChange('response')}
						>
							<AccordionSummary expandIcon={<SvgIcon>lucide:chevron-down</SvgIcon>}>
								<Typography variant="h6">Response Details</Typography>
							</AccordionSummary>
							<AccordionDetails>
								<Grid container spacing={2}>
									{logEntry.data.status_code && (
										<Grid item xs={12} md={4}>
											<Paper sx={{ p: 2, bgcolor: theme.palette.background.default }}>
												<Typography variant="subtitle2" color="textSecondary">
													Status Code
												</Typography>
												<Typography variant="body2">{logEntry.data.status_code}</Typography>
											</Paper>
										</Grid>
									)}
									{logEntry.data.response_size && (
										<Grid item xs={12} md={4}>
											<Paper sx={{ p: 2, bgcolor: theme.palette.background.default }}>
												<Typography variant="subtitle2" color="textSecondary">
													Response Size
												</Typography>
												<Typography variant="body2">{logEntry.data.response_size} bytes</Typography>
											</Paper>
										</Grid>
									)}
									{logEntry.data.duration_ms && (
										<Grid item xs={12} md={4}>
											<Paper sx={{ p: 2, bgcolor: theme.palette.background.default }}>
												<Typography variant="subtitle2" color="textSecondary">
													Duration
												</Typography>
												<Typography variant="body2">{logEntry.data.duration_ms} ms</Typography>
											</Paper>
										</Grid>
									)}
									{logEntry.data.response_body && (
										<Grid item xs={12}>
											<Paper sx={{ p: 2, bgcolor: theme.palette.background.default }}>
												<Box display="flex" alignItems="center" justifyContent="between" mb={1}>
													<Typography variant="subtitle2" color="textSecondary">
														Response Body
													</Typography>
													<Tooltip title="Copy to clipboard">
														<IconButton
															size="small"
															onClick={() =>
																copyToClipboard(
																	formatJsonString(logEntry.data.response_body) ||
																		logEntry.data.response_body ||
																		''
																)
															}
														>
															<SvgIcon size={16}>lucide:copy</SvgIcon>
														</IconButton>
													</Tooltip>
												</Box>
												<Typography
													variant="body2"
													sx={{
														fontFamily: 'monospace',
														bgcolor: theme.palette.mode === 'dark' ? 'grey.900' : 'grey.100',
														p: 1,
														borderRadius: 1,
														whiteSpace: 'pre-wrap',
														maxHeight: '300px',
														overflow: 'auto'
													}}
												>
													{formatJsonString(logEntry.data.response_body) || logEntry.data.response_body}
												</Typography>
											</Paper>
										</Grid>
									)}
								</Grid>
							</AccordionDetails>
						</Accordion>
					)}

					{/* Raw Data */}
					<Accordion expanded={expandedSection === 'raw'} onChange={handleAccordionChange('raw')}>
						<AccordionSummary expandIcon={<SvgIcon>lucide:chevron-down</SvgIcon>}>
							<Typography variant="h6">Raw Log Data</Typography>
						</AccordionSummary>
						<AccordionDetails>
							<Paper sx={{ p: 2, bgcolor: theme.palette.background.default }}>
								<Box display="flex" alignItems="center" justifyContent="between" mb={1}>
									<Typography variant="subtitle2" color="textSecondary">
										Complete Log Entry (JSON)
									</Typography>
									<Tooltip title="Copy to clipboard">
										<IconButton
											size="small"
											onClick={() => copyToClipboard(JSON.stringify(logEntry, null, 2))}
										>
											<SvgIcon size={16}>lucide:copy</SvgIcon>
										</IconButton>
									</Tooltip>
								</Box>
								<Typography
									variant="body2"
									sx={{
										fontFamily: 'monospace',
										bgcolor: theme.palette.mode === 'dark' ? 'grey.900' : 'grey.100',
										p: 1,
										borderRadius: 1,
										whiteSpace: 'pre-wrap',
										maxHeight: '400px',
										overflow: 'auto'
									}}
								>
									{JSON.stringify(logEntry, null, 2)}
								</Typography>
							</Paper>
						</AccordionDetails>
					</Accordion>
				</Box>
			</DialogContent>

			<DialogActions>
				<Button onClick={onClose} variant="outlined">
					Close
				</Button>
			</DialogActions>
		</Dialog>
	);
}
