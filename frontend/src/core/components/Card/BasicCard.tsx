import React, { ReactElement } from 'react';
import { SLOT_TYPES, CardTable, CardHeader } from './CardSlots';
import { Card, Space } from 'antd';

type CardProps = {
  children: React.ReactNode;
};

export const BasicCard: React.FC<CardProps> & {
  // Loading: typeof CardLoading;
  Header: typeof CardHeader;
  Table: typeof CardTable;
  // Modal: typeof CardModal;
} = ({ children }) => {
  let headerContent: ReactElement[] = [];
  let tableContent: ReactElement | null = null;
  // let modalContent: ReactElement | null = null;
  const otherContent: ReactElement[] = [];

  //@ts-ignore
  const headHeight = import.meta.env.VITE_APP_CODE_VERSION === 'CE' ? 'var(--ce-app-head-height)' : 'var(--ee-app-head-height)';

  React.Children.forEach(children, child => {
    if (!React.isValidElement(child)) return;

    const slotType = (child.type as any)?.slotType;

    switch (slotType) {
      case SLOT_TYPES.HEADER:
        headerContent.push(child);
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
      {/* Header Section */}
      {headerContent.length > 0 && headerContent.map((header, index) => (
        <div className='w-full text-sm font-medium flex items-center justify-between' key={index}>
          {header}
        </div>
      ))}

      {/* Table Section */}
      {tableContent && <div className="flex-1 overflow-auto">
        <div className="h-full text-xs justify-between">
          {tableContent && <>{tableContent}</>}
        </div>
      </div>}

      {/* Other Content */}
      {otherContent.length > 0 && otherContent.map((content, index) => (
          <React.Fragment key={index}>{content}</React.Fragment>
        ))
      }
    </Card>
  );
};

BasicCard.Header = CardHeader;
BasicCard.Table = CardTable;
