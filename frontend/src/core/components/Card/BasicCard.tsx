// Card.tsx
import React, { ReactElement } from 'react';
import { SLOT_TYPES, CardLoading, CardFilter, CardAction, CardTable, CardModal } from './CardSlots';
import { Card, Space } from 'antd';

type CardProps = {
  children: React.ReactNode;
};

export const BasicCard: React.FC<CardProps> & {
  // Loading: typeof CardLoading;
  Filter?: typeof CardFilter;
  Action?: typeof CardAction;
  Table: typeof CardTable;
  // Modal: typeof CardModal;
} = ({ children }) => {
  // let loadingContent: ReactElement | null = null;
  let filterContent: ReactElement | null = null;
  let actionContent: ReactElement | null = null;
  let tableContent: ReactElement | null = null;
  // let modalContent: ReactElement | null = null;
  const otherContent: ReactElement[] = [];

  //@ts-ignore
  const headHeight = import.meta.env.VITE_APP_CODE_VERSION === 'CE' ? 'var(--ce-app-head-height)' : 'var(--ee-app-head-height)';

  React.Children.forEach(children, child => {
    if (!React.isValidElement(child)) return;

    const slotType = (child.type as any).slotType;

    switch (slotType) {
      case SLOT_TYPES.FILTER:
        filterContent = child;
        break;
      case SLOT_TYPES.ACTION:
        actionContent = child;
        break;
      case SLOT_TYPES.TABLE:
        tableContent = child;
        break;
      default:
        otherContent.push(child);
        break;
    }
  });

  return (
    <Card
      style={{
        height: 'calc(100vh - ' + headHeight +')',
      }}
      styles={{
        body: {
          height: '100%',
          overflow: 'hidden',
          display: 'flex',
          flexDirection: 'column',
          padding: '12px 24px',
        },
      }}
    >

      <div className="flex items-center justify-between text-sm border-b font-medium border-[var(--ant-color-border)]">
        <Space className="flex-grow w-full">
          {filterContent && <>{filterContent}</>}
        </Space>
        <Space className="flex-grow w-full">
          {actionContent && <>{actionContent}</>}
        </Space>
      </div>

      <div className="flex-1 overflow-auto">
        <div className="h-full text-xs justify-between">
          {tableContent && <>{tableContent}</>}
        </div>
      </div>

      <div className="card-other">{otherContent}</div>
    </Card>
  );
};

BasicCard.Filter = CardFilter;
BasicCard.Action = CardAction;
BasicCard.Table = CardTable;
