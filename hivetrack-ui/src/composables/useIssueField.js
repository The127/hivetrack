import { useMutation, useQueryClient } from "@tanstack/vue-query";
import { updateIssue } from "@/api/issues";

/**
 * Factory for simple issue field mutations that follow the pattern:
 * call updateIssue → invalidate issue + issues queries.
 *
 * @param {import('vue').Ref<string>} slug — project slug ref
 * @param {import('vue').Ref<number>} number — issue number ref
 * @param {object} opts
 * @param {(value: any) => object} opts.toPayload — transforms the mutated value into the API payload
 * @param {string[][]} [opts.extraInvalidate] — additional query key prefixes to invalidate
 */
export function useIssueField(slug, number, opts) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (value) =>
      updateIssue(slug.value, number.value, opts.toPayload(value)),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["issue", slug.value, number.value],
      });
      queryClient.invalidateQueries({ queryKey: ["issues", slug.value] });
      if (opts.extraInvalidate) {
        for (const key of opts.extraInvalidate) {
          queryClient.invalidateQueries({ queryKey: key });
        }
      }
    },
  });
}
