import React from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';
import { AdapterDayjs } from '@mui/x-date-pickers/AdapterDayjs';
import { ThemeProvider } from './theme/ThemeContext';
import { AppConfigProvider } from './context/AppConfigContext';
import Layout from './components/Layout';
import Home from './pages/Home';
import Reports from './pages/Reports';
import Upload from './pages/Upload';
import ReportDetail from './pages/ReportDetail';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
      staleTime: 30_000,
    },
  },
});

const App: React.FC = () => (
  <QueryClientProvider client={queryClient}>
    <LocalizationProvider dateAdapter={AdapterDayjs}>
      <ThemeProvider>
        <AppConfigProvider>
        <BrowserRouter>
          <Routes>
            <Route element={<Layout />}>
              <Route path="/" element={<Home />} />
              <Route path="/reports" element={<Reports />} />
              <Route path="/upload" element={<Upload />} />
              <Route path="/reports/:id" element={<ReportDetail />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </AppConfigProvider>
      </ThemeProvider>
    </LocalizationProvider>
  </QueryClientProvider>
);

export default App;
