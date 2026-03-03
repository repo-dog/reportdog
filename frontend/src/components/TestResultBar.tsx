import React from 'react';
import { Box, Tooltip, Typography } from '@mui/material';

interface TestResultBarProps {
  total: number;
  passed: number;
  failed: number;
  skipped: number;
  /** Height in px, default 8 */
  height?: number;
  /** Show numeric labels, default false */
  showLabels?: boolean;
}

const COLORS = {
  passed: '#4caf50',
  failed: '#f44336',
  error: '#ff9800',
  skipped: '#bdbdbd',
};

const TestResultBar: React.FC<TestResultBarProps> = ({
  total,
  passed,
  failed,
  skipped,
  height = 8,
  showLabels = false,
}) => {
  if (total === 0) {
    return (
      <Box
        sx={{
          width: '100%',
          height,
          borderRadius: height / 2,
          bgcolor: 'action.disabledBackground',
        }}
      />
    );
  }

  const error = Math.max(0, total - passed - failed - skipped);
  const segments = [
    { key: 'passed', count: passed, color: COLORS.passed, label: 'Passed' },
    { key: 'failed', count: failed, color: COLORS.failed, label: 'Failed' },
    { key: 'error', count: error, color: COLORS.error, label: 'Error' },
    { key: 'skipped', count: skipped, color: COLORS.skipped, label: 'Skipped' },
  ].filter((s) => s.count > 0);

  return (
    <Box>
      <Tooltip
        title={
          <Box sx={{ textAlign: 'left' }}>
            {segments.map((s) => (
              <Typography key={s.key} variant="caption" display="block">
                <Box
                  component="span"
                  sx={{
                    display: 'inline-block',
                    width: 8,
                    height: 8,
                    borderRadius: '50%',
                    bgcolor: s.color,
                    mr: 0.75,
                    verticalAlign: 'middle',
                  }}
                />
                {s.label}: {s.count} ({((s.count / total) * 100).toFixed(0)}%)
              </Typography>
            ))}
          </Box>
        }
        arrow
        placement="top"
      >
        <Box
          sx={{
            display: 'flex',
            width: '100%',
            height,
            borderRadius: height / 2,
            overflow: 'hidden',
          }}
        >
          {segments.map((s) => (
            <Box
              key={s.key}
              sx={{
                width: `${(s.count / total) * 100}%`,
                bgcolor: s.color,
                transition: 'width 0.3s ease',
                minWidth: s.count > 0 ? 2 : 0,
              }}
            />
          ))}
        </Box>
      </Tooltip>
      {showLabels && (
        <Box sx={{ display: 'flex', gap: 1, mt: 0.5, flexWrap: 'wrap' }}>
          {segments.map((s) => (
            <Typography
              key={s.key}
              variant="caption"
              sx={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: 0.5,
                color: s.color,
                fontWeight: 600,
                fontSize: '0.68rem',
                lineHeight: 1,
              }}
            >
              <Box
                component="span"
                sx={{
                  display: 'inline-block',
                  width: 6,
                  height: 6,
                  borderRadius: '50%',
                  bgcolor: s.color,
                  flexShrink: 0,
                }}
              />
              {s.count} {s.label}
            </Typography>
          ))}
        </Box>
      )}
    </Box>
  );
};

export default TestResultBar;
