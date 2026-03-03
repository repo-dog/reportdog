import React, { useState } from 'react';
import {
  Chip,
  Stack,
  IconButton,
  TextField,
  Box,
  Tooltip,
} from '@mui/material';
import { Add } from '@mui/icons-material';

interface Tag {
  key: string;
  value: string;
}

interface Props {
  tags: Tag[];
  onAdd?: (key: string, value: string) => void;
  onRemove?: (key: string, value: string) => void;
  onClick?: (key: string, value: string) => void;
  editable?: boolean;
}

const TagChips: React.FC<Props> = ({ tags, onAdd, onRemove, onClick, editable = false }) => {
  const [adding, setAdding] = useState(false);
  const [key, setKey] = useState('');
  const [value, setValue] = useState('');

  const handleAdd = () => {
    if (key.trim() && value.trim() && onAdd) {
      onAdd(key.trim(), value.trim());
      setKey('');
      setValue('');
      setAdding(false);
    }
  };

  return (
    <Stack direction="row" spacing={0.5} flexWrap="wrap" alignItems="center" useFlexGap>
      {tags.map((tag) => (
        <Chip
          key={`${tag.key}:${tag.value}`}
          label={`${tag.key}: ${tag.value}`}
          size="small"
          variant="outlined"
          clickable={!!onClick}
          onClick={onClick ? () => onClick(tag.key, tag.value) : undefined}
          onDelete={editable && onRemove ? () => onRemove(tag.key, tag.value) : undefined}
          sx={onClick ? { cursor: 'pointer', '&:hover': { borderColor: 'primary.main' } } : undefined}
        />
      ))}
      {editable && !adding && (
        <Tooltip title="Add tag">
          <IconButton
            size="small"
            onClick={() => setAdding(true)}
            sx={{
              border: '1px dashed',
              borderColor: 'divider',
              borderRadius: 1,
              p: 0.5,
              '&:hover': { borderColor: 'primary.main', bgcolor: 'action.hover' },
            }}
          >
            <Add fontSize="small" />
          </IconButton>
        </Tooltip>
      )}
      {adding && (
        <Box sx={{ display: 'flex', gap: 0.5, alignItems: 'center' }}>
          <TextField
            size="small"
            placeholder="key"
            value={key}
            onChange={(e) => setKey(e.target.value)}
            sx={{ width: 80 }}
            slotProps={{ htmlInput: { style: { fontSize: '0.75rem', padding: '4px 8px' } }}}
          />
          <TextField
            size="small"
            placeholder="value"
            value={value}
            onChange={(e) => setValue(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleAdd()}
            sx={{ width: 100 }}
            slotProps={{ htmlInput: { style: { fontSize: '0.75rem', padding: '4px 8px' } }}}
          />
          <Chip label="Add" size="small" color="primary" onClick={handleAdd} clickable />
          <Chip label="Cancel" size="small" onClick={() => setAdding(false)} clickable />
        </Box>
      )}
    </Stack>
  );
};

export default TagChips;
