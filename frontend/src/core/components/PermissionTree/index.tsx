import React, { useState, useEffect } from 'react';
import { Tree, Button, Card, Flex } from 'antd';
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
  value?: React.Key[];
  defaultValue?: React.Key[];
  onChange?: (value: React.Key[]) => void;
  subjectId?: string | number;
  subjectType: 'role' | 'user';
  onSave?: (checkedKeys: React.Key[]) => void;
  className?: string;
  hasSaveButton?: boolean;
  actionsLocation?: 'top' | 'bottom';
  actionStyle?: React.CSSProperties;
  style?: React.CSSProperties;
  styles?: Record<SemanticDOM, React.CSSProperties>;
}

function PermissionTree({
  value,
  defaultValue,
  onChange,
  subjectId,
  subjectType,
  onSave,
  className,
  hasSaveButton = true,
  actionsLocation = 'bottom',
  actionStyle,
  style,
  styles
}: PermissionTreeProps) {
  const [expandedKeys, setExpandedKeys] = useState<React.Key[]>([]);
  const [internalCheckedKeys, setInternalCheckedKeys] = useState<React.Key[]>(defaultValue || []);
  const [permissionTreeData, setPermissionTreeData] = useState<PermissionItem[]>([]);
  const [allKeys, setAllKeys] = useState<React.Key[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const { t, i18n } = useTranslation('common');

  // 判断是否为受控模式
  const isControlled = value !== undefined;
  // 获取当前选中的键
  const checkedKeys = isControlled ? value : internalCheckedKeys;

  const onCheck = (checkedKeysValue: React.Key[] | { checked: React.Key[]; halfChecked: React.Key[] }) => {
    const newCheckedKeys = Array.isArray(checkedKeysValue) ? checkedKeysValue : checkedKeysValue.checked;

    if (!isControlled) {
      setInternalCheckedKeys(newCheckedKeys);
    }
    onChange?.(newCheckedKeys);
  };

  const handleSelectAll = () => {
    if (!isControlled) {
      setInternalCheckedKeys(allKeys);
    }
    onChange?.(allKeys);
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

  const fetchPermissionList = async () => {
    setLoading(true);
    try {
      const params = { language: i18n.language };
      const allPermissions = await getAllPermissionApi(params);
      setPermissionTreeData(allPermissions || []);
      const { allKeys, expandedKeys } = loopTree(allPermissions || []);
      setExpandedKeys(expandedKeys);
      setAllKeys(allKeys);
      return { allPermissions, allKeys, expandedKeys };
    } catch (error) {
      console.error('获取权限列表失败:', error);
      return { allPermissions: [], allKeys: [], expandedKeys: [] };
    } finally {
      setLoading(false);
    }
  };

  const fetchSubjectPermissions = async () => {
    if (!subjectId) return [];
    try {
      const subjectPermissions = await getSubjectPermissionApi({
        subjectId,
        subjectType,
      });
      const initialCheckedKeys = (subjectPermissions || []).map((permission: any) => permission.featureId);
      if (!isControlled) {
        setInternalCheckedKeys(initialCheckedKeys);
      }
      onChange?.(initialCheckedKeys);
      return initialCheckedKeys;
    } catch (error) {
      console.error('获取主体权限失败:', error);
      return [];
    }
  };

  useEffect(() => {
    // 总是获取权限列表
    fetchPermissionList();
    // 只有在有 subjectId 时才获取主体权限
    if (subjectId) {
      fetchSubjectPermissions();
    }
  }, [subjectId, i18n.language]);

  // Todo: 可以把使用插槽的方式进行优化
  const actionButtons = (
    <Flex justify='flex-end' className='w-full pr-8' style={actionStyle}>
      <Button
        type="primary"
        className="m-4"
        onClick={handleSelectAll}
        icon={<BsCheckAll />}
      >
        {t('selectAll')}
      </Button>
      { hasSaveButton && (
        <Button type="primary" className="m-4" onClick={() => onSave?.(checkedKeys)}>
          {t('save')}
        </Button>
      )}
    </Flex>
  )

  return (
    <>
    <Card
      className={className}
      style={{ height: 'calc(100vh - 60px)', overflow: 'auto', ...style }}
      styles={styles}
    >
      <LoadingSpinner loading={loading} />
      { actionsLocation == 'top' && actionButtons }
      <Tree
        checkable={true}
        onExpand={setExpandedKeys}
        expandedKeys={expandedKeys}
        onCheck={onCheck}
        checkedKeys={checkedKeys}
        defaultExpandAll={true}
        treeData={permissionTreeData}
        fieldNames={{ title: 'featureName', key: 'featureId' }}
      />
    </Card>
    { actionsLocation === 'bottom' && actionButtons }
    </>
  );
}

export default PermissionTree;