'use client';

import { create } from 'zustand';
import { createJSONStorage, persist } from 'zustand/middleware';

export type RankSortMode = 'count' | 'tokens';
export type ChartMetricType = 'count' | 'tokens';
export type ChartPeriod = '1' | '7' | '30';

interface HomeViewState {
    rankSortMode: RankSortMode;
    chartMetricType: ChartMetricType;
    chartPeriod: ChartPeriod;
    setRankSortMode: (value: RankSortMode) => void;
    setChartMetricType: (value: ChartMetricType) => void;
    setChartPeriod: (value: ChartPeriod) => void;
}

export const useHomeViewStore = create<HomeViewState>()(
    persist(
        (set) => ({
            rankSortMode: 'count',
            chartMetricType: 'count',
            chartPeriod: '1',
            setRankSortMode: (value) => set({ rankSortMode: value }),
            setChartMetricType: (value) => set({ chartMetricType: value }),
            setChartPeriod: (value) => set({ chartPeriod: value }),
        }),
        {
            name: 'home-view-options-storage',
            storage: createJSONStorage(() => localStorage),
            partialize: (state) => ({
                rankSortMode: state.rankSortMode,
                chartMetricType: state.chartMetricType,
                chartPeriod: state.chartPeriod,
            }),
            merge: (persisted, current) => {
                const p = (persisted ?? {}) as Partial<HomeViewState>;
                const rankSortMode = p.rankSortMode === 'tokens' ? 'tokens' : 'count';
                const chartMetricType = p.chartMetricType === 'tokens' ? 'tokens' : 'count';
                return {
                    ...current,
                    ...p,
                    rankSortMode,
                    chartMetricType,
                };
            },
        }
    )
);
