import { useState } from 'react';
import { NavLink, Outlet, useLocation } from 'react-router-dom';
import {
  AppBar, Box, Button, Chip, Divider, Drawer, IconButton, List, ListItemButton, ListItemIcon,
  ListItemText, Toolbar, Typography, useMediaQuery, useTheme,
} from '@mui/material';
import MenuIcon from '@mui/icons-material/Menu';
import PeopleIcon from '@mui/icons-material/People';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import AssessmentIcon from '@mui/icons-material/Assessment';
import ShowChartIcon from '@mui/icons-material/ShowChart';
import CompareArrowsIcon from '@mui/icons-material/CompareArrows';
import GroupsIcon from '@mui/icons-material/Groups';
import TuneIcon from '@mui/icons-material/Tune';
import HelpOutlineIcon from '@mui/icons-material/HelpOutline';
import LogoutIcon from '@mui/icons-material/Logout';
import { useAuth } from '../auth/AuthContext';

const DRAWER = 248;

interface NavItem { to: string; label: string; icon: JSX.Element; admin?: boolean }

const NAV: NavItem[] = [
  { to: '/test', label: 'Пройти тест', icon: <PlayArrowIcon /> },
  { to: '/results', label: 'Результаты', icon: <AssessmentIcon /> },
  { to: '/dynamics', label: 'Динамика', icon: <ShowChartIcon /> },
  { to: '/subjects', label: 'Испытуемые', icon: <PeopleIcon />, admin: true },
  { to: '/compare', label: 'Сравнение', icon: <CompareArrowsIcon />, admin: true },
  { to: '/groups', label: 'Группа', icon: <GroupsIcon />, admin: true },
  { to: '/settings', label: 'Настройки', icon: <TuneIcon />, admin: true },
  { to: '/help', label: 'Справка', icon: <HelpOutlineIcon /> },
];

export function Layout() {
  const { role, logout } = useAuth();
  const theme = useTheme();
  const desktop = useMediaQuery(theme.breakpoints.up('md'));
  const [open, setOpen] = useState(false);
  const location = useLocation();
  const items = NAV.filter((n) => !n.admin || role === 'admin');

  const drawer = (
    <Box sx={{ width: DRAWER }} role="navigation">
      <Toolbar />
      <List>
        {items.map((n) => (
          <ListItemButton
            key={n.to} component={NavLink} to={n.to} onClick={() => setOpen(false)}
            selected={location.pathname.startsWith(n.to)}
          >
            <ListItemIcon>{n.icon}</ListItemIcon>
            <ListItemText primary={n.label} />
          </ListItemButton>
        ))}
      </List>
      <Divider />
    </Box>
  );

  return (
    <Box sx={{ display: 'flex' }}>
      <AppBar position="fixed" sx={{ zIndex: (t) => t.zIndex.drawer + 1 }}>
        <Toolbar>
          {!desktop && (
            <IconButton color="inherit" edge="start" onClick={() => setOpen(true)} sx={{ mr: 1 }} aria-label="меню">
              <MenuIcon />
            </IconButton>
          )}
          <Typography variant="h6" sx={{ flexGrow: 1 }}>Таблицы Шульте</Typography>
          <Chip
            size="small" color="secondary" sx={{ mr: 2, color: 'white' }}
            label={role === 'admin' ? 'Режим психолога' : 'Режим испытуемого'}
          />
          <Button color="inherit" startIcon={<LogoutIcon />} onClick={logout}>Выйти</Button>
        </Toolbar>
      </AppBar>
      <Drawer
        variant={desktop ? 'permanent' : 'temporary'} open={desktop || open} onClose={() => setOpen(false)}
        sx={{ width: desktop ? DRAWER : 0, '& .MuiDrawer-paper': { width: DRAWER, boxSizing: 'border-box' } }}
      >
        {drawer}
      </Drawer>
      <Box component="main" sx={{ flexGrow: 1, p: { xs: 2, md: 3 }, minWidth: 0 }}>
        <Toolbar />
        <Outlet />
      </Box>
    </Box>
  );
}
