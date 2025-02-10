import { CAccordion, CAccordionBody, CAccordionHeader, CAccordionItem } from '@coreui/react'
import { Transfer, TransferProps, Image, Checkbox, Button, Form, ConfigProvider } from 'antd'
import _ from 'lodash'
import { useEffect, useState } from 'react'
import { useChatAIContext } from 'src/pro/contexts/ChatAIContext'
import styles from './index.module.scss'
import { MdDeleteOutline } from 'react-icons/md'
import { useUpdateEffect } from 'react-use'
const ComAccordionTransfer = () => {
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
  const accordion = (direction, item, index) => {
    return (
      <div
        key={item}
        draggable={direction === 'right'}
        onDragStart={() => handleDragStart(index)}
        onDragOver={(e) => handleDragOver(e, index)}
        onDrop={(e) => handleDrop(e, index)}
      >
        <CAccordion className={styles.accordion}>
          <CAccordionItem itemKey={item}>
            <CAccordionHeader className={styles.accordionHeader}>
              <span className="ml-2 h-[20px]">{dataViewType[item].title}</span>
            </CAccordionHeader>
            <CAccordionBody className="p-0">
              <Image width={'100%'} src={dataViewType[item].exampleImg} />
            </CAccordionBody>
          </CAccordionItem>
        </CAccordion>
      </div>
    )
  }
  return (
    <Transfer
      dataSource={dataViewTypeList}
      titles={['可选的数据类型', '最终展示数据类型']}
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
                <div className="flex-1 p-2 w-0">{accordion(direction, item, index)}</div>
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
