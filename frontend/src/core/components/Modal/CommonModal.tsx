import React, { ReactNode } from 'react';
import { Modal, Form, Button } from 'antd';
import { useTranslation } from 'react-i18next';

interface CommonModalProps {
  title: string;
  open: boolean;
  onCancel: () => void;
  onOk?: () => void;
  width?: number | string;
  children: ReactNode;
  footer?: ReactNode | null;
  okText?: string;
  cancelText?: string;
  confirmLoading?: boolean;
  maskClosable?: boolean;
  destroyOnClose?: boolean;
  centered?: boolean;
  className?: string;
  formId?: string;
}

/**
 * Common Ant Modal component
 */
const CommonModal: React.FC<CommonModalProps> = ({
  title,
  open,
  onCancel,
  onOk,
  width = 520,
  children,
  footer,
  okText,
  cancelText,
  confirmLoading = false,
  maskClosable = true,
  destroyOnClose = true,
  centered = true,
  className,
  formId
}) => {
  const { t } = useTranslation('common')

  // Default bottom buttons
  const defaultFooter = (
    <>
      <Button onClick={onCancel}>
        {cancelText || t('cancel')}
      </Button>
      <Button
          type="primary"
          onClick={onOk}
          loading={confirmLoading}
          htmlType={formId ? 'submit' : 'button'}
          form={formId}
      >
        {okText || t('confirm')}
      </Button>
    </>
  );

  return (
    <Modal
      title={title}
      open={open}
      onCancel={onCancel}
      width={width}
      footer={footer === undefined ? defaultFooter : footer}
      maskClosable={maskClosable}
      destroyOnClose={destroyOnClose}
      centered={centered}
      className={className}
      confirmLoading={confirmLoading}
    >
      {children}
    </Modal>
  );
};

export default CommonModal;