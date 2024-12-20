import { Button, Card, Tree } from 'antd'
import Checkbox from 'antd/es/checkbox/Checkbox'
import { useEffect, useState } from 'react'
import { BsCheckAll } from 'react-icons/bs'
import { configMenuApi, getAllPermissionApi } from 'src/core/api/permission'
import LoadingSpinner from 'src/core/components/Spinner'
import { useUserContext } from 'src/core/contexts/UserContext'
import { showToast } from 'src/core/utils/toast'

function MenuManagePage() {
  const { menuItems, getUserPermission } = useUserContext()
  const [expandedKeys, setExpandedKeys] = useState([])
  const [checkedKeys, setCheckedKeys] = useState([])
  const [selectedKeys, setSelectedKeys] = useState([])
  const [autoExpandParent, setAutoExpandParent] = useState(true)
  const [permissionTreeData, setPermissionTreeData] = useState([])
  const [allKeys, setAllKeys] = useState([])
  const [loading, setLoading] = useState(true)
  const onExpand = (expandedKeysValue) => {
    setExpandedKeys(expandedKeysValue)
    setAutoExpandParent(false)
  }
  const onCheck = (checkedKeysValue) => {
    setCheckedKeys(checkedKeysValue)
  }
  const onSelect = (selectedKeysValue, info) => {
    setSelectedKeys(selectedKeysValue)
  }

  const loopTree = (treeData = [], key = 'featureId') => {
    const allKeys = []
    const expandedKeys = []

    treeData.forEach((item) => {
      allKeys.push(item[key])

      // 如果有子节点，记录到 expandedKeys
      if (item?.children?.length > 0) {
        expandedKeys.push(item[key])

        const { allKeys: allResult, expandedKeys: expandedResult } = loopTree(item.children, key)
        expandedKeys.push(...expandedResult)
        allKeys.push(...allResult)
      }
    })
    return { allKeys, expandedKeys }
  }
  // 获取所有叶子节点
  const loopTreeForLeaf = (treeData = [], key = 'featureId') => {
    const leafKeys = []

    treeData.forEach((item) => {
      // 如果有子节点，记录到 expandedKeys
      if (item?.children?.length > 0) {
        const keyResult = loopTreeForLeaf(item.children, key)
        leafKeys.push(...keyResult)
      } else {
        leafKeys.push(item[key])
      }
    })
    return leafKeys
  }
  const getAllPermission = () => {
    setLoading(true)
    getAllPermissionApi()
      .then((res) => {
        setPermissionTreeData(res || [])
        // 展开所有节点
        const { allKeys, expandedKeys } = loopTree(res || [])

        setExpandedKeys(expandedKeys)
        setAllKeys(allKeys)
        //勾选已有权限（目前仅菜单
      })
      .finally(() => {
        setLoading(false)
      })
  }
  useEffect(() => {
    const leafKeys = loopTreeForLeaf(menuItems, 'itemId')
    setCheckedKeys(leafKeys)
  }, [menuItems])
  useEffect(() => {
    getAllPermission()
  }, [])

  //保存配置
  function configMenu() {
    setLoading(true)
    const params = new URLSearchParams()
    checkedKeys.forEach((value) => params.append('permissionList', value))
    configMenuApi(params)
      .then((res) => {
        showToast({
          title: '菜单配置成功',
          color: 'success',
        })
      })
      .catch((error) => {
        console.error(error)
      })
      .finally(() => {
        getUserPermission()
        setLoading(false)
      })
  }
  return (
    <>
      <Card style={{ height: 'calc(100vh - 60px)' }}>
        <LoadingSpinner loading={loading} />
        <Button
          type="primary"
          className="mx-4 mb-4"
          onClick={() => setCheckedKeys(allKeys)}
          icon={<BsCheckAll />}
        >
          选择全部
        </Button>
        <Tree
          checkable
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
        <Button type="primary" className="m-4" onClick={configMenu}>
          保存
        </Button>
      </Card>
    </>
  )
}
export default MenuManagePage
