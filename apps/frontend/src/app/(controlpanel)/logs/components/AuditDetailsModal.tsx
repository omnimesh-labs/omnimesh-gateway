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
import { AuditLogEntry } from '@/lib/types';

interface AuditDetailsModalProps {
	open: boolean;
	onClose: () => void;
	auditEntry: AuditLogEntry | null;
}

const getActionColor = (action: string): 'default' | 'primary' | 'success' | 'warning' | 'error' => {
	switch (action.toLowerCase()) {
		case 'create':
		case 'created':
			return 'success';
		case 'update':
		case 'updated':
		case 'modify':
		case 'modified':
			return 'primary';
		case 'delete':
		case 'deleted':
		case 'remove':
		case 'removed':
			return 'error';
		case 'login':
		case 'logout':
		case 'access':
			return 'warning';
		default:
			return 'default';
	}
};

const copyToClipboard = async (text: string) => {
	try {
		await navigator.clipboard.writeText(text);
	} catch (err) {
		console.error('Failed to copy text: ', err);
	}
};

const formatJsonData = (data: any): string => {
	if (typeof data === 'string') return data;
	return JSON.stringify(data, null, 2);
};

export default function AuditDetailsModal({ open, onClose, auditEntry }: AuditDetailsModalProps) {
	const theme = useTheme();
	const [expandedSection, setExpandedSection] = useState<string | false>('basic');

	if (!auditEntry) return null;

	const handleAccordionChange = (panel: string) => (event: React.SyntheticEvent, isExpanded: boolean) => {
		setExpandedSection(isExpanded ? panel : false);
	};

	return (
		<Dialog open={open} onClose={onClose} maxWidth="lg" fullWidth>
			<DialogTitle>
				<Box display="flex" alignItems="center" justifyContent="space-between">
					<Typography variant="h6">Audit Log Details</Typography>
					<Box display="flex" alignItems="center" gap={1}>
						<Chip
							size="small"
							label={auditEntry.action.toUpperCase()}
							color={getActionColor(auditEntry.action)}
						/>
						<Chip
							size="small"
							label={auditEntry.resource_type}
							color="default"
							variant="outlined"
						/>
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
											{new Date(auditEntry.created_at).toLocaleString()}
										</Typography>
									</Paper>
								</Grid>
								<Grid item xs={12} md={6}>
									<Paper sx={{ p: 2, bgcolor: theme.palette.background.default }}>
										<Typography variant="subtitle2" color="textSecondary">
											Audit ID
										</Typography>
										<Box display="flex" alignItems="center" gap={1}>
											<Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
												{auditEntry.id}
											</Typography>
											<Tooltip title="Copy to clipboard">
												<IconButton size="small" onClick={() => copyToClipboard(auditEntry.id)}>
													<SvgIcon size={16}>lucide:copy</SvgIcon>
												</IconButton>
											</Tooltip>
										</Box>
									</Paper>
								</Grid>
								<Grid item xs={12} md={6}>
									<Paper sx={{ p: 2, bgcolor: theme.palette.background.default }}>
										<Typography variant="subtitle2" color="textSecondary">
											Action
										</Typography>
										<Typography variant="body2">{auditEntry.action}</Typography>
									</Paper>
								</Grid>
								<Grid item xs={12} md={6}>
									<Paper sx={{ p: 2, bgcolor: theme.palette.background.default }}>
										<Typography variant="subtitle2" color="textSecondary">
											Resource Type
										</Typography>
										<Typography variant="body2">{auditEntry.resource_type}</Typography>
									</Paper>
								</Grid>
								{auditEntry.resource_id && (
									<Grid item xs={12}>
										<Paper sx={{ p: 2, bgcolor: theme.palette.background.default }}>
											<Typography variant="subtitle2" color="textSecondary">
												Resource ID
											</Typography>
											<Box display="flex" alignItems="center" gap={1}>
												<Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
													{auditEntry.resource_id}
												</Typography>
												<Tooltip title="Copy to clipboard">
													<IconButton
														size="small"
														onClick={() => copyToClipboard(auditEntry.resource_id || '')}
													>
														<SvgIcon size={16}>lucide:copy</SvgIcon>
													</IconButton>
												</Tooltip>
											</Box>
										</Paper>
									</Grid>
								)}
							</Grid>
						</AccordionDetails>
					</Accordion>

					{/* Actor Information */}
					<Accordion expanded={expandedSection === 'actor'} onChange={handleAccordionChange('actor')}>
						<AccordionSummary expandIcon={<SvgIcon>lucide:chevron-down</SvgIcon>}>
							<Typography variant="h6">Actor Information</Typography>
						</AccordionSummary>
						<AccordionDetails>
							<Grid container spacing={2}>
								<Grid item xs={12} md={6}>
									<Paper sx={{ p: 2, bgcolor: theme.palette.background.default }}>
										<Typography variant="subtitle2" color="textSecondary">
											Actor ID
										</Typography>
										<Box display="flex" alignItems="center" gap={1}>
											<Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
												{auditEntry.actor_id}
											</Typography>
											<Tooltip title="Copy to clipboard">
												<IconButton size="small" onClick={() => copyToClipboard(auditEntry.actor_id)}>
													<SvgIcon size={16}>lucide:copy</SvgIcon>
												</IconButton>
											</Tooltip>
										</Box>
									</Paper>
								</Grid>
								{auditEntry.actor_ip && (
									<Grid item xs={12} md={6}>
										<Paper sx={{ p: 2, bgcolor: theme.palette.background.default }}>
											<Typography variant="subtitle2" color="textSecondary">
												Actor IP Address
											</Typography>
											<Typography variant="body2">{auditEntry.actor_ip}</Typography>
										</Paper>
									</Grid>
								)}
								<Grid item xs={12} md={6}>
									<Paper sx={{ p: 2, bgcolor: theme.palette.background.default }}>
										<Typography variant="subtitle2" color="textSecondary">
											Organization ID
										</Typography>
										<Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
											{auditEntry.organization_id}
										</Typography>
									</Paper>
								</Grid>
							</Grid>
						</AccordionDetails>
					</Accordion>

					{/* Change Details */}
					{(auditEntry.old_values || auditEntry.new_values) && (
						<Accordion expanded={expandedSection === 'changes'} onChange={handleAccordionChange('changes')}>
							<AccordionSummary expandIcon={<SvgIcon>lucide:chevron-down</SvgIcon>}>
								<Typography variant="h6">Change Details</Typography>
							</AccordionSummary>
							<AccordionDetails>
								<Grid container spacing={2}>
									{auditEntry.old_values && (
										<Grid item xs={12} md={6}>
											<Paper sx={{ p: 2, bgcolor: theme.palette.background.default }}>
												<Box display="flex" alignItems="center" justifyContent="between" mb={1}>
													<Typography variant="subtitle2" color="textSecondary">
														Old Values
													</Typography>
													<Tooltip title="Copy to clipboard">
														<IconButton
															size="small"
															onClick={() => copyToClipboard(formatJsonData(auditEntry.old_values))}
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
														maxHeight: '200px',
														overflow: 'auto'
													}}
												>
													{formatJsonData(auditEntry.old_values)}
												</Typography>
											</Paper>
										</Grid>
									)}
									{auditEntry.new_values && (
										<Grid item xs={12} md={auditEntry.old_values ? 6 : 12}>
											<Paper sx={{ p: 2, bgcolor: theme.palette.background.default }}>
												<Box display="flex" alignItems="center" justifyContent="between" mb={1}>
													<Typography variant="subtitle2" color="textSecondary">
														New Values
													</Typography>
													<Tooltip title="Copy to clipboard">
														<IconButton
															size="small"
															onClick={() => copyToClipboard(formatJsonData(auditEntry.new_values))}
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
														maxHeight: '200px',
														overflow: 'auto'
													}}
												>
													{formatJsonData(auditEntry.new_values)}
												</Typography>
											</Paper>
										</Grid>
									)}
								</Grid>
							</AccordionDetails>
						</Accordion>
					)}

					{/* Metadata */}
					{auditEntry.metadata && Object.keys(auditEntry.metadata).length > 0 && (
						<Accordion expanded={expandedSection === 'metadata'} onChange={handleAccordionChange('metadata')}>
							<AccordionSummary expandIcon={<SvgIcon>lucide:chevron-down</SvgIcon>}>
								<Typography variant="h6">Metadata</Typography>
							</AccordionSummary>
							<AccordionDetails>
								<Paper sx={{ p: 2, bgcolor: theme.palette.background.default }}>
									<Box display="flex" alignItems="center" justifyContent="between" mb={1}>
										<Typography variant="subtitle2" color="textSecondary">
											Additional Metadata
										</Typography>
										<Tooltip title="Copy to clipboard">
											<IconButton
												size="small"
												onClick={() => copyToClipboard(formatJsonData(auditEntry.metadata))}
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
										{formatJsonData(auditEntry.metadata)}
									</Typography>
								</Paper>
							</AccordionDetails>
						</Accordion>
					)}

					{/* Raw Data */}
					<Accordion expanded={expandedSection === 'raw'} onChange={handleAccordionChange('raw')}>
						<AccordionSummary expandIcon={<SvgIcon>lucide:chevron-down</SvgIcon>}>
							<Typography variant="h6">Raw Audit Data</Typography>
						</AccordionSummary>
						<AccordionDetails>
							<Paper sx={{ p: 2, bgcolor: theme.palette.background.default }}>
								<Box display="flex" alignItems="center" justifyContent="between" mb={1}>
									<Typography variant="subtitle2" color="textSecondary">
										Complete Audit Entry (JSON)
									</Typography>
									<Tooltip title="Copy to clipboard">
										<IconButton
											size="small"
											onClick={() => copyToClipboard(JSON.stringify(auditEntry, null, 2))}
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
									{JSON.stringify(auditEntry, null, 2)}
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
