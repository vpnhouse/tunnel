// This file is a compatibility layer for makeStyles migration
// Components should be gradually migrated to styled/sx
import { SxProps, Theme } from '@mui/material/styles';

// Dummy useStyles hook for backward compatibility during migration
export default function useStyles(_props?: Record<string, unknown>) {
    return new Proxy({} as Record<string, string>, {
        get: (_, prop) => String(prop)
    });
}

export type { SxProps, Theme };
