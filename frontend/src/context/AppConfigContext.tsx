import React, { createContext, useContext } from 'react';
import { useQuery } from '@tanstack/react-query';
import axios from 'axios';

interface AppConfig {
  manualUploadEnabled: boolean;
}

const defaultConfig: AppConfig = { manualUploadEnabled: true };

const AppConfigContext = createContext<AppConfig>(defaultConfig);

export const useAppConfig = () => useContext(AppConfigContext);

export const AppConfigProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { data } = useQuery<AppConfig>({
    queryKey: ['app-config'],
    queryFn: async () => {
      const baseURL = (import.meta as any).env?.VITE_API_BASE_URL
        ? `${(import.meta as any).env.VITE_API_BASE_URL}/api/v1`
        : '/api/v1';
      const { data } = await axios.get(`${baseURL}/config`);
      return {
        manualUploadEnabled: data.manual_upload_enabled ?? true,
      };
    },
    staleTime: Infinity,
    retry: 1,
  });

  return (
    <AppConfigContext.Provider value={data ?? defaultConfig}>
      {children}
    </AppConfigContext.Provider>
  );
};
