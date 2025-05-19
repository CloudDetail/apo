export const SLOT_TYPES = {
  LOADING: Symbol('CardLoading'),
  FILTER: Symbol('CardFilter'),
  ACTION: Symbol('CardAction'),
  TABLE: Symbol('CardTable'),
  MODAL: Symbol('CardModal')
}

export const CardLoading: React.FC<React.PropsWithChildren> & { slotType?: symbol } = ({ children }) => {
  return <>{children}</>;
};
CardLoading.slotType = SLOT_TYPES.LOADING;

export const CardFilter: React.FC<React.PropsWithChildren> & { slotType?: symbol } = ({ children }) => {
  return <>{children}</>;
};
CardFilter.slotType = SLOT_TYPES.FILTER;

export const CardAction: React.FC<React.PropsWithChildren> & { slotType?: symbol } = ({ children }) => {
  return <>{children}</>;
};
CardAction.slotType = SLOT_TYPES.ACTION;

export const CardTable: React.FC<React.PropsWithChildren> & { slotType?: symbol } = ({ children }) => {
  return <>{children}</>;
};
CardTable.slotType = SLOT_TYPES.TABLE;

export const CardModal: React.FC<React.PropsWithChildren> & { slotType?: symbol } = ({ children }) => {
  return <>{children}</>;
};
CardModal.slotType = SLOT_TYPES.MODAL;