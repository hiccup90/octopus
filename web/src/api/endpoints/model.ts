import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../client';
import { logger } from '@/lib/logger';

/**
 * LLM 价格信息
 */
export interface LLMPrice {
    input: number;
    output: number;
    cache_read: number;
    cache_write: number;
}

/**
 * LLM 模型信息
 */
export interface LLMInfo extends LLMPrice {
    name: string;
}

/**
 * LLM 渠道关联信息
 */
export interface LLMChannel {
    name: string;
    enabled: boolean;
    channel_id: number;
    channel_name: string;
}

/**
 * 获取 LLM 模型列表 Hook
 * 
 * @example
 * const { data: models, isLoading, error } = useModelList();
 * 
 * if (isLoading) return <Loading />;
 * if (error) return <Error message={error.message} />;
 * 
 * models?.forEach(model => console.log(model.name, model.input));
 */
export function useModelList() {
    return useQuery({
        queryKey: ['models', 'list'],
        queryFn: async () => {
            return apiClient.get<LLMInfo[]>('/api/v1/model/list');
        },
        refetchInterval: 30000,
        refetchOnMount: 'always',
    });
}

/**
 * 获取 LLM 模型与渠道关联列表 Hook
 * 
 * @example
 * const { data: channelModels, isLoading, error } = useModelChannelList();
 * 
 * if (isLoading) return <Loading />;
 * if (error) return <Error message={error.message} />;
 * 
 * channelModels?.forEach(item => console.log(item.name, item.channel_name));
 */
export function useModelChannelList() {
    return useQuery({
        queryKey: ['models', 'channel'],
        queryFn: async () => {
            return apiClient.get<LLMChannel[]>('/api/v1/model/channel');
        },
        refetchInterval: 30000,
    });
}

/**
 * 更新 LLM 模型 Hook
 * 
 * @example
 * const updateModel = useUpdateModel();
 * 
 * updateModel.mutate({
 *   name: 'gpt-4',
 *   input: 0.03,
 *   output: 0.06,
 *   cache_read: 0.015,
 *   cache_write: 0.03,
 * });
 */
export function useUpdateModel() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (data: LLMInfo) => {
            return apiClient.post<LLMInfo>('/api/v1/model/update', data);
        },
        onSuccess: (data) => {
            logger.log('模型更新成功:', data);
            queryClient.invalidateQueries({ queryKey: ['models', 'list'] });
        },
        onError: (error) => {
            logger.error('模型更新失败:', error);
        },
    });
}

/**
 * 创建 LLM 模型 Hook
 * 
 * @example
 * const createModel = useCreateModel();
 * 
 * createModel.mutate({
 *   name: 'gpt-4',
 *   input: 0.03,
 *   output: 0.06,
 *   cache_read: 0.015,
 *   cache_write: 0.03,
 * });
 */
export function useCreateModel() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (data: LLMInfo) => {
            return apiClient.post<LLMInfo>('/api/v1/model/create', data);
        },
        onSuccess: (data) => {
            logger.log('模型创建成功:', data);
            queryClient.invalidateQueries({ queryKey: ['models', 'list'] });
        },
        onError: (error) => {
            logger.error('模型创建失败:', error);
        },
    });
}

/**
 * 删除 LLM 模型 Hook
 * 
 * @example
 * const deleteModel = useDeleteModel();
 * 
 * deleteModel.mutate('gpt-4'); // 删除名称为 'gpt-4' 的模型
 */
export function useDeleteModel() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (name: string) => {
            return apiClient.post<null>('/api/v1/model/delete', { name });
        },
        onSuccess: () => {
            logger.log('模型删除成功');
            queryClient.invalidateQueries({ queryKey: ['models', 'list'] });
        },
        onError: (error) => {
            logger.error('模型删除失败:', error);
        },
    });
}

