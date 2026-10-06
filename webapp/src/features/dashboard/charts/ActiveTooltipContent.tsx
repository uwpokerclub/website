import { DefaultTooltipContent, TooltipProps } from "recharts";
import type { NameType, ValueType } from "recharts/types/component/DefaultTooltipContent";

/**
 * Recharts' default tooltip, rendered only while it is showing.
 *
 * Recharts keeps an inactive tooltip in the DOM — hidden, but still laid out at its
 * last position. When the chart then narrows (a rotated phone, a collapsed sidenav),
 * that invisible box sits past the card's edge and the page scrolls sideways. With
 * no content while inactive, the wrapper has nothing to lay out.
 */
export function ActiveTooltipContent(props: TooltipProps<ValueType, NameType>) {
  return props.active ? <DefaultTooltipContent {...props} /> : null;
}
