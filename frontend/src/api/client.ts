import axios from 'axios';
import type {
  ListReportsResponse,
  TestReport,
  ExecutionHistoryItem,
  TestHistoryItem,
  Stats,
} from '../types';

export interface TagInfo {
  key: string;
  value: string;
  count: number;
}

const meta = import.meta as ImportMeta & { env?: Record<string, string> };

const api = axios.create({
  baseURL: meta.env?.VITE_API_BASE_URL
    ? `${meta.env.VITE_API_BASE_URL}/api/v1`
    : '/api/v1',
});

// Reports
export const listReports = async (params: Record<string, string>) => {
  const { data } = await api.get<ListReportsResponse>('/reports', { params });
  return data;
};

export const getReport = async (id: string) => {
  const { data } = await api.get<TestReport>(`/reports/${id}`);
  return data;
};

// Ingest
export const uploadReport = async (formData: FormData) => {
  const { data } = await api.post('/reports/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
  return data;
};

export const ingestRawXML = async (
  xml: string,
  executionName: string,
  name?: string,
  tags?: string
) => {
  const { data } = await api.post('/reports/ingest', xml, {
    headers: {
      'Content-Type': 'application/xml',
      'X-Execution-Name': executionName,
      ...(name && { 'X-Report-Name': name }),
      ...(tags && { 'X-Tags': tags }),
    },
  });
  return data;
};

// Tags
export const addTags = async (reportId: string, tags: { key: string; value: string }[]) => {
  const { data } = await api.post(`/reports/${reportId}/tags`, { tags });
  return data;
};

export const removeTag = async (reportId: string, key: string, value: string) => {
  const { data } = await api.delete(`/reports/${reportId}/tags`, { data: { key, value } });
  return data;
};

export const listTags = async (key?: string) => {
  const { data } = await api.get<TagInfo[]>('/tags', { params: key ? { key } : {} });
  return data;
};

export const listKnownTagKeys = async () => {
  const { data } = await api.get<string[]>('/tags/keys');
  return data;
};

// History
export const getExecutionHistory = async (executionName: string, limit = 50) => {
  const { data } = await api.get<ExecutionHistoryItem[]>(
    `/executions/${encodeURIComponent(executionName)}/reports`,
    { params: { limit } }
  );
  return data;
};

export const getTestHistory = async (
  executionName: string,
  testName: string,
  limit = 100
) => {
  const { data } = await api.get<TestHistoryItem[]>(
    `/executions/${encodeURIComponent(executionName)}/tests/${encodeURIComponent(testName)}/history`,
    { params: { limit } }
  );
  return data;
};

// Stats
export const getStats = async () => {
  const { data } = await api.get<Stats>('/stats');
  return data;
};

export const getExecutionNames = async () => {
  const { data } = await api.get<string[]>('/execution-names');
  return data;
};
