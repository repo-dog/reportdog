import React from 'react';
import { formatDate } from '../utils/formatDate';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
  IconButton,
  Skeleton,
  Box,
} from '@mui/material';
import { Close } from '@mui/icons-material';
import { useQuery } from '@tanstack/react-query';
import { getTestHistory } from '../api/client';
import StatusChip from './StatusChip';

interface Props {
  open: boolean;
  onClose: () => void;
  executionName: string;
  testName: string;
}

const TestHistoryDialog: React.FC<Props> = ({ open, onClose, executionName, testName }) => {
  const { data, isLoading } = useQuery({
    queryKey: ['test-history', executionName, testName],
    queryFn: () => getTestHistory(executionName, testName, 50),
    enabled: open && !!executionName && !!testName,
  });

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
      <DialogTitle sx={{ display: 'flex', alignItems: 'center', py: 1.5, pr: 1 }}>
        <Box sx={{ flex: 1, minWidth: 0 }}>
          <Typography variant="subtitle2" fontWeight={600} noWrap>
            Test History
          </Typography>
          <Typography variant="caption" color="text.secondary" noWrap display="block">
            {testName}
          </Typography>
        </Box>
        <IconButton onClick={onClose} size="small">
          <Close fontSize="small" />
        </IconButton>
      </DialogTitle>
      <DialogContent sx={{ px: 2, pb: 2 }}>
        <TableContainer>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>Date</TableCell>
                <TableCell>Status</TableCell>
                <TableCell align="right">Duration</TableCell>
                <TableCell>Failure</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {isLoading
                ? Array.from({ length: 5 }).map((_, i) => (
                    <TableRow key={i}>
                      {Array.from({ length: 4 }).map((_, j) => (
                        <TableCell key={j}><Skeleton /></TableCell>
                      ))}
                    </TableRow>
                  ))
                : data?.map((item, i) => (
                    <TableRow key={i}>
                      <TableCell>
                        <Typography variant="caption">
                          {formatDate(item.uploaded_at)}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <StatusChip status={item.status} />
                      </TableCell>
                      <TableCell align="right">
                        <Typography variant="caption">{item.duration_sec.toFixed(2)}s</Typography>
                      </TableCell>
                      <TableCell>
                        <Typography variant="caption" color="text.secondary" noWrap sx={{ maxWidth: 300, display: 'block' }}>
                          {item.failure_msg || '—'}
                        </Typography>
                      </TableCell>
                    </TableRow>
                  ))}
              {!isLoading && (!data || data.length === 0) && (
                <TableRow>
                  <TableCell colSpan={4} align="center" sx={{ py: 3 }}>
                    <Typography variant="body2" color="text.secondary">
                      No history found
                    </Typography>
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </TableContainer>
      </DialogContent>
    </Dialog>
  );
};

export default TestHistoryDialog;
