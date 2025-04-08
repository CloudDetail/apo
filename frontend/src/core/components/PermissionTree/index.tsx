import React, { useState, useEffect } from 'react';
import { Tree, Button, Card } from 'antd';
import { BsCheckAll } from 'react-icons/bs';
import LoadingSpinner from 'src/core/components/Spinner';
import { useTranslation } from 'react-i18next';
import { getAllPermissionApi, getSubjectPermissionApi } from 'src/core/api/permission';

// 定义权限项的接口
interface PermissionItem {
  featureId: string;
  featureName: string;
  children?: PermissionItem[];
  [key: string]: any;
}

// 定义组件的属性接口
interface PermissionTreeProps {
  subjectId?: string | number;
  subjectType: 'role' | 'user';
  onSave?: (checkedKeys: React.Key[]) => void;
  readOnly?: boolean;
  className?: string;
}

/**
 * 通用权限树组件
 * 用于角色和用户的权限管理
 */
function PermissionTree({
  subjectId,
  subjectType,
  onSave,
  readOnly = false,
  className
}: PermissionTreeProps) {
  const [expandedKeys, setExpandedKeys] = useState<React.Key[]>([]);
  const [checkedKeys, setCheckedKeys] = useState<React.Key[]>([]);
  const [selectedKeys, setSelectedKeys] = useState<React.Key[]>([]);
  const [autoExpandParent, setAutoExpandParent] = useState<boolean>(true);
  const [permissionTreeData, setPermissionTreeData] = useState<PermissionItem[]>([]);
  const [allKeys, setAllKeys] = useState<React.Key[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const { t, i18n } = useTranslation('common');

  const onExpand = (expandedKeysValue: React.Key[]) => {
    setExpandedKeys(expandedKeysValue);
    setAutoExpandParent(false);
  };

  const onCheck = (checkedKeysValue: React.Key[] | { checked: React.Key[]; halfChecked: React.Key[] }) => {
    if (Array.isArray(checkedKeysValue)) {
      setCheckedKeys(checkedKeysValue);
    } else {
      setCheckedKeys(checkedKeysValue.checked);
    }
  };

  const onSelect = (selectedKeysValue: React.Key[]) => {
    setSelectedKeys(selectedKeysValue);
  };

  // 递归遍历树结构收集所有键和展开的键
  const loopTree = (treeData: PermissionItem[] = [], key: keyof PermissionItem = 'featureId') => {
    const allKeys: React.Key[] = [];
    const expandedKeys: React.Key[] = [];

    treeData.forEach((item) => {
      allKeys.push(item[key] as React.Key);

      if (item?.children?.length > 0) {
        expandedKeys.push(item[key] as React.Key);
        const { allKeys: allResult, expandedKeys: expandedResult } = loopTree(item.children, key);
        expandedKeys.push(...expandedResult);
        allKeys.push(...allResult);
      }
    });

    return { allKeys, expandedKeys };
  };

  const fetchData = async () => {
    if (!subjectId) return;

    setLoading(true);
    try {
      const params = { language: i18n.language };
      const [allPermissions, subjectPermissions] = await Promise.all([
        getAllPermissionApi(params),
        getSubjectPermissionApi({
          subjectId,
          subjectType,
        }),
      ]);

      setPermissionTreeData(allPermissions || []);
      const { allKeys, expandedKeys } = loopTree(allPermissions || []);

      setExpandedKeys(expandedKeys);
      setAllKeys(allKeys);
      setCheckedKeys((subjectPermissions || []).map((permission: any) => permission.featureId));
    } catch (error) {
      console.error('获取权限失败', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (subjectId) fetchData();
  }, [subjectId, i18n.language]);

  const handleSave = () => {
    if (onSave) {
      onSave(checkedKeys);
    }
  };

  return (
    <Card className={className} style={{ height: 'calc(100vh - 60px)', overflow: 'auto' }}>
      <LoadingSpinner loading={loading} />
      {!readOnly && (
        <Button
          type="primary"
          className="mx-4 mb-4"
          onClick={() => setCheckedKeys(allKeys)}
          icon={<BsCheckAll />}
        >
          {t('selectAll')}
        </Button>
      )}
      <Tree
        checkable={!readOnly}
        onExpand={onExpand}
        expandedKeys={expandedKeys}
        autoExpandParent={autoExpandParent}
        onCheck={onCheck}
        checkedKeys={checkedKeys}
        onSelect={onSelect}
        selectedKeys={selectedKeys}
        defaultExpandAll={true}
        treeData={permissionTreeData}
        fieldNames={{ title: 'featureName', key: 'featureId' }}
      />
      {!readOnly && (
        <Button type="primary" className="m-4" onClick={handleSave}>
          {t('save')}
        </Button>
      )}
    </Card>
  );
}

export default PermissionTree;