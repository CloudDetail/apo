import { Button, Flex, Tabs } from 'antd'
import { useEffect, useState } from 'react'
import { MdAdd } from 'react-icons/md'
import CustomDataViewTab from './CustomDataViewTab'
import SlowDataTypeExample from './SlowDataTypeExample'
import { deleteCustomDataView, getCustomList } from 'src/pro/api/custom'
import { useSelector } from 'react-redux'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { useChatAIContext } from 'src/pro/contexts/ChatAIContext'
import React from 'react'
import ErrorDataTypeExample from './ErrorDataTypeExample'
import styles from './index.module.scss'
import { useTranslation } from 'react-i18next'
import { notify } from 'src/core/utils/notify'
const i18nKey = 'causeConfirm'
const DataViewConfirmStep = () => {
  const { t } = useTranslation('pro/chatAI')
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const analysisNode = useChatAIContext((ctx) => ctx.analysisNode)
  const causeConfirmvisibility = useChatAIContext((ctx) => ctx.causeConfirmvisibility)
  const [activeKey, setActiveKey] = useState('slow')
  const defaultItems = [
    {
      key: 'slow',
      label: t(`${i18nKey}.slowFault`),
      children: <SlowDataTypeExample />,
      closable: false,
      type: ['slow'],
    },
    {
      key: 'error',
      label: t(`${i18nKey}.errorFault`),
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
        label: t(`${i18nKey}.custom`) + prev.length,
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
            notify({
              message: t(`${i18nKey}.deleteSuccess`, { name: targetItem.label }),
              type: 'success',
            }),
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
    <div className=" w-full h-full flex flex-col justify-between">
      <Tabs
        type="editable-card"
        items={items}
        activeKey={activeKey}
        onChange={setActiveKey}
        onEdit={onEdit}
        addIcon={
          <Flex className={styles.textEllipsis}>
            <MdAdd size={20} />
            {t(`${i18nKey}.customData`)}
          </Flex>
        }
      />
      <div className="text-right">
        {(activeKey === 'slow' || activeKey === 'error') && (
          <Button type="primary" onClick={onClick} className="">
            {t(`${i18nKey}.confirmView`)}
          </Button>
        )}
      </div>
    </div>
  )
}
export default React.memo(DataViewConfirmStep)
