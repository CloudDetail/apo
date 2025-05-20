export const SLOT_TYPES = {
  LOADING: Symbol('CardLoading'),
  HEADER: Symbol('CardHeader'),
  TABLE: Symbol('CardTable'),
  MODAL: Symbol('CardModal')
}

export const CardHeader: React.FC<React.PropsWithChildren> & { slotType?: symbol } = ({ children }) => {
  return <>{children}</>;
};
CardHeader.slotType = SLOT_TYPES.HEADER;

export const CardTable: React.FC<React.PropsWithChildren> & { slotType?: symbol } = ({ children }) => {
  return <>{children}</>;
};
CardTable.slotType = SLOT_TYPES.TABLE;