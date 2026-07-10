import {
  Box,
  MenuItem,
  Paper,
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

export type DataTableColumn<T> = {
  key: string;
  label: string;
  /** Custom cell renderer; defaults to String(row[key]). */
  render?: (row: T) => ReactNode;
  align?: "left" | "right" | "center";
};

export type DataTableFilter<T> = {
  key: string;
  label: string;
  options: string[];
  /** Value for a row, defaults to String(row[key]). */
  valueOf?: (row: T) => string;
};

type DataTableProps<T> = {
  rows: T[];
  columns: DataTableColumn<T>[];
  getRowKey: (row: T) => string;
  /** Free-text search over these values for each row. */
  searchOf?: (row: T) => string;
  searchPlaceholder?: string;
  filters?: DataTableFilter<T>[];
  pageSize?: number;
  emptyMessage?: string;
  /** Rich empty content (e.g. <EmptyState/>) shown when there are no rows;
   * falls back to the emptyMessage text. */
  emptyState?: ReactNode;
  toolbarActions?: ReactNode;
  /** When set, rows become clickable (and keyboard-activatable) and this fires
   * with the row — used to open a detail dialog or navigate to a detail page. */
  onRowClick?: (row: T) => void;
};

/**
 * Public, client-side data table with search, filters, and pagination. Used to
 * let anyone browse citizen data (missing persons, incidents, aid) without an
 * account; reporting sits behind a separate auth-gated button.
 */
export function DataTable<T>({
  rows,
  columns,
  getRowKey,
  searchOf,
  searchPlaceholder = "Search",
  filters = [],
  pageSize = 8,
  emptyMessage = "No records to show.",
  emptyState,
  toolbarActions,
  onRowClick,
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
        if (!selected) {
          continue;
        }
        const value = filter.valueOf
          ? filter.valueOf(row)
          : String((row as Record<string, unknown>)[filter.key] ?? "");
        if (value !== selected) {
          return false;
        }
      }
      return true;
    });
  }, [rows, query, filters, filterValues, searchOf]);

  // Clamp the page so a shrinking `rows` prop (e.g. a refresh returning fewer
  // records) can't strand the view on an out-of-range, blank page.
  const lastPage = Math.max(0, Math.ceil(filtered.length / rowsPerPage) - 1);
  const currentPage = Math.min(page, lastPage);
  const pageStart = currentPage * rowsPerPage;
  const pageRows = filtered.slice(pageStart, pageStart + rowsPerPage);

  const setFilter = (key: string, value: string) => {
    setFilterValues((current) => ({ ...current, [key]: value }));
    setPage(0);
  };

  return (
    <Paper
      className="surface nadaa-datatable"
      elevation={0}
      sx={{ overflow: "hidden" }}
    >
      <Stack
        className="nadaa-datatable__toolbar"
        direction={{ xs: "column", md: "row" }}
        spacing={1.5}
        sx={{
          alignItems: { xs: "stretch", md: "center" },
          p: 2,
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
            onChange={(event) => setFilter(filter.key, event.target.value)}
            select
            size="small"
            sx={{ minWidth: 160 }}
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
        <Table className="nadaa-dt-table" size="small" stickyHeader>
          <TableHead>
            <TableRow>
              {columns.map((column) => (
                <TableCell align={column.align} key={column.key}>
                  {column.label}
                </TableCell>
              ))}
            </TableRow>
          </TableHead>
          <TableBody>
            {pageRows.length === 0 ? (
              <TableRow>
                <TableCell colSpan={columns.length} sx={{ borderBottom: 0 }}>
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
                <TableRow
                  hover
                  key={getRowKey(row)}
                  onClick={onRowClick ? () => onRowClick(row) : undefined}
                  onKeyDown={
                    onRowClick
                      ? (event) => {
                          if (event.key === "Enter" || event.key === " ") {
                            event.preventDefault();
                            onRowClick(row);
                          }
                        }
                      : undefined
                  }
                  role={onRowClick ? "button" : undefined}
                  tabIndex={onRowClick ? 0 : undefined}
                  sx={onRowClick ? { cursor: "pointer" } : undefined}
                >
                  {columns.map((column) => (
                    <TableCell align={column.align} key={column.key}>
                      {column.render
                        ? column.render(row)
                        : String(
                            (row as Record<string, unknown>)[column.key] ?? "",
                          )}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </TableContainer>
      <TablePagination
        className="nadaa-datatable__pagination"
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
    </Paper>
  );
}
