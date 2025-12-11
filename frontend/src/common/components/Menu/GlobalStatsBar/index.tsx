import { useCallback, useEffect, useState, SyntheticEvent } from 'react';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import Box from '@mui/material/Box';
import { styled } from '@mui/material/styles';

import { components } from '@schema';
import { fetchData } from '@root/store/utils';
import { GLOBAL_STATS } from '@constants/apiPaths';
import { BYTES_MEASURE_LIMITS, FETCH_STATS_INTERVAL, TABS } from '@common/components/Menu/GlobalStatsBar/constant';

type TabType = 'stats_global' | 'stats_iprose' | 'stats_proxy' | 'stats_wireguard';

const StatsRoot = styled(Box)(({ theme }) => ({
  marginTop: 'auto',
  padding: '16px',
  [theme.breakpoints.down('lg')]: {
    display: 'none'
  }
}));

const TabsContainer = styled(Box)({
  marginBottom: 12
});

const StyledTabs = styled(Tabs)({
  minHeight: 32,
  '& .MuiTab-root': {
    minHeight: 32,
    padding: '4px 8px',
    fontSize: 11,
    minWidth: 'auto'
  }
});

const StatsContent = styled(Box)(({ theme }) => ({
  fontSize: 12,
  color: theme.palette.text.secondary
}));

const StatsRow = styled(Box)({
  display: 'flex',
  justifyContent: 'space-between',
  marginBottom: 4,
  '& span:last-of-type': {
    fontWeight: 500
  }
});

const GlobalStatsBar = () => {
  const [stats, setStats] = useState<components['schemas']['ServiceStatus'] | null>(null);
  const [activeTab, setActiveTab] = useState<TabType>('stats_global');

  const convertBytes = useCallback((bytes: number): string => {
    for (const [index, metric] of BYTES_MEASURE_LIMITS.entries()) {
      const { limit, label } = metric;

      if (bytes < limit) {
        return `${(bytes / (1000 ** index)).toFixed(2)} ${label}`;
      }
    }

    return `${bytes} B`;
  }, []);

  const fetchStats = useCallback(async () => {
    const response = await fetchData(GLOBAL_STATS);
    const data: components['schemas']['ServiceStatus'] = await response.json();

    setStats(data);
  }, []);

  useEffect(() => {
    fetchStats();
    const fetchStatsInterval = setInterval(fetchStats, FETCH_STATS_INTERVAL);

    return () => {
      clearInterval(fetchStatsInterval);
    };
  }, [fetchStats]);

  const handleTabChange = (_event: SyntheticEvent, newValue: TabType) => {
    setActiveTab(newValue);
  };

  if (!stats) {
    return null;
  }

  const emptyStats = {
    peers_total: 0,
    peers_active: 0,
    traffic_up: '0 B',
    traffic_down: '0 B',
    speed_down: '0 Bps',
    speed_up: '0 Bps'
  };

  const activeStats = stats?.[activeTab];

  const displayStats = activeStats
    ? {
      ...activeStats,
      traffic_up: convertBytes(activeStats.traffic_up ?? 0),
      traffic_down: convertBytes(activeStats.traffic_down ?? 0),
      speed_up: `${convertBytes(activeStats.speed_up ?? 0)}ps`,
      speed_down: `${convertBytes(activeStats.speed_down ?? 0)}ps`
    }
    : emptyStats;
  const { peers_active, peers_total, traffic_up, traffic_down, speed_up, speed_down } = displayStats;

  return (
    <StatsRoot>
      <TabsContainer>
        <StyledTabs
          value={activeTab}
          onChange={handleTabChange}
          indicatorColor="primary"
          textColor="primary"
          variant="fullWidth"
        >
          {TABS.map((tab) => (
            <Tab key={tab.value} value={tab.value} label={tab.label} />
          ))}
        </StyledTabs>
      </TabsContainer>

      <StatsContent>
        <StatsRow>
          <span>Total peers:</span>
          <span>{peers_total}</span>
        </StatsRow>

        <StatsRow>
          <span>Peers active:</span>
          <span>{peers_active}</span>
        </StatsRow>

        <StatsRow>
          <span>Upstream traffic:</span>
          <span>{traffic_up}</span>
        </StatsRow>

        <StatsRow>
          <span>Downstream traffic:</span>
          <span>{traffic_down}</span>
        </StatsRow>

        <StatsRow>
          <span>Upstream speed:</span>
          <span>{speed_up}</span>
        </StatsRow>

        <StatsRow>
          <span>Downstream speed:</span>
          <span>{speed_down}</span>
        </StatsRow>
      </StatsContent>
    </StatsRoot>
  );
};

export default GlobalStatsBar;
