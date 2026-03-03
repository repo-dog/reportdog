import React, { useState } from 'react';
import {
  Box,
  Drawer,
  List,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Typography,
  IconButton,
  Divider,
  Tooltip,
  useTheme,
  useMediaQuery,
  alpha,
} from '@mui/material';
import {
  SpaceDashboardRounded,
  SummarizeRounded,
  CloudUploadRounded,
  DarkModeRounded,
  LightModeRounded,
  ChevronLeftRounded,
  MenuRounded,
} from '@mui/icons-material';
import { useNavigate, useLocation, Outlet } from 'react-router-dom';
import { useThemeMode } from '../theme/ThemeContext';
import { useAppConfig } from '../context/AppConfigContext';

const DRAWER_WIDTH = 220;
const DRAWER_COLLAPSED = 60;

const ALL_NAV_ITEMS = [
  { label: 'Home', path: '/', icon: <SpaceDashboardRounded />, key: 'home' },
  { label: 'Test Reports', path: '/reports', icon: <SummarizeRounded />, key: 'reports' },
  { label: 'Upload', path: '/upload', icon: <CloudUploadRounded />, key: 'upload' },
];

const Layout: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { mode, toggle } = useThemeMode();
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const [collapsed, setCollapsed] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);
  const { manualUploadEnabled } = useAppConfig();

  const NAV_ITEMS = manualUploadEnabled
    ? ALL_NAV_ITEMS
    : ALL_NAV_ITEMS.filter((item) => item.key !== 'upload');

  const drawerWidth = collapsed ? DRAWER_COLLAPSED : DRAWER_WIDTH;

  const drawerContent = (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        height: '100%',
        overflow: 'hidden',
        bgcolor: mode === 'light'
          ? alpha(theme.palette.primary.main, 0.04)
          : alpha(theme.palette.primary.main, 0.06),
      }}
    >
      {/* Brand header */}
      <Box
        sx={{
          display: 'flex',
          alignItems: 'center',
          px: collapsed ? 1 : 2,
          py: 1.5,
          gap: 1,
          minHeight: 48,
          justifyContent: collapsed ? 'center' : 'flex-start',
        }}
      >
        <Box
          component="img"
          src="/favicon.svg"
          alt="ReportDog"
          sx={{ width: 28, height: 28, flexShrink: 0 }}
        />
        {!collapsed && (
          <Typography variant="subtitle1" fontWeight={700} noWrap>
            ReportDog
          </Typography>
        )}
      </Box>
      <Divider />

      {/* Nav items */}
      <List sx={{ flex: 1, pt: 1, px: 0.5 }}>
        {NAV_ITEMS.map((item) => {
          const isActive =
            item.path === '/'
              ? location.pathname === '/'
              : location.pathname.startsWith(item.path);
          return (
            <Tooltip
              key={item.path}
              title={collapsed ? item.label : ''}
              placement="right"
              arrow
            >
              <ListItemButton
                selected={isActive}
                onClick={() => {
                  navigate(item.path);
                  if (isMobile) setMobileOpen(false);
                }}
                sx={{
                  borderRadius: 1.5,
                  mb: 0.5,
                  minHeight: 40,
                  px: collapsed ? 1.5 : 2,
                  justifyContent: collapsed ? 'center' : 'flex-start',
                  ...(isActive && {
                    bgcolor: alpha(theme.palette.primary.main, 0.12),
                    '&:hover': {
                      bgcolor: alpha(theme.palette.primary.main, 0.18),
                    },
                  }),
                }}
              >
                <ListItemIcon
                  sx={{
                    minWidth: collapsed ? 0 : 36,
                    color: isActive ? 'primary.main' : 'text.secondary',
                    justifyContent: 'center',
                  }}
                >
                  {item.icon}
                </ListItemIcon>
                {!collapsed && (
                  <ListItemText
                    primary={item.label}
                    primaryTypographyProps={{
                      variant: 'body2',
                      fontWeight: isActive ? 600 : 400,
                    }}
                  />
                )}
              </ListItemButton>
            </Tooltip>
          );
        })}
      </List>

      <Divider />
      {/* Bottom controls */}
      <Box
        sx={{
          display: 'flex',
          flexDirection: collapsed ? 'column' : 'row',
          alignItems: 'center',
          justifyContent: collapsed ? 'center' : 'space-between',
          px: collapsed ? 0 : 1,
          py: 1,
          gap: 0.5,
        }}
      >
        <Tooltip title={mode === 'dark' ? 'Light mode' : 'Dark mode'} placement="right" arrow>
          <IconButton onClick={toggle} size="small">
            {mode === 'dark' ? <LightModeRounded fontSize="small" /> : <DarkModeRounded fontSize="small" />}
          </IconButton>
        </Tooltip>
        {!isMobile && (
          <Tooltip title={collapsed ? 'Expand' : 'Collapse'} placement="right" arrow>
            <IconButton onClick={() => setCollapsed(!collapsed)} size="small">
              {collapsed ? <MenuRounded fontSize="small" /> : <ChevronLeftRounded fontSize="small" />}
            </IconButton>
          </Tooltip>
        )}
      </Box>
    </Box>
  );

  return (
    <Box sx={{ display: 'flex', minHeight: '100vh' }}>
      {/* Mobile hamburger */}
      {isMobile && (
        <IconButton
          onClick={() => setMobileOpen(true)}
          sx={{
            position: 'fixed',
            top: 8,
            left: 8,
            zIndex: theme.zIndex.drawer + 1,
            bgcolor: 'background.paper',
            boxShadow: 1,
            '&:hover': { bgcolor: 'action.hover' },
          }}
          size="small"
        >
          <MenuRounded fontSize="small" />
        </IconButton>
      )}

      {/* Sidebar drawer */}
      {isMobile ? (
        <Drawer
          variant="temporary"
          open={mobileOpen}
          onClose={() => setMobileOpen(false)}
          ModalProps={{ keepMounted: true }}
          sx={{
            '& .MuiDrawer-paper': {
              width: DRAWER_WIDTH,
              boxSizing: 'border-box',
              borderRight: 'none',
            },
          }}
        >
          {drawerContent}
        </Drawer>
      ) : (
        <Drawer
          variant="permanent"
          sx={{
            width: drawerWidth,
            flexShrink: 0,
            transition: theme.transitions.create('width', {
              duration: theme.transitions.duration.shorter,
            }),
            '& .MuiDrawer-paper': {
              width: drawerWidth,
              boxSizing: 'border-box',
              borderRight: 'none',
              overflowX: 'hidden',
              transition: theme.transitions.create('width', {
                duration: theme.transitions.duration.shorter,
              }),
              boxShadow: mode === 'light'
                ? '1px 0 8px rgba(0,0,0,0.06)'
                : '1px 0 8px rgba(0,0,0,0.3)',
            },
          }}
        >
          {drawerContent}
        </Drawer>
      )}

      {/* Main content */}
      <Box
        component="main"
        sx={{
          flex: 1,
          p: { xs: 2, md: 3 },
          maxWidth: 1200,
          width: '100%',
          ml: isMobile ? 0 : undefined,
        }}
      >
        <Outlet />
      </Box>
    </Box>
  );
};

export default Layout;
