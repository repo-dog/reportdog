import React from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Grid,
  Button,
  Skeleton,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Chip,
  Stack,
} from '@mui/material';
import {
  Assessment,
  TrendingUp,
  Error as ErrorIcon,
  ArrowForward,
} from '@mui/icons-material';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { getStats, listReports } from '../api/client';
import TestResultBar from '../components/TestResultBar';

const StatCard: React.FC<{
  title: string;
  value: string | number;
  icon: React.ReactNode;
  color: string;
}> = ({ title, value, icon, color }) => (
  <Card>
    <CardContent sx={{ py: 1.5, px: 2, '&:last-child': { pb: 1.5 } }}>
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
        <Box
          sx={{
            bgcolor: color,
            borderRadius: 1.5,
            p: 0.75,
            display: 'flex',
            color: 'white',
          }}
        >
          {icon}
        </Box>
        <Box>
          <Typography variant="caption" color="text.secondary">
            {title}
          </Typography>
          <Typography variant="h6" fontWeight={700} lineHeight={1.2}>
            {value}
          </Typography>
        </Box>
      </Box>
    </CardContent>
  </Card>
);

const Home: React.FC = () => {
  const navigate = useNavigate();
  const { data: stats, isLoading } = useQuery({
    queryKey: ['stats'],
    queryFn: getStats,
  });

  const { data: recentData, isLoading: recentLoading } = useQuery({
    queryKey: ['reports', 'recent-10'],
    queryFn: () => listReports({ page: '1', page_size: '10' }),
  });

  return (
    <Box>
      <Box sx={{ mb: 3 }}>
        <Typography variant="h5" fontWeight={700}>
          Dashboard
        </Typography>
        <Typography variant="body2" color="text.secondary">
          Test execution overview for the last 7 days
        </Typography>
      </Box>

      <Grid container spacing={2} sx={{ mb: 3 }}>
        {isLoading ? (
          Array.from({ length: 4 }).map((_, i) => (
            <Grid size={{ xs: 6, md: 3}} key={i}>
              <Skeleton variant="rounded" height={80} />
            </Grid>
          ))
        ) : (
          <>
            <Grid size={{ xs: 6, md: 3}}>
              <StatCard
                title="Total Reports"
                value={stats?.total_reports ?? 0}
                icon={<Assessment fontSize="small" />}
                color="#1976d2"
              />
            </Grid>
            <Grid size={{ xs: 6, md: 3}}>
              <StatCard
                title="Last 7 Days"
                value={stats?.reports_last_7_days ?? 0}
                icon={<TrendingUp fontSize="small" />}
                color="#2e7d32"
              />
            </Grid>
            <Grid size={{ xs: 6, md: 3}}>
              <StatCard
                title="Pass Rate"
                value={`${(stats?.overall_pass_rate ?? 0).toFixed(1)}%`}
                icon={<TrendingUp fontSize="small" />}
                color="#ed6c02"
              />
            </Grid>
            <Grid size={{ xs: 6, md: 3}}>
              <StatCard
                title="Failed (7d)"
                value={stats?.total_failed_last_7_days ?? 0}
                icon={<ErrorIcon fontSize="small" />}
                color="#d32f2f"
              />
            </Grid>
          </>
        )}
      </Grid>

      {/* Recent Runs */}
      <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 1.5 }}>
        <Typography variant="h6" fontWeight={600}>
          Recent Runs
        </Typography>
        <Button
          size="small"
          endIcon={<ArrowForward fontSize="small" />}
          onClick={() => navigate('/reports')}
        >
          View all
        </Button>
      </Box>
      <Card>
        <TableContainer>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>Execution</TableCell>
                <TableCell>Uploaded</TableCell>
                <TableCell sx={{ minWidth: 160 }}>Results</TableCell>
                <TableCell align="center">Duration</TableCell>
                <TableCell>Tags</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {recentLoading
                ? Array.from({ length: 5 }).map((_, i) => (
                    <TableRow key={i}>
                      {Array.from({ length: 5 }).map((_, j) => (
                        <TableCell key={j}>
                          <Skeleton />
                        </TableCell>
                      ))}
                    </TableRow>
                  ))
                : recentData?.reports?.map((r) => {
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
                          {r.tags?.slice(0, 2).map((t) => (
                            <Chip
                              key={`${t.key}:${t.value}`}
                              label={`${t.key}:${t.value}`}
                              size="small"
                              variant="outlined"
                              clickable
                              onClick={(e) => {
                                e.stopPropagation();
                                navigate(`/reports?tags=${encodeURIComponent(t.key)}:${encodeURIComponent(t.value)}`);
                              }}
                              sx={{ cursor: 'pointer', '&:hover': { borderColor: 'primary.main' } }}
                            />
                          ))}
                          {(r.tags?.length ?? 0) > 2 && (
                            <Chip label={`+${(r.tags?.length ?? 0) - 2}`} size="small" />
                          )}
                        </Stack>
                      </TableCell>
                    </TableRow>
                    );
                  })}
              {!recentLoading && (!recentData?.reports || recentData.reports.length === 0) && (
                <TableRow>
                  <TableCell colSpan={5} align="center" sx={{ py: 4 }}>
                    <Typography color="text.secondary">No reports yet</Typography>
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </TableContainer>
      </Card>
    </Box>
  );
};

export default Home;
