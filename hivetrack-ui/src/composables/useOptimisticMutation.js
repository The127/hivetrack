import { useMutation } from "@tanstack/vue-query";

/**
 * Wraps useMutation with the standard optimistic update boilerplate:
 * cancel in-flight queries → snapshot → apply optimistic update → rollback on error → invalidate on settle.
 *
 * @param {import('@tanstack/vue-query').QueryClient} queryClient
 * @param {object} opts
 * @param {() => unknown[]} opts.queryKey — function returning the query key (evaluated lazily so reactive values work)
 * @param {(vars: any) => Promise} opts.mutationFn
 * @param {(oldData: any, vars: any) => any} opts.updater — returns new cache value from old + mutation vars
 * @param {Array<() => unknown[]>} [opts.invalidateKeys] — extra keys to invalidate on settle (queryKey is always invalidated)
 * @param {() => void} [opts.onSettled] — extra callback after invalidation
 * @returns {import('@tanstack/vue-query').UseMutationReturnType}
 */
export function useOptimisticMutation(queryClient, opts) {
  return useMutation({
    mutationFn: opts.mutationFn,
    onMutate: async (vars) => {
      const key = opts.queryKey();
      await queryClient.cancelQueries({ queryKey: key });
      const previous = queryClient.getQueryData(key);
      queryClient.setQueryData(key, (old) => {
        if (!old) return old;
        return opts.updater(old, vars);
      });
      return { previous, key };
    },
    onError: (_err, _vars, context) => {
      if (context?.previous) {
        queryClient.setQueryData(context.key, context.previous);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: opts.queryKey() });
      if (opts.invalidateKeys) {
        for (const keyFn of opts.invalidateKeys) {
          queryClient.invalidateQueries({ queryKey: keyFn() });
        }
      }
      opts.onSettled?.();
    },
  });
}

/**
 * Common updater: merge fields into a matching item in a paginated { items: [] } response.
 */
export function updateItemByNumber(number, fields) {
  return (old) => ({
    ...old,
    items: old.items.map((i) =>
      i.number === number ? { ...i, ...fields } : i,
    ),
  });
}
