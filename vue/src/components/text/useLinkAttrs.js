// Shared link behaviour for components/text/Link.vue and InlineLink.vue, which
// render the same thing (a <RouterLink> for `to`, an <a> for `href`) and differ
// only in styling.

import { computed } from "vue";

/** Props every link component accepts. Spread into defineProps(). */
export const linkProps = {
  href: { type: String, default: "" },
  to: { type: String, default: "" },
  target: { type: String, default: undefined },
  rel: { type: String, default: undefined },
};

/**
 * `rel` for the rendered <a>: an explicit `rel` prop wins, otherwise
 * target="_blank" links get noopener/noreferrer so the opened page cannot
 * reach back through window.opener.
 */
export function useComputedRel(props) {
  return computed(() => {
    if (props.rel !== undefined) return props.rel;
    if (props.target === "_blank") return "noopener noreferrer";
    return undefined;
  });
}
