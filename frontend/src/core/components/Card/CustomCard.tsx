/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Card } from "antd"
import type { CardProps } from "antd/es/card";
import { ReactNode, CSSProperties } from "react";

interface CardPropsStyles {
  header?: CSSProperties;
  body?: CSSProperties;
  extra?: CSSProperties;
  title?: CSSProperties;
  actions?: CSSProperties;
  cover?: CSSProperties;
}

interface CustomCardProps extends CardProps {
  children: ReactNode;
  styleType?: "alerts" | "system";
}

export default function CustomCard({
  children,
  styleType,
  style,
  styles,
  ...restProps
}: CustomCardProps): ReactNode {
  const headHeight = import.meta.env.VITE_APP_CODE_VERSION === 'CE' ? '60px' : '100px';
  const defaultStyles: CardPropsStyles = styleType === 'alerts'
    ? {
        body: {
          height: '100%',
          overflow: 'hidden',
          display: 'flex',
          flexDirection: 'column',
          padding: '12px 24px',
        },
      }
    : {};
  
  return (
    <Card
      {...restProps}
      style={{  // Style of outermost container
        height: 'calc(100vh - ' + headHeight +')',
        ...style,
      }}
      styles={{  // Styles of internal sub-components
        ...defaultStyles,
        ...styles,
      }}
    >
      {children}
    </Card>
  )
}