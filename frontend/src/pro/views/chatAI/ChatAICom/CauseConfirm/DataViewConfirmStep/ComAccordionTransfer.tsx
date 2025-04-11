import { Transfer, TransferProps, Image, Checkbox, Button, Form, Collapse } from 'antd'
import _ from 'lodash'
import { useEffect, useState } from 'react'
import { useChatAIContext } from 'src/pro/contexts/ChatAIContext'
import { MdDeleteOutline } from 'react-icons/md'
import { useUpdateEffect } from 'react-use'
import { useTranslation } from 'react-i18next'
import { SlArrowDown } from 'react-icons/sl'
import styles from './index.module.scss'
const i18nKey = 'causeConfirm'
const ComAccordionTransfer = () => {
  const { t } = useTranslation('pro/chatAI')
  const form = Form.useFormInstance()
  const dataViewType = useChatAIContext((ctx) => ctx.dataViewType)
  const dataViewTypeList = useChatAIContext((ctx) => ctx.dataViewTypeList)
  const [targetKeys, setTargetKeys] = useState<React.Key[]>([])
  const [sourceKeys, setSourceKeys] = useState<React.Key[]>(
    dataViewTypeList.map((item) => item.key),
  )
  const [draggingIndex, setDraggingIndex] = useState(null)
  const [selectedKeys, setSelectedKeys] = useState<React.Key[]>([])
  const handleChange: TransferProps['onChange'] = (newTargetKeys, direction, moveKeys) => {
    if (direction === 'left') {
      setTargetKeys(newTargetKeys)
    } else {
      setTargetKeys([...targetKeys, ...moveKeys])
    }
    setSourceKeys(_.difference(sourceKeys, moveKeys))
    setSelectedKeys(_.difference(newTargetKeys, moveKeys))
  }

  const handleDragStart = (index) => {
    setDraggingIndex(index)
  }
  const handleDragOver = (e, index) => {
    e.preventDefault()
    e.dataTransfer.dropEffect = 'move'
  }

  const handleDrop = (e, index) => {
    e.preventDefault()
    const updatedData = [...targetKeys]
    const draggingItem = updatedData.splice(draggingIndex, 1)[0]
    updatedData.splice(index, 0, draggingItem)
    setTargetKeys(updatedData)
    setDraggingIndex(null)
  }
  const onChange = (checked, key) => {
    setSelectedKeys((prev) => {
      const result = [...prev]
      if (checked) {
        result.push(key)
      } else {
        return result.filter((item) => item !== key)
      }
      return result
    })
  }

  useUpdateEffect(() => {
    const oldComponents = form.getFieldValue('type')

    if (!_.isEqual(oldComponents, targetKeys)) {
      form.setFieldValue('type', targetKeys)
    }
  }, [targetKeys])
  useEffect(() => {
    let newTargetKeys = form.getFieldValue('type')
    if (newTargetKeys && targetKeys?.length === 0) {
      setSelectedKeys([])
      setSourceKeys(_.difference(sourceKeys, newTargetKeys))
      setTargetKeys(newTargetKeys)
    }
  }, [form.getFieldValue('type')])
  const panelStyle = {
    marginBottom: 12,
    border: 'none',
  }
  const accordion = (direction, item, index) => {
    return (
      <div
        key={item}
        draggable={direction === 'right'}
        onDragStart={() => handleDragStart(index)}
        onDragOver={(e) => handleDragOver(e, index)}
        onDrop={(e) => handleDrop(e, index)}
        className="h-full flex items-center grow"
      >
        <Collapse
          className={styles.collapse}
          bordered={false}
          defaultActiveKey={[item]}
          expandIconPosition={'end'}
          items={[
            {
              key: item,
              forceRender: true,
              label: dataViewType[item].title,
              children: <Image width={'100%'} src={dataViewType[item].exampleImg} />,
              style: panelStyle,
            },
          ]}
          expandIcon={({ isActive }) => {
            return isActive ? (
              <SlArrowDown className={'rotate-180'} size={10} />
            ) : (
              <SlArrowDown className={'rotate-0'} size={10} />
            )
          }}
        />
      </div>
    )
  }
  return (
    <Transfer
      dataSource={dataViewTypeList}
      titles={[t(`${i18nKey}.availableDataTypes`), t(`${i18nKey}.finalDisplayedDataTypes`)]}
      targetKeys={targetKeys}
      selectedKeys={selectedKeys}
      onChange={handleChange}
      showSelectAll={false}
      render={(item) => item.title}
      oneWay
      style={{ marginBottom: 16 }}
      className="w-full"
    >
      {(props) => {
        const { direction, onItemRemove } = props
        return (
          <div className="h-[180px] overflow-auto">
            {(direction === 'left' ? sourceKeys : targetKeys).map((item, index) => (
              <div key={item} className="w-full flex items-center overflow-hidden">
                {direction === 'left' && (
                  <Checkbox
                    className="mx-2"
                    onChange={(e) => onChange(e.target.checked, item)}
                  ></Checkbox>
                )}
                <div className="flex-1 w-0 pl-2">{accordion(direction, item, index)}</div>
                {direction === 'right' && (
                  <Button
                    color="danger"
                    variant="text"
                    size="small"
                    className="mr-0 flex-shrink-0 flex-grow-0"
                    icon={<MdDeleteOutline />}
                    onClick={(e) => {
                      onItemRemove([item])
                      let newSourceKeys = sourceKeys
                      newSourceKeys.push(item)
                      setSourceKeys(newSourceKeys)
                    }}
                  ></Button>
                )}
              </div>
            ))}
          </div>
        )
      }}
    </Transfer>
  )
}
export default ComAccordionTransfer
