import {
  Box,
  MenuItem,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TablePagination,
  TableRow,
  TextField,
  Typography,
} from "@mui/material";
import { Search } from "lucide-react";
import { type ReactNode, useMemo, useState } from "react";

export type DataColumn<T> = {
  key: string;
  label: string;
  render?: (row: T) => ReactNode;
  align?: "left" | "right" | "center";
};

export type DataFilter<T> = {
  key: string;
  label: string;
  options: string[];
  valueOf: (row: T) => string;
};

type DataTableProps<T> = {
  rows: T[];
  columns: DataColumn<T>[];
  getRowKey: (row: T) => string;
  /** Enables the search box; return the searchable text for a row. */
  searchOf?: (row: T) => string;
  searchPlaceholder?: string;
  filters?: DataFilter<T>[];
  /** Trailing per-row actions column (View / Edit / Delete). */
  rowActions?: (row: T) => ReactNode;
  rowActionsLabel?: string;
  /** Header-right slot, e.g. an "Add" button. */
  toolbarActions?: ReactNode;
  pageSize?: number;
  emptyState?: ReactNode;
  emptyMessage?: string;
};

/**
 * Command-center data grid: client-side search, filters, and pagination over a
 * `cc-*`-styled MUI table with an optional trailing actions column. Shared by
 * the entity management tabs (shelters, relief, aid, …).
 */
export function DataTable<T>({
  rows,
  columns,
  getRowKey,
  searchOf,
  searchPlaceholder = "Search",
  filters = [],
  rowActions,
  rowActionsLabel = "Actions",
  toolbarActions,
  pageSize = 8,
  emptyState,
  emptyMessage = "No records to show.",
}: DataTableProps<T>) {
  const [query, setQuery] = useState("");
  const [filterValues, setFilterValues] = useState<Record<string, string>>({});
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(pageSize);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    return rows.filter((row) => {
      if (q && searchOf && !searchOf(row).toLowerCase().includes(q)) {
        return false;
      }
      for (const filter of filters) {
        const selected = filterValues[filter.key];
        if (selected && filter.valueOf(row) !== selected) {
          return false;
        }
      }
      return true;
    });
  }, [rows, query, filters, filterValues, searchOf]);

  const lastPage = Math.max(0, Math.ceil(filtered.length / rowsPerPage) - 1);
  const currentPage = Math.min(page, lastPage);
  const pageStart = currentPage * rowsPerPage;
  const pageRows = filtered.slice(pageStart, pageStart + rowsPerPage);
  const colCount = columns.length + (rowActions ? 1 : 0);

  return (
    <Box>
      <Stack
        direction={{ xs: "column", md: "row" }}
        spacing={1.5}
        sx={{
          alignItems: { xs: "stretch", md: "center" },
          mb: 1.5,
          flexWrap: "wrap"
        }}>
        {searchOf ? (
          <TextField
            onChange={(event) => {
              setQuery(event.target.value);
              setPage(0);
            }}
            placeholder={searchPlaceholder}
            size="small"
            sx={{ minWidth: { xs: "100%", md: 240 } }}
            value={query}
            slotProps={{
              input: {
                startAdornment: (
                  <Search
                    size={16}
                    style={{ marginRight: 8, opacity: 0.6, flexShrink: 0 }}
                  />
                ),
              }
            }}
          />
        ) : null}
        {filters.map((filter) => (
          <TextField
            key={filter.key}
            label={filter.label}
            onChange={(event) => {
              setFilterValues((current) => ({
                ...current,
                [filter.key]: event.target.value,
              }));
              setPage(0);
            }}
            select
            size="small"
            sx={{ minWidth: 150 }}
            value={filterValues[filter.key] ?? ""}
          >
            <MenuItem value="">All</MenuItem>
            {filter.options.map((option) => (
              <MenuItem key={option} value={option}>
                {option}
              </MenuItem>
            ))}
          </TextField>
        ))}
        {toolbarActions ? (
          <Box sx={{ ml: { md: "auto" } }}>{toolbarActions}</Box>
        ) : null}
      </Stack>
      <TableContainer sx={{ maxWidth: "100%" }}>
        <Table size="small">
          <TableHead>
            <TableRow>
              {columns.map((column) => (
                <TableCell
                  align={column.align}
                  key={column.key}
                  sx={{ fontWeight: 800, whiteSpace: "nowrap" }}
                >
                  {column.label}
                </TableCell>
              ))}
              {rowActions ? (
                <TableCell align="right" sx={{ fontWeight: 800 }}>
                  {rowActionsLabel}
                </TableCell>
              ) : null}
            </TableRow>
          </TableHead>
          <TableBody>
            {pageRows.length === 0 ? (
              <TableRow>
                <TableCell colSpan={colCount} sx={{ borderBottom: 0 }}>
                  {emptyState ?? (
                    <Typography
                      align="center"
                      sx={{ py: 3, color: "text.secondary" }}
                    >
                      {emptyMessage}
                    </Typography>
                  )}
                </TableCell>
              </TableRow>
            ) : (
              pageRows.map((row) => (
                <TableRow hover key={getRowKey(row)}>
                  {columns.map((column) => (
                    <TableCell align={column.align} key={column.key}>
                      {column.render
                        ? column.render(row)
                        : String(
                            (row as Record<string, unknown>)[column.key] ?? "",
                          )}
                    </TableCell>
                  ))}
                  {rowActions ? (
                    <TableCell align="right">{rowActions(row)}</TableCell>
                  ) : null}
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </TableContainer>
      <TablePagination
        component="div"
        count={filtered.length}
        onPageChange={(_event, next) => setPage(next)}
        onRowsPerPageChange={(event) => {
          setRowsPerPage(parseInt(event.target.value, 10));
          setPage(0);
        }}
        page={currentPage}
        rowsPerPage={rowsPerPage}
        rowsPerPageOptions={[8, 16, 32]}
      />
    </Box>
  );
}
