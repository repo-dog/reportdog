import React, { useState, useCallback } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  TextField,
  Button,
  Chip,
  Stack,
  Alert,
  Tab,
  Tabs,
  LinearProgress,
} from '@mui/material';
import { CloudUpload, ContentPaste, Send } from '@mui/icons-material';
import { useMutation } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { uploadReport, ingestRawXML } from '../api/client';
import { useAppConfig } from '../context/AppConfigContext';

const Upload: React.FC = () => {
  const navigate = useNavigate();
  const { manualUploadEnabled } = useAppConfig();
  const [tab, setTab] = useState(0);
  const [file, setFile] = useState<File | null>(null);
  const [xmlText, setXmlText] = useState('');
  const [executionName, setExecutionName] = useState('');
  const [reportName, setReportName] = useState('');
  const [tagInput, setTagInput] = useState('');
  const [tags, setTags] = useState<{ key: string; value: string }[]>([]);
  const [error, setError] = useState('');
  const [dragOver, setDragOver] = useState(false);

  const uploadMutation = useMutation({
    mutationFn: async () => {
      setError('');
      if (!executionName.trim()) throw new Error('Execution name is required');

      if (tab === 0) {
        if (!file) throw new Error('Please select a file');
        const formData = new FormData();
        formData.append('file', file);
        formData.append('execution_name', executionName.trim());
        if (reportName.trim()) formData.append('name', reportName.trim());
        if (tags.length)
          formData.append('tags', tags.map((t) => `${t.key}:${t.value}`).join(','));
        return uploadReport(formData);
      } else {
        if (!xmlText.trim()) throw new Error('Please paste XML content');
        return ingestRawXML(
          xmlText,
          executionName.trim(),
          reportName.trim() || undefined,
          tags.length ? tags.map((t) => `${t.key}:${t.value}`).join(',') : undefined
        );
      }
    },
    onSuccess: (data) => navigate(`/reports/${data.report_id}`),
    onError: (err: unknown) => {
      const e = err as { response?: { data?: { details?: string } }; message?: string };
      setError(e?.response?.data?.details || e?.message || 'Upload failed');
    },
  });

  const addTag = () => {
    const parts = tagInput.split(':');
    if (parts.length === 2 && parts[0].trim() && parts[1].trim()) {
      setTags([...tags, { key: parts[0].trim(), value: parts[1].trim() }]);
      setTagInput('');
    }
  };

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
    const droppedFile = e.dataTransfer.files[0];
    if (droppedFile) setFile(droppedFile);
  }, []);

  if (!manualUploadEnabled) {
    return (
      <Box sx={{ maxWidth: 640, mx: 'auto', textAlign: 'center', mt: 8 }}>
        <CloudUpload sx={{ fontSize: 64, color: 'text.disabled', mb: 2 }} />
        <Typography variant="h5" fontWeight={700} sx={{ mb: 1 }}>
          Manual Upload Disabled
        </Typography>
        <Typography color="text.secondary">
          Manual report uploads have been disabled by the administrator.
          Reports can still be submitted via the API.
        </Typography>
      </Box>
    );
  }

  return (
    <Box sx={{ maxWidth: 640, mx: 'auto' }}>
      <Typography variant="h5" fontWeight={700} sx={{ mb: 2 }}>
        Upload Test Report
      </Typography>

      <Card>
        <CardContent sx={{ p: 2 }}>
          <Tabs value={tab} onChange={(_, v) => setTab(v)} sx={{ mb: 2 }}>
            <Tab label="File Upload" icon={<CloudUpload />} iconPosition="start" sx={{ minHeight: 40 }} />
            <Tab label="Paste XML" icon={<ContentPaste />} iconPosition="start" sx={{ minHeight: 40 }} />
          </Tabs>

          <Stack spacing={2}>
            <TextField
              label="Execution Name"
              required
              size="small"
              value={executionName}
              onChange={(e) => setExecutionName(e.target.value)}
              placeholder="e.g., api-smoke, ui-regression"
            />
            <TextField
              label="Report Name (optional)"
              size="small"
              value={reportName}
              onChange={(e) => setReportName(e.target.value)}
            />

            {tab === 0 ? (
              <Box
                onDragOver={(e) => { e.preventDefault(); setDragOver(true); }}
                onDragLeave={() => setDragOver(false)}
                onDrop={handleDrop}
                onClick={() => document.getElementById('file-input')?.click()}
                sx={{
                  border: '2px dashed',
                  borderColor: dragOver ? 'primary.main' : 'divider',
                  borderRadius: 2,
                  p: 3,
                  textAlign: 'center',
                  cursor: 'pointer',
                  bgcolor: dragOver ? 'action.hover' : 'transparent',
                  transition: 'all 0.2s',
                }}
              >
                <input
                  id="file-input"
                  type="file"
                  accept=".xml"
                  hidden
                  onChange={(e) => setFile(e.target.files?.[0] || null)}
                />
                <CloudUpload sx={{ fontSize: 32, color: 'text.secondary', mb: 0.5 }} />
                <Typography variant="body2" color="text.secondary">
                  {file ? file.name : 'Drop JUnit XML file here or click to browse'}
                </Typography>
              </Box>
            ) : (
              <TextField
                multiline
                rows={8}
                placeholder="Paste your JUnit XML here..."
                value={xmlText}
                onChange={(e) => setXmlText(e.target.value)}
                size="small"
                sx={{ '& .MuiInputBase-root': { fontFamily: 'monospace', fontSize: '0.75rem' } }}
              />
            )}

            <Box>
              <Typography variant="caption" color="text.secondary" sx={{ mb: 0.5, display: 'block' }}>
                Tags (key:value)
              </Typography>
              <Box sx={{ display: 'flex', gap: 0.5, alignItems: 'center' }}>
                <TextField
                  size="small"
                  placeholder="build:1042"
                  value={tagInput}
                  onChange={(e) => setTagInput(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && addTag()}
                  sx={{ flex: 1 }}
                  slotProps={{ htmlInput: { style: { fontSize: '0.8125rem', padding: '6px 10px' } }}}
                />
                <Button size="small" onClick={addTag}>
                  Add
                </Button>
              </Box>
              <Stack direction="row" spacing={0.5} sx={{ mt: 0.5 }} flexWrap="wrap" useFlexGap>
                {tags.map((t, i) => (
                  <Chip
                    key={i}
                    label={`${t.key}: ${t.value}`}
                    size="small"
                    onDelete={() => setTags(tags.filter((_, j) => j !== i))}
                  />
                ))}
              </Stack>
            </Box>

            {error && <Alert severity="error" sx={{ py: 0.5 }}>{error}</Alert>}
            {uploadMutation.isPending && <LinearProgress />}

            <Button
              variant="contained"
              startIcon={<Send />}
              onClick={() => uploadMutation.mutate()}
              disabled={uploadMutation.isPending}
            >
              Submit Report
            </Button>
          </Stack>
        </CardContent>
      </Card>
    </Box>
  );
};

export default Upload;
