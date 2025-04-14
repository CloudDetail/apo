import React from 'react';
import { useTranslation } from 'react-i18next';
import CommonModal from 'src/core/components/Modal/CommonModal';
import PermissionTree from 'src/core/components/PermissionTree';
import { Role } from 'src/core/types/role';

interface PermissionModalProps {
  visible: boolean;
  selectedRole: Role | null;
  onCancel: () => void;
  onSave: (checkedKeys: React.Key[]) => void;
}

export const PermissionModal: React.FC<PermissionModalProps> = ({
  visible,
  selectedRole,
  onCancel,
  onSave,
}) => {
  const { t } = useTranslation('core/roleManage');

  return (
    <CommonModal
      title={t('index.configPermission')}
      open={visible}
      onCancel={onCancel}
      width={800}
      footer={null}
    >
      {selectedRole && (
        <PermissionTree
          subjectId={selectedRole.roleId}
          subjectType="role"
          onSave={onSave}
          style={{ height: 'calc(100vh - 210px)', overflow: 'auto' }}
          styles={{ body: { padding: '8px' } }}
          actionStyle={{ paddingBlockEnd: '0px', justifyContent: 'flex-end' }}
        />
      )}
    </CommonModal>
  );
};
