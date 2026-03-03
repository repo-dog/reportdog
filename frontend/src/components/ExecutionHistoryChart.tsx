import React, { useState } from 'react';
import { formatDateShort } from '../utils/formatDate';
import {
  Card,
  Typography,
  Skeleton,
  Box,
  IconButton,
  Collapse,
  Stack,
} from '@mui/material';
import { ExpandMore, ExpandLess, Timeline } from '@mui/icons-material';
import {
  ResponsiveContainer,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
} from 'recharts';
import { useQuery } from '@tanstack/react-query';
import { getExecutionHistory } from '../api/client';

interface Props {
  executionName: string;
}

const ExecutionHistoryChart: React.FC<Props> = ({ executionName }) => {
  const [expanded, setExpanded] = useState(false);
  const { data, isLoading } = useQuery({
    queryKey: ['execution-history', executionName],
    queryFn: () => getExecutionHistory(executionName, 30),
    enabled: !!executionName,
  });

  if (isLoading) return <Skeleton variant="rounded" height={48} />;
  if (!data || data.length <= 1) return null;

  const chartData = [...data].reverse().map((d) => ({
    date: formatDateShort(d.uploaded_at),
    passed: d.passed,
    failed: d.failed,
    skipped: d.skipped,
    duration: parseFloat(d.duration_sec.toFixed(1)),
  }));

  return (
    <Card>
      <Box
        sx={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          px: 2,
          py: 0.75,
          cursor: 'pointer',
          userSelect: 'none',
          '&:hover': { bgcolor: 'action.hover' },
        }}
        onClick={() => setExpanded(!expanded)}
      >
        <Stack direction="row" spacing={1} alignItems="center">
          <Timeline fontSize="small" color="action" />
          <Typography variant="subtitle2" fontWeight={600}>
            Execution History
          </Typography>
          <Typography variant="caption" color="text.secondary">
            ({data.length} runs)
          </Typography>
        </Stack>
        <IconButton size="small" sx={{ p: 0.25 }}>
          {expanded ? <ExpandLess fontSize="small" /> : <ExpandMore fontSize="small" />}
        </IconButton>
      </Box>
      <Collapse in={expanded}>
        <Box sx={{ px: 2, pb: 1.5, pt: 0.5 }}>
          <Box sx={{ width: '100%', height: 160 }}>
            <ResponsiveContainer>
              <AreaChart data={chartData} margin={{ top: 5, right: 5, left: -10, bottom: 0 }}>
                <CartesianGrid strokeDasharray="3 3" opacity={0.3} />
                <XAxis dataKey="date" tick={{ fontSize: 10 }} />
                <YAxis tick={{ fontSize: 10 }} />
                <Tooltip contentStyle={{ fontSize: 12 }} />
                <Legend iconSize={10} wrapperStyle={{ fontSize: 11 }} />
                <Area type="monotone" dataKey="passed" stackId="1" stroke="#2e7d32" fill="#4caf50" fillOpacity={0.6} />
                <Area type="monotone" dataKey="failed" stackId="1" stroke="#d32f2f" fill="#f44336" fillOpacity={0.6} />
                <Area type="monotone" dataKey="skipped" stackId="1" stroke="#9e9e9e" fill="#bdbdbd" fillOpacity={0.4} />
              </AreaChart>
            </ResponsiveContainer>
          </Box>
        </Box>
      </Collapse>
    </Card>
  );
};

export default ExecutionHistoryChart;
