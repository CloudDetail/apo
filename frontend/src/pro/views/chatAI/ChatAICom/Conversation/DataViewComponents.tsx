import { CAccordion, CAccordionBody, CAccordionHeader, CAccordionItem } from '@coreui/react'
import { Card } from 'antd'
import React, { useEffect, useRef, useState } from 'react'
import { useChatAIContext } from 'src/pro/contexts/ChatAIContext'
import styles from './index.module.scss'
import DataViewCom from 'src/pro/components/DataViewCom'
interface DataViewComponentsProps {
  components: any[]
}

const DataViewComponents = ({ components }: DataViewComponentsProps) => {
  const dataViewType = useChatAIContext((ctx) => ctx.dataViewType)
  const causeNode = useChatAIContext((ctx) => ctx.causeNode)

  // 使用 useRef 来存储初始化值
  const [analyserNode, setAnalyserNode] = useState(null)

  useEffect(() => {
    // 在组件初始化时设置值
    if (analyserNode === null) {
      setAnalyserNode(causeNode)
    }
  }, [causeNode])
  return (
    <div className="w-full">
      {components.map((item, index) => (
        <CAccordion
          activeItemKey={item}
          className={styles.fadeItem}
          style={{ animationDelay: index * 0.2 }}
        >
          <CAccordionItem itemKey={item}>
            <CAccordionHeader>
              <span className="ml-2 h-[20px]">{dataViewType[item.type].title}</span>
            </CAccordionHeader>
            <CAccordionBody className="p-0 max-h-[500px] w-full overflow-hidden">
              {analyserNode && (
                <DataViewCom type={item.type} data={item.data} analyserNode={analyserNode} />
              )}
            </CAccordionBody>
          </CAccordionItem>
        </CAccordion>
      ))}
    </div>
  )
}

export default React.memo(DataViewComponents)
