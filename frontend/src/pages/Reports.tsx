import React, { useState, useMemo, useEffect, useRef, useCallback } from 'react';
import {
  Box,
  Card,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
  TextField,
  Chip,
  TablePagination,
  Skeleton,
  IconButton,
  Stack,
  InputAdornment,
  Autocomplete,
  LinearProgress,
  CircularProgress,
  Fade,
} from '@mui/material';
import { Visibility, Search as SearchIcon, CheckCircleOutline, ErrorOutline, ListAlt, AddCircleOutline } from '@mui/icons-material';
import { useQuery, keepPreviousData } from '@tanstack/react-query';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { listReports, listTags } from '../api/client';
import type { TagInfo } from '../api/client';
import TestResultBar from '../components/TestResultBar';

const Reports: React.FC = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  // Free-text search — local input + debounced API query
  const [searchText, setSearchText] = useState(searchParams.get('q') || '');
  const [debouncedSearch, setDebouncedSearch] = useState(searchText);
  const debounceTimer = useRef<ReturnType<typeof setTimeout>>(undefined);

  const handleSearchChange = useCallback((value: string) => {
    setSearchText(value);
    clearTimeout(debounceTimer.current);
    debounceTimer.current = setTimeout(() => {
      setDebouncedSearch(value);
      // Reset to page 1 and update URL
      const newParams = new URLSearchParams(searchParams);
      if (value.trim()) newParams.set('q', value.trim());
      else newParams.delete('q');
      newParams.set('page', '1');
      navigate(`/reports?${newParams.toString()}`, { replace: true });
    }, 350);
  }, [searchParams, navigate]);

  const page = parseInt(searchParams.get('page') || '1');
  const pageSize = parseInt(searchParams.get('page_size') || '20');
  const executionName = searchParams.get('execution_name') || '';
  const status = searchParams.get('status') || 'all';

  // Parse tags from URL: format is "key1:value1,key2:value2"
  const tagsParam = searchParams.get('tags') || '';
  // Also support legacy tag_key/tag_value params for backwards compat
  const legacyTagKey = searchParams.get('tag_key') || '';
  const legacyTagValue = searchParams.get('tag_value') || '';

  const parsedUrlTags = useMemo<TagInfo[]>(() => {
    if (tagsParam) {
      return tagsParam.split(',').filter(Boolean).map((s) => {
        const idx = s.indexOf(':');
        return idx > 0
          ? { key: s.slice(0, idx), value: s.slice(idx + 1), count: 0 }
          : { key: s, value: '', count: 0 };
      }).filter((t) => t.key && t.value);
    }
    if (legacyTagKey && legacyTagValue) {
      return [{ key: legacyTagKey, value: legacyTagValue, count: 0 }];
    }
    return [];
  }, [tagsParam, legacyTagKey, legacyTagValue]);

  // Serialize tags array to URL param string
  const encodeTagsParam = (tags: TagInfo[]) =>
    tags.map((t) => `${t.key}:${t.value}`).join(',');

  // Tag selection — real state so Autocomplete can control it synchronously
  const [selectedTags, setSelectedTags] = useState<TagInfo[]>(() => parsedUrlTags);

  // Helper to update URL with current tags
  const setTagsInUrl = useCallback((tags: TagInfo[], baseParams?: URLSearchParams) => {
    const newParams = new URLSearchParams(baseParams ?? searchParams);
    // Clean up legacy params
    newParams.delete('tag_key');
    newParams.delete('tag_value');
    if (tags.length > 0) {
      newParams.set('tags', encodeTagsParam(tags));
    } else {
      newParams.delete('tags');
    }
    newParams.set('page', '1');
    return newParams;
  }, [searchParams]);

  // Sync URL → state when navigating from external sources (e.g. tag clicks in table)
  const skipUrlSync = useRef(false);
  useEffect(() => {
    if (skipUrlSync.current) {
      skipUrlSync.current = false;
      return;
    }
    const isDifferent =
      parsedUrlTags.length !== selectedTags.length ||
      parsedUrlTags.some((ut, i) => ut.key !== selectedTags[i]?.key || ut.value !== selectedTags[i]?.value);
    if (isDifferent) {
      setSelectedTags(parsedUrlTags);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [parsedUrlTags]);

  // Derive query params from selectedTags state (always up-to-date)
  const params: Record<string, string> = {
    page: String(page),
    page_size: String(pageSize),
  };
  if (executionName) params.execution_name = executionName;
  if (status !== 'all') params.status = status;
  if (selectedTags.length > 0) {
    params.tag_key = selectedTags[0].key;
    params.tag_value = selectedTags[0].value;
  }
  if (debouncedSearch.trim()) params.q = debouncedSearch.trim();

  const { data, isLoading, isFetching } = useQuery({
    queryKey: ['reports', params],
    queryFn: () => listReports(params),
    placeholderData: keepPreviousData,
  });

  // True while user is still typing (debounce hasn't fired yet)
  const isTyping = searchText.trim() !== debouncedSearch.trim();
  // Show loading state when fetching new data or waiting for debounce
  const isSearching = isFetching || isTyping;

  // Load available tags for autocomplete
  const { data: allTags } = useQuery({
    queryKey: ['tags'],
    queryFn: () => listTags(),
  });

  const updateParam = (key: string, value: string) => {
    const newParams = new URLSearchParams(searchParams);
    if (value) newParams.set(key, value);
    else newParams.delete(key);
    if (key !== 'page') newParams.set('page', '1');
    navigate(`/reports?${newParams.toString()}`, { replace: true });
  };

  // Client-side filter for additional tags beyond the first (API handles one tag + search)
  const filteredReports = useMemo(() => {
    if (!data?.reports) return [];
    if (selectedTags.length <= 1) return data.reports;

    const extraTags = selectedTags.slice(1);
    return data.reports.filter((r) =>
      extraTags.every((st) =>
        r.tags?.some((t) => t.key === st.key && t.value === st.value)
      )
    );
  }, [data?.reports, selectedTags]);

  return (
    <Box>
      <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2 }}>
        <Typography variant="h5" fontWeight={700}>
          Test Reports
        </Typography>
      </Box>

      {/* Search bar + status chips + tag chips */}
      <Card sx={{ p: 1.5, mb: 2 }}>
        <Stack spacing={1.5}>
          <Stack direction="row" spacing={1} alignItems="center">
            <TextField
              placeholder="Search by name, execution, or tag…"
              size="small"
              value={searchText}
              onChange={(e) => handleSearchChange(e.target.value)}
              sx={{ flex: 1 }}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    {isSearching ? (
                      <CircularProgress size={16} sx={{ color: 'primary.main' }} />
                    ) : (
                      <SearchIcon fontSize="small" color="action" />
                    )}
                  </InputAdornment>
                ),
              }}
            />
            <Stack direction="row" spacing={0.5} sx={{ flexShrink: 0 }}>
              {[
                { value: 'all', label: 'All', icon: <ListAlt sx={{ fontSize: 16 }} />, color: 'default' as const },
                { value: 'passed', label: 'Passed', icon: <CheckCircleOutline sx={{ fontSize: 16 }} />, color: 'success' as const },
                { value: 'failed', label: 'Failed', icon: <ErrorOutline sx={{ fontSize: 16 }} />, color: 'error' as const },
              ].map((s) => {
                const isActive = status === s.value;
                return (
                  <Chip
                    key={s.value}
                    icon={s.icon}
                    label={s.label}
                    size="small"
                    variant={isActive ? 'filled' : 'outlined'}
                    color={s.color}
                    clickable
                    onClick={() => updateParam('status', s.value === 'all' ? '' : s.value)}
                    sx={{
                      fontWeight: isActive ? 700 : 500,
                      px: 0.5,
                      ...(isActive && s.value === 'all' && {
                        bgcolor: 'action.selected',
                        borderColor: 'divider',
                      }),
                    }}
                  />
                );
              })}
            </Stack>
          </Stack>
          <Autocomplete
            multiple
            size="small"
            options={(allTags ?? []).filter(
              (opt) => !selectedTags.some((st) => st.key === opt.key && st.value === opt.value)
            )}
            filterSelectedOptions
            getOptionLabel={(opt) => `${opt.key}:${opt.value}`}
            isOptionEqualToValue={(opt, val) => opt.key === val.key && opt.value === val.value}
            value={selectedTags}
            onChange={(_, newValue) => {
              // Update state synchronously so Autocomplete works immediately
              setSelectedTags(newValue);
              // Also update URL for bookmarkability; skip the URL→state sync
              skipUrlSync.current = true;
              const newParams = setTagsInUrl(newValue);
              navigate(`/reports?${newParams.toString()}`, { replace: true });
            }}
            renderTags={(value, getTagProps) =>
              value.map((tag, index) => {
                const { key, ...chipProps } = getTagProps({ index });
                return (
                  <Chip
                    key={key}
                    label={`${tag.key}:${tag.value}`}
                    size="small"
                    variant="outlined"
                    color="primary"
                    {...chipProps}
                  />
                );
              })
            }
            renderInput={(params) => (
              <TextField {...params} placeholder="Filter by tags…" />
            )}
            noOptionsText="No tags found"
          />
        </Stack>
      </Card>

      <Card sx={{ position: 'relative', overflow: 'hidden' }}>
        <Fade in={isSearching && !isLoading} unmountOnExit>
          <LinearProgress
            sx={{
              position: 'absolute',
              top: 0,
              left: 0,
              right: 0,
              zIndex: 1,
              height: 2,
            }}
          />
        </Fade>
        <TableContainer>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>Execution</TableCell>
                <TableCell>Uploaded</TableCell>
                <TableCell sx={{ minWidth: 180 }}>Results</TableCell>
                <TableCell align="center">Duration</TableCell>
                <TableCell>Tags</TableCell>
                <TableCell align="center" width={48}></TableCell>
              </TableRow>
            </TableHead>
            <TableBody
              sx={{
                transition: 'opacity 0.2s ease',
                opacity: isSearching && !isLoading ? 0.5 : 1,
              }}
            >
              {isLoading
                ? Array.from({ length: pageSize > 10 ? 8 : 5 }).map((_, i) => (
                    <TableRow key={i}>
                      <TableCell><Skeleton variant="text" width="70%" /></TableCell>
                      <TableCell><Skeleton variant="text" width="80%" /></TableCell>
                      <TableCell><Skeleton variant="rounded" height={18} /></TableCell>
                      <TableCell align="center"><Skeleton variant="text" width={40} sx={{ mx: 'auto' }} /></TableCell>
                      <TableCell><Skeleton variant="rounded" width={80} height={22} /></TableCell>
                      <TableCell align="center"><Skeleton variant="circular" width={24} height={24} sx={{ mx: 'auto' }} /></TableCell>
                    </TableRow>
                  ))
                : filteredReports.map((r) => {
                    const hasFailed = r.failed > 0;
                    return (
                    <TableRow
                      key={r.id}
                      hover
                      sx={{
                        cursor: 'pointer',
                        borderLeft: 3,
                        borderLeftColor: hasFailed ? 'error.main' : 'success.main',
                        bgcolor: hasFailed
                          ? 'rgba(244,67,54,0.04)'
                          : 'rgba(76,175,80,0.04)',
                        '&:hover': {
                          bgcolor: hasFailed
                            ? 'rgba(244,67,54,0.08) !important'
                            : 'rgba(76,175,80,0.08) !important',
                        },
                      }}
                      onClick={() => navigate(`/reports/${r.id}`)}
                    >
                      <TableCell>
                        <Typography variant="body2" fontWeight={600}>
                          {r.execution_name}
                        </Typography>
                        {r.name && (
                          <Typography variant="caption" color="text.secondary">
                            {r.name}
                          </Typography>
                        )}
                      </TableCell>
                      <TableCell>
                        <Typography variant="caption">
                          {new Date(r.uploaded_at).toLocaleString()}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <TestResultBar
                          total={r.total_tests}
                          passed={r.passed}
                          failed={r.failed}
                          skipped={r.skipped}
                          showLabels
                        />
                      </TableCell>
                      <TableCell align="center">
                        <Typography variant="caption">{r.duration_sec.toFixed(1)}s</Typography>
                      </TableCell>
                      <TableCell>
                        <Stack direction="row" spacing={0.25} flexWrap="wrap" useFlexGap>
                          {r.tags?.slice(0, 3).map((t) => {
                            const alreadySelected = selectedTags.some(
                              (st) => st.key === t.key && st.value === t.value
                            );
                            return (
                              <Chip
                                key={`${t.key}:${t.value}`}
                                label={
                                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.25 }}>
                                    <span>{`${t.key}:${t.value}`}</span>
                                    {!alreadySelected && (
                                      <AddCircleOutline
                                        sx={{ fontSize: 14, ml: 0.25, color: 'primary.main', '&:hover': { color: 'primary.dark' } }}
                                      />
                                    )}
                                  </Box>
                                }
                                size="small"
                                variant="outlined"
                                clickable
                                onClick={(e) => {
                                  e.stopPropagation();
                                  if (alreadySelected) {
                                    // Already in filter — navigate to single-tag view
                                    navigate(`/reports?tags=${encodeURIComponent(t.key)}:${encodeURIComponent(t.value)}`);
                                  } else {
                                    // Append to existing selected tags
                                    const newTags = [...selectedTags, { key: t.key, value: t.value, count: 0 }];
                                    setSelectedTags(newTags);
                                    skipUrlSync.current = true;
                                    const newParams = setTagsInUrl(newTags);
                                    navigate(`/reports?${newParams.toString()}`, { replace: true });
                                  }
                                }}
                                sx={{ cursor: 'pointer', '&:hover': { borderColor: 'primary.main' } }}
                              />
                            );
                          })}
                          {(r.tags?.length ?? 0) > 3 && (
                            <Chip label={`+${(r.tags?.length ?? 0) - 3}`} size="small" />
                          )}
                        </Stack>
                      </TableCell>
                      <TableCell align="center">
                        <IconButton size="small">
                          <Visibility fontSize="small" />
                        </IconButton>
                      </TableCell>
                    </TableRow>
                    );
                  })}
              {!isLoading && filteredReports.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6} align="center" sx={{ py: 4 }}>
                    <Typography color="text.secondary">No reports found</Typography>
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </TableContainer>
        {data && data.total > 0 && (
          <TablePagination
            component="div"
            count={data.total}
            page={page - 1}
            onPageChange={(_, p) => updateParam('page', String(p + 1))}
            rowsPerPage={pageSize}
            onRowsPerPageChange={(e) => updateParam('page_size', e.target.value)}
            rowsPerPageOptions={[10, 20, 50]}
          />
        )}
      </Card>
    </Box>
  );
};

export default Reports;
