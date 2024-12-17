import { Button, Flex, Input, Radio, Tabs, TabsProps } from 'antd'
import { useEffect, useState } from 'react'
import { MdAdd } from 'react-icons/md'
import CustomDataViewTab from './CustomDataViewTab'
import SlowDataTypeExample from './SlowDataTypeExample'
import { deleteCustomDataView, getCustomList } from 'src/pro/api/custom'
import streamFetch from 'src/pro/util/streamFetch'
import { useSelector } from 'react-redux'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { useChatAIContext } from 'src/pro/contexts/ChatAIContext'
import { getStep } from 'src/core/utils/step'
import React from 'react'
import { showToast } from 'src/core/utils/toast'
import ErrorDataTypeExample from './ErrorDataTypeExample'

const DataViewConfirmStep = () => {
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const analysisNode = useChatAIContext((ctx) => ctx.analysisNode)
  const causeConfirmvisibility = useChatAIContext((ctx) => ctx.causeConfirmvisibility)
  const [activeKey, setActiveKey] = useState('slow')
  const defaultItems = [
    {
      key: 'slow',
      label: '慢故障推荐数据',
      children: <SlowDataTypeExample />,
      closable: false,
      type: ['slow'],
    },
    {
      key: 'error',
      label: '错故障推荐数据',
      children: <ErrorDataTypeExample />,
      closable: false,
      type: ['log'],
    },
  ]
  const [items, setItems] = useState<any>(defaultItems)
  const addCard = () => {
    setActiveKey('custom_' + items.length)
    setItems((prev) => {
      const result = [...prev]
      result.push({
        key: 'custom_' + prev.length,
        label: '自定义' + prev.length,
        children: <CustomDataViewTab />,
        closable: true,
        newAdd: true,
      })
      return result
    })
  }
  const checkActiveKey = (deleteKey: string) => {
    if (deleteKey === activeKey) {
      setActiveKey('slow')
    }
  }
  const deleteCustom = (targetKey: string) => {
    const targetItemIndex = items.findIndex((item) => item.key === targetKey)
    const targetItem = items[targetItemIndex]
    if (targetItemIndex > -1) {
      if (targetItem.newAdd) {
        const result = [...items]
        result.splice(targetItemIndex, 1)
        setItems(result)
        checkActiveKey(targetKey)
      } else {
        deleteCustomDataView(targetItem.label)
          .then((res) => {
            showToast({
              title: ' 删除名为' + targetItem.label + '的自定义数据类型成功',
            })
            getCustomItems()
          })
          .finally(() => {
            checkActiveKey(targetKey)
          })
      }
    }
  }
  const getCustomItems = () => {
    getCustomList().then((res) => {
      const newItems = [...defaultItems]
      res.customs?.map((item) => {
        newItems.push({
          key: 'custom_' + item.name,
          label: item.name,
          children: <CustomDataViewTab defaultValue={item} />,
          closable: true,
          type: item.type,
        })
      })

      setItems(newItems)
      if (!newItems.some((item) => item.key === activeKey)) {
        setActiveKey('slow')
      }
    })
  }
  useEffect(() => {
    if (causeConfirmvisibility) getCustomItems()
  }, [causeConfirmvisibility])
  const onClick = () => {
    const type = items.find((item) => item.key === activeKey)?.type || []
    analysisNode(type)
  }
  const onEdit = (
    targetKey: React.MouseEvent | React.KeyboardEvent | string,
    action: 'add' | 'remove',
  ) => {
    if (action === 'add') {
      addCard()
    } else {
      deleteCustom(targetKey)
    }
  }
  return (
    <div className=" w-full h-full">
      <Tabs
        type="editable-card"
        items={items}
        activeKey={activeKey}
        onChange={setActiveKey}
        onEdit={onEdit}
        addIcon={
          <Flex>
            <MdAdd size={20} />
            用户自定义数据确认故障
          </Flex>
        }
      />
      <div className="absolute bottom-2 right-2">
        {(activeKey === 'slow' || activeKey === 'error') && (
          <Button type="primary" onClick={onClick}>
            确认根因并查看数据
          </Button>
        )}
      </div>
    </div>
  )
}
export default React.memo(DataViewConfirmStep)
