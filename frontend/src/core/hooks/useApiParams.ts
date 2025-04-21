/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState } from 'react';

/**
 * Hook for building API parameters and sending requests
 * @param apiFunction API function to call
 * @returns Object containing parameter builder and request methods
 */
export function useApiParams<T = any, P = any>(
  apiFunction: (params: URLSearchParams | P) => Promise<T>
) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [data, setData] = useState<T | null>(null);

  /**
   * Build URLSearchParams and send request
   * @param paramsObj Parameter object
   * @param options Request options
   * @returns Request result
   */
  const sendRequest = async (
    paramsObj: Record<string, any>,
    options?: {
      onSuccess?: (data: T) => void;
      onError?: (error: Error) => void;
      useURLSearchParams?: boolean; // Whether to use URLSearchParams
    }
  ): Promise<T | null> => {
    const { onSuccess, onError, useURLSearchParams = true } = options || {};

    setLoading(true);
    setError(null);

    try {
      let params: URLSearchParams | P;

      if (useURLSearchParams) {
        // Build URLSearchParams
        params = new URLSearchParams();

        // Process key-value pairs
        Object.entries(paramsObj).forEach(([key, value]) => {
          if (value !== undefined && value !== null) {
            if (Array.isArray(value)) {
              // Handle array values
              value.forEach(item => {
                if (item !== undefined && item !== null) {
                  (params as URLSearchParams).append(key, item.toString());
                }
              });
            } else {
              // Handle primitive values
              (params as URLSearchParams).append(key, value.toString());
            }
          }
        });
      } else {
        // Use object parameters directly
        params = paramsObj as P;
      }

      // Send request
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