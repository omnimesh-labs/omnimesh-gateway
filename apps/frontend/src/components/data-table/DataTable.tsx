import {
	MaterialReactTable,
	useMaterialReactTable,
	MaterialReactTableProps,
	MRT_Icons,
	MRT_RowData
} from 'material-react-table';
import { defaults } from '../../utils/lodashReplacements';
import { useMemo, useEffect, useRef } from 'react';
import SvgIcon from '@fuse/core/SvgIcon';
import { Theme } from '@mui/material/styles';
import DataTableTopToolbar from './DataTableTopToolbar';
import { useThemeMediaQuery } from '@fuse/hooks';

const tableIcons: Partial<MRT_Icons> = {
	ArrowDownwardIcon: (props) => <SvgIcon {...props}>lucide:arrow-down</SvgIcon>,
	ClearAllIcon: () => <SvgIcon>lucide:brush-cleaning</SvgIcon>,
	DensityLargeIcon: () => <SvgIcon>lucide:rows-2</SvgIcon>,
	DensityMediumIcon: () => <SvgIcon>lucide:rows-3</SvgIcon>,
	DensitySmallIcon: () => <SvgIcon>lucide:rows-4</SvgIcon>,
	DragHandleIcon: () => <SvgIcon>lucide:grip-vertical</SvgIcon>,
	FilterListIcon: (props) => <SvgIcon {...props}>lucide:list-filter</SvgIcon>,
	FilterListOffIcon: () => <SvgIcon>lucide:funnel</SvgIcon>,
	FullscreenExitIcon: () => <SvgIcon>lucide:log-in</SvgIcon>,
	FullscreenIcon: () => <SvgIcon>lucide:log-out</SvgIcon>,
	SearchIcon: (props) => <SvgIcon {...props}>lucide:search</SvgIcon>,
	SearchOffIcon: () => <SvgIcon>lucide:search-x</SvgIcon>,
	ViewColumnIcon: () => <SvgIcon>lucide:columns-3-cog</SvgIcon>,
	MoreVertIcon: () => <SvgIcon>lucide:ellipsis-vertical</SvgIcon>,
	MoreHorizIcon: () => <SvgIcon>lucide:ellipsis</SvgIcon>,
	SortIcon: (props) => <SvgIcon {...props}>lucide:arrow-down-up</SvgIcon>,
	PushPinIcon: (props) => <SvgIcon {...props}>lucide:pin</SvgIcon>,
	VisibilityOffIcon: () => <SvgIcon>lucide:eye-off</SvgIcon>
};

function DataTable<TData extends MRT_RowData>(props: MaterialReactTableProps<TData>) {
	const { columns, data, initialState, ...rest } = props;

	// Ensure data and columns are never null/undefined to prevent Material React Table errors
	const safeData = data ?? [];
	const safeColumns = columns ?? [];
	const isMobile = useThemeMediaQuery((theme) => theme.breakpoints.down('lg'));
	const isMountedRef = useRef(false);

	// Merge initial state with defaults, avoiding pagination reset during render
	const mergedInitialState = useMemo(
		() => ({
			density: 'compact',
			showColumnFilters: false,
			showGlobalFilter: true,
			columnPinning: {
				left: isMobile ? [] : ['mrt-row-expand', 'mrt-row-select'],
				right: isMobile ? [] : ['mrt-row-actions']
			},
			pagination: {
				pageSize: 15,
				pageIndex: 0
			},
			enableFullScreenToggle: false,
			...initialState
		}),
		[initialState, isMobile]
	);

	const tableDefaults = useMemo(
		() =>
			defaults(rest, {
				initialState: mergedInitialState,
				enableFullScreenToggle: false,
				enableColumnFilterModes: true,
				enableColumnOrdering: true,
				enableGrouping: true,
				enableColumnPinning: true,
				enableFacetedValues: true,
				enableRowActions: true,
				enableRowSelection: true,
				muiBottomToolbarProps: {
					className: 'flex items-center min-h-14 h-14'
				},
				muiTablePaperProps: {
					elevation: 0,
					square: true,
					className: 'flex flex-col flex-auto h-full'
				},
				muiTableContainerProps: {
					className: 'flex-auto'
				},
				enableStickyHeader: true,
				// enableStickyFooter: true,
				paginationDisplayMode: 'pages',
				positionToolbarAlertBanner: 'top',
				muiPaginationProps: {
					color: 'secondary',
					rowsPerPageOptions: [10, 20, 30],
					shape: 'rounded',
					variant: 'outlined',
					showRowsPerPage: false
				},
				muiSearchTextFieldProps: {
					placeholder: 'Search',
					sx: { minWidth: '300px' },
					variant: 'outlined',
					size: 'small'
				},
				muiFilterTextFieldProps: {
					variant: 'outlined',
					size: 'small',
					sx: {
						'& .MuiInputAdornment-root': {
							padding: 0,
							margin: 0
						},
						'& .MuiInputBase-root': {
							padding: 0
						},
						'& .MuiInputBase-input': {
							padding: 0
						}
					}
				},
				muiSelectAllCheckboxProps: {
					size: 'small'
				},
				muiSelectCheckboxProps: {
					size: 'small'
				},
				muiTableBodyRowProps: ({ row, table }) => {
					const { density } = table.getState();

					if (density === 'compact') {
						return {
							sx: {
								backgroundColor: 'initial',
								opacity: 1,
								boxShadow: 'none',
								height: row.getIsPinned() ? `${37}px` : undefined
							}
						};
					}

					return {
						sx: {
							backgroundColor: 'initial',
							opacity: 1,
							boxShadow: 'none',
							// Set a fixed height for pinned rows
							height: row.getIsPinned() ? `${density === 'comfortable' ? 53 : 69}px` : undefined
						}
					};
				},
				muiTableHeadCellProps: ({ column }) => ({
					sx: {
						'& .Mui-TableHeadCell-Content-Labels': {
							flex: 1,
							justifyContent: 'space-between'
						},
						'& .Mui-TableHeadCell-Content-Actions': {
							'& > button': {
								marginX: '2px'
							}
						},
						'& .MuiFormHelperText-root': {
							textAlign: 'center',
							marginX: 0,
							color: (theme: Theme) => theme.vars.palette.text.disabled,
							fontSize: 11
						},
						backgroundColor: (theme) =>
							column.getIsPinned() ? theme.vars.palette.background.paper : 'inherit'
					}
				}),
				mrtTheme: (theme) => ({
					baseBackgroundColor: theme.palette.background.paper,
					menuBackgroundColor: theme.palette.background.paper,
					pinnedRowBackgroundColor: theme.palette.background.paper,
					pinnedColumnBackgroundColor: theme.palette.background.paper
				}),
				renderTopToolbar: (_props) => <DataTableTopToolbar<TData> {..._props} />,
				icons: tableIcons,
				positionActionsColumn: 'last'
			} as Partial<MaterialReactTableProps<TData>>),

		[rest, mergedInitialState]
	);

	// Track mount status
	useEffect(() => {
		isMountedRef.current = true;
		return () => {
			isMountedRef.current = false;
		};
	}, []);

	const tableOptions = useMemo(
		() => ({
			columns: safeColumns,
			data: safeData,
			...tableDefaults,
			...rest,
			// Override autoResetPageIndex to prevent unwanted resets
			autoResetPageIndex: false
		}),
		[safeColumns, safeData, tableDefaults, rest]
	);

	const tableInstance = useMaterialReactTable<TData>(tableOptions);

	return <MaterialReactTable table={tableInstance} />;
}

export default DataTable;
