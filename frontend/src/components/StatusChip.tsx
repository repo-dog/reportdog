import React from 'react';
import { Chip } from '@mui/material';
import { CheckCircle, Cancel, Error, SkipNext } from '@mui/icons-material';

interface Props {
  status: string;
  size?: 'small' | 'medium';
}

const config: Record<string, { color: any; icon: React.ReactElement }> = {
  passed: { color: 'success', icon: <CheckCircle fontSize="small" /> },
  failed: { color: 'error', icon: <Cancel fontSize="small" /> },
  error: { color: 'warning', icon: <Error fontSize="small" /> },
  skipped: { color: 'default', icon: <SkipNext fontSize="small" /> },
};

const StatusChip: React.FC<Props> = ({ status, size = 'small' }) => {
  const cfg = config[status] || config.passed;
  return <Chip label={status} color={cfg.color} icon={cfg.icon} size={size} variant="outlined" />;
};

export default StatusChip;
