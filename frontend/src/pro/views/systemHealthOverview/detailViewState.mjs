export const shouldRenderDetailBody = ({ selectedSystem, detailLoading }) =>
  Boolean(selectedSystem) && !detailLoading
