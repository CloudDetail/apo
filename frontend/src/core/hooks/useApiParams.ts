import { useState } from 'react';

/**
 * 用于构建 API 参数并发送请求的钩子
 * @param apiFunction 要调用的 API 函数
 * @returns 包含构建参数和发送请求方法的对象
 */
export function useApiParams<T = any, P = any>(
  apiFunction: (params: URLSearchParams | P) => Promise<T>
) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [data, setData] = useState<T | null>(null);

  /**
   * 构建 URLSearchParams 并发送请求
   * @param paramsObj 参数对象
   * @param options 请求选项
   * @returns 请求结果
   */
  const sendRequest = async (
    paramsObj: Record<string, any>,
    options?: {
      onSuccess?: (data: T) => void;
      onError?: (error: Error) => void;
      useURLSearchParams?: boolean; // 是否使用 URLSearchParams
    }
  ): Promise<T | null> => {
    const { onSuccess, onError, useURLSearchParams = true } = options || {};

    setLoading(true);
    setError(null);

    try {
      let params: URLSearchParams | P;

      if (useURLSearchParams) {
        // 构建 URLSearchParams
        params = new URLSearchParams();

        // 处理普通键值对
        Object.entries(paramsObj).forEach(([key, value]) => {
          if (value !== undefined && value !== null) {
            if (Array.isArray(value)) {
              // 处理数组值
              value.forEach(item => {
                if (item !== undefined && item !== null) {
                  (params as URLSearchParams).append(key, item.toString());
                }
              });
            } else {
              // 处理普通值
              (params as URLSearchParams).append(key, value.toString());
            }
          }
        });
      } else {
        // 直接使用对象参数
        params = paramsObj as P;
      }

      // 发送请求
      const result = await apiFunction(params);

      setData(result);
      onSuccess?.(result);

      return result;
    } catch (err) {
      const error = err instanceof Error ? err : new Error(String(err));
      setError(error);
      onError?.(error);
      return null;
    } finally {
      setLoading(false);
    }
  };

  return {
    loading,
    error,
    data,
    sendRequest
  };
}