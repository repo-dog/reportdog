import React, { useState } from 'react';
import { formatDate } from '../utils/formatDate';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Accordion,
  AccordionSummary,
  AccordionDetails,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Chip,
  IconButton,
  Button,
  Skeleton,
  Alert,
  Tooltip,
  Stack,
  Collapse,
  TextField,
  InputAdornment,
} from '@mui/material';
import {
  ExpandMore,
  History,
  ArrowBack,
  KeyboardArrowDown,
  KeyboardArrowRight,
  Search as SearchIcon,
} from '@mui/icons-material';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getReport, addTags, removeTag } from '../api/client';
import StatusChip from '../components/StatusChip';
import TagChips from '../components/TagChips';
import ExecutionHistoryChart from '../components/ExecutionHistoryChart';
import TestHistoryDialog from '../components/TestHistoryDialog';

const ReportDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [historyTest, setHistoryTest] = useState<string | null>(null);
  const [expandedCase, setExpandedCase] = useState<string | null>(null);
  const [caseSearch, setCaseSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState<Set<string>>(new Set());

  const { data: report, isLoading, error } = useQuery({
    queryKey: ['report', id],
    queryFn: () => getReport(id!),
    enabled: !!id,
  });

  const addTagMutation = useMutation({
    mutationFn: ({ key, value }: { key: string; value: string }) =>
      addTags(id!, [{ key, value }]),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['report', id] }),
  });

  const removeTagMutation = useMutation({
    mutationFn: ({ key, value }: { key: string; value: string }) => removeTag(id!, key, value),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['report', id] }),
  });

  if (isLoading) {
    return (
      <Box>
        <Skeleton variant="rounded" height={120} sx={{ mb: 2 }} />
        <Skeleton variant="rounded" height={200} />
      </Box>
    );
  }

  if (error || !report) {
    return <Alert severity="error">Report not found</Alert>;
  }

  const passRate = report.total_tests > 0
    ? ((report.passed / report.total_tests) * 100).toFixed(1)
    : '0';

  return (
    <Box>
      <Button
        startIcon={<ArrowBack />}
        onClick={() => navigate(-1)}
        sx={{ mb: 1 }}
        size="small"
      >
        Back to Reports
      </Button>

      {/* Summary */}
      <Card sx={{ mb: 2 }}>
        <CardContent sx={{ py: 1.5, px: 2 }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: 1 }}>
            <Box>
              <Typography variant="h6" fontWeight={700}>
                {report.execution_name}
              </Typography>
              {report.name && (
                <Typography variant="body2" color="text.secondary">
                  {report.name}
                </Typography>
              )}
              <Typography variant="caption" color="text.secondary">
                {formatDate(report.timestamp || report.uploaded_at)} · {report.source}
              </Typography>
            </Box>
            <Stack direction="row" spacing={1} alignItems="center">
              <Chip label={`${report.total_tests} tests`} size="small" />
              <Chip label={`${report.passed} passed`} size="small" color="success" variant="outlined" />
              {report.failed > 0 && (
                <Chip label={`${report.failed} failed`} size="small" color="error" variant="outlined" />
              )}
              {report.skipped > 0 && (
                <Chip label={`${report.skipped} skipped`} size="small" variant="outlined" />
              )}
              <Chip label={`${passRate}%`} size="small" color={parseFloat(passRate) < 100 ? 'warning' : 'success'} />
              <Chip label={`${report.duration_sec.toFixed(1)}s`} size="small" variant="outlined" />
            </Stack>
          </Box>
          <Box sx={{ mt: 1 }}>
            <TagChips
              tags={report.tags || []}
              editable
              onAdd={(key, value) => addTagMutation.mutate({ key, value })}
              onRemove={(key, value) => removeTagMutation.mutate({ key, value })}
              onClick={(key, value) => navigate(`/reports?tags=${encodeURIComponent(key)}:${encodeURIComponent(value)}`)}
            />
          </Box>
        </CardContent>
      </Card>

      {/* Execution History Chart */}
      <Box sx={{ mb: 2 }}>
        <ExecutionHistoryChart executionName={report.execution_name} />
      </Box>

      {/* Test case search & filters */}
      <Card sx={{ p: 1.5, mb: 1.5 }}>
        <Stack spacing={1}>
          <TextField
            placeholder="Search test cases by name or class…"
            size="small"
            fullWidth
            value={caseSearch}
            onChange={(e) => setCaseSearch(e.target.value)}
            InputProps={{
              startAdornment: (
                <InputAdornment position="start">
                  <SearchIcon fontSize="small" color="action" />
                </InputAdornment>
              ),
            }}
          />
          <Stack direction="row" spacing={0.75} alignItems="center">
            <Typography variant="caption" color="text.secondary" sx={{ mr: 0.5 }}>
              Status:
            </Typography>
            {(['passed', 'failed', 'error', 'skipped'] as const).map((s) => {
              const active = statusFilter.has(s);
              const colorMap: Record<string, 'success' | 'error' | 'warning' | 'default'> = {
                passed: 'success',
                failed: 'error',
                error: 'warning',
                skipped: 'default',
              };
              return (
                <Chip
                  key={s}
                  label={s.charAt(0).toUpperCase() + s.slice(1)}
                  size="small"
                  variant={active ? 'filled' : 'outlined'}
                  color={colorMap[s]}
                  onClick={() => {
                    const next = new Set(statusFilter);
                    if (next.has(s)) next.delete(s);
                    else next.add(s);
                    setStatusFilter(next);
                  }}
                  sx={{ cursor: 'pointer' }}
                />
              );
            })}
            {(caseSearch || statusFilter.size > 0) && (
              <Chip
                label="Clear"
                size="small"
                variant="outlined"
                onDelete={() => {
                  setCaseSearch('');
                  setStatusFilter(new Set());
                }}
                sx={{ ml: 0.5 }}
              />
            )}
          </Stack>
        </Stack>
      </Card>

      {/* Suites */}
      <Typography variant="subtitle1" fontWeight={600} sx={{ mb: 1 }}>
        Test Suites ({report.suites?.length || 0})
      </Typography>

      {report.suites?.map((suite) => (
        <Accordion
          key={suite.id}
          disableGutters
          sx={{ mb: 0.5 }}
          defaultExpanded={report.suites?.length === 1}
        >
          <AccordionSummary expandIcon={<ExpandMore />} sx={{ minHeight: 40, '& .MuiAccordionSummary-content': { my: 0.5 } }}>
            <Stack direction="row" spacing={1} alignItems="center" sx={{ width: '100%', pr: 1 }}>
              <Typography variant="body2" fontWeight={600} sx={{ flex: 1 }}>
                {suite.name}
              </Typography>
              <Chip label={`${suite.total_tests}`} size="small" />
              <Chip label={`${suite.passed} ✓`} size="small" color="success" variant="outlined" />
              {suite.failed > 0 && (
                <Chip label={`${suite.failed} ✗`} size="small" color="error" variant="outlined" />
              )}
              <Typography variant="caption" color="text.secondary">
                {suite.duration_sec.toFixed(1)}s
              </Typography>
            </Stack>
          </AccordionSummary>
          <AccordionDetails sx={{ p: 0 }}>
            <TableContainer>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell width={32}></TableCell>
                    <TableCell>Test Case</TableCell>
                    <TableCell>Class</TableCell>
                    <TableCell align="center">Status</TableCell>
                    <TableCell align="right">Duration</TableCell>
                    <TableCell align="center" width={40}>
                      <Tooltip title="View test history">
                        <History fontSize="small" color="action" />
                      </Tooltip>
                    </TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {suite.cases?.filter((tc) => {
                    const q = caseSearch.trim().toLowerCase();
                    if (q && !tc.name.toLowerCase().includes(q) && !(tc.classname?.toLowerCase().includes(q))) return false;
                    if (statusFilter.size > 0 && !statusFilter.has(tc.status)) return false;
                    return true;
                  }).map((tc) => {
                    const statusColor =
                      tc.status === 'failed' || tc.status === 'error'
                        ? { border: 'error.main', bg: 'rgba(244,67,54,0.04)', hover: 'rgba(244,67,54,0.08) !important' }
                        : tc.status === 'skipped'
                        ? { border: 'grey.400', bg: 'rgba(158,158,158,0.04)', hover: 'rgba(158,158,158,0.08) !important' }
                        : { border: 'success.main', bg: 'rgba(76,175,80,0.04)', hover: 'rgba(76,175,80,0.08) !important' };
                    return (
                    <React.Fragment key={tc.id}>
                      <TableRow
                        hover
                        sx={{
                          borderLeft: 3,
                          borderLeftColor: statusColor.border,
                          bgcolor: statusColor.bg,
                          '&:hover': { bgcolor: statusColor.hover },
                        }}
                      >
                        <TableCell sx={{ p: 0.5 }}>
                          {(tc.failure_text?.trim() || tc.system_out?.trim() || tc.system_err?.trim()) && (
                            <IconButton
                              size="small"
                              onClick={() => setExpandedCase(expandedCase === tc.id ? null : tc.id)}
                            >
                              {expandedCase === tc.id
                                ? <KeyboardArrowDown fontSize="small" />
                                : <KeyboardArrowRight fontSize="small" />}
                            </IconButton>
                          )}
                        </TableCell>
                        <TableCell>
                          <Typography variant="body2">{tc.name}</Typography>
                          {tc.failure_msg && (
                            <Typography variant="caption" color="error" display="block" noWrap sx={{ maxWidth: 400 }}>
                              {tc.failure_msg}
                            </Typography>
                          )}
                        </TableCell>
                        <TableCell>
                          <Typography variant="caption" color="text.secondary">
                            {tc.classname || '—'}
                          </Typography>
                        </TableCell>
                        <TableCell align="center">
                          <StatusChip status={tc.status} />
                        </TableCell>
                        <TableCell align="right">
                          <Typography variant="caption">{tc.duration_sec.toFixed(2)}s</Typography>
                        </TableCell>
                        <TableCell align="center">
                          <IconButton
                            size="small"
                            onClick={(e) => {
                              e.stopPropagation();
                              setHistoryTest(tc.name);
                            }}
                          >
                            <History fontSize="small" />
                          </IconButton>
                        </TableCell>
                      </TableRow>
                      {(tc.failure_text?.trim() || tc.system_out?.trim() || tc.system_err?.trim()) && (
                        <TableRow>
                          <TableCell colSpan={6} sx={{ p: 0, border: 0 }}>
                            <Collapse in={expandedCase === tc.id}>
                              <Box sx={{ p: 1.5, bgcolor: 'action.hover' }}>
                                {tc.failure_text && (
                                  <Box sx={{ mb: 1 }}>
                                    <Typography variant="caption" fontWeight={600}>Stack Trace</Typography>
                                    <Box
                                      component="pre"
                                      sx={{
                                        fontSize: '0.7rem',
                                        overflow: 'auto',
                                        maxHeight: 200,
                                        bgcolor: 'background.paper',
                                        p: 1,
                                        borderRadius: 1,
                                        m: 0,
                                        mt: 0.5,
                                      }}
                                    >
                                      {tc.failure_text}
                                    </Box>
                                  </Box>
                                )}
                                {tc.system_out && (
                                  <Box sx={{ mb: 1 }}>
                                    <Typography variant="caption" fontWeight={600}>System Out</Typography>
                                    <Box
                                      component="pre"
                                      sx={{ fontSize: '0.7rem', overflow: 'auto', maxHeight: 150, bgcolor: 'background.paper', p: 1, borderRadius: 1, m: 0, mt: 0.5 }}
                                    >
                                      {tc.system_out}
                                    </Box>
                                  </Box>
                                )}
                                {tc.system_err && (
                                  <Box>
                                    <Typography variant="caption" fontWeight={600}>System Err</Typography>
                                    <Box
                                      component="pre"
                                      sx={{ fontSize: '0.7rem', overflow: 'auto', maxHeight: 150, bgcolor: 'background.paper', p: 1, borderRadius: 1, m: 0, mt: 0.5 }}
                                    >
                                      {tc.system_err}
                                    </Box>
                                  </Box>
                                )}
                              </Box>
                            </Collapse>
                          </TableCell>
                        </TableRow>
                      )}
                    </React.Fragment>
                    );
                  })}
                </TableBody>
              </Table>
            </TableContainer>
          </AccordionDetails>
        </Accordion>
      ))}

      {/* Test History Dialog */}
      <TestHistoryDialog
        open={!!historyTest}
        onClose={() => setHistoryTest(null)}
        executionName={report.execution_name}
        testName={historyTest || ''}
      />
    </Box>
  );
};

export default ReportDetail;
