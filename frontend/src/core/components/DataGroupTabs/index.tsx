import { useUserContext } from 'src/core/contexts/UserContext'
import { Button, Tabs } from 'antd'
import { AiOutlineSetting } from 'react-icons/ai'
import { useNavigate } from 'react-router-dom'

export default function DataGroupTabs({ children }) {
  const { dataGroupList } = useUserContext()
  const navigate = useNavigate()
  const getTabItems = () => {
    return dataGroupList.map((dataGroup) => ({
      label: dataGroup.groupName,
      key: dataGroup.groupId,
      closable: false,
      children: children(dataGroup.groupId),
    }))
  }
  return (
    <>
      <Tabs
        defaultActiveKey="1"
        items={getTabItems()}
        animated={true}
        tabBarExtraContent={
          <Button
            color="primary"
            variant="outlined"
            className="ml-3"
            icon={<AiOutlineSetting />}
            onClick={() => {
              navigate('/data-group')
            }}
          >
            管理数据组
          </Button>
        }
      />
    </>
  )
}
