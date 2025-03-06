export interface GlobalStats {
  peers_total: number;
  peers_connected: number;
  peers_active_1h: number;
  peers_active_1d: number;
  traffic_up: string;
  traffic_down: string;
  traffic_up_speed: string;
  traffic_down_speed: string;
}

export interface GlobalStatsResponse {
  peers_total: number;
  peers_connected: number;
  peers_active_1h: number;
  peers_active_1d: number;
  traffic_up: number;
  traffic_down: number;
  traffic_up_speed: number;
  traffic_down_speed: number;
}

export type TabType = 'all' | 'wireguard' | 'iprose';
