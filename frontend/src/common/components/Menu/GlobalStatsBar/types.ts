export interface GlobalStats {
  peers_total: number;
  peers_connected: number;
  peers_active_1h: number;
  peers_active_1d: number;
  traffic_up: string;
  traffic_down: string;
}

export interface GlobalStatsResponse {
  peers_total: number;
  peers_connected: number;
  peers_active_1h: number;
  peers_active_1d: number;
  traffic_up: number;
  traffic_down: number;
}
