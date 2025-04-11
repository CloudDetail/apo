import { Collapse } from 'antd'
import React, { useEffect, useState } from 'react'
import { useChatAIContext } from 'src/pro/contexts/ChatAIContext'
import styles from './index.module.scss'
import DataViewCom from 'src/pro/components/DataViewCom'
import { SlArrowDown } from 'react-icons/sl'
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
  const panelStyle = {
    marginBottom: 12,
    border: 'none',
  }
  return (
    <div className="w-full">
      {components.map((item, index) => (
        <Collapse
          className={styles.collapse}
          bordered={false}
          defaultActiveKey={[item.type]}
          expandIconPosition={'end'}
          items={[
            {
              key: item.type,
              forceRender: true,
              label: dataViewType[item.type].title,
              children: analyserNode && (
                <DataViewCom type={item.type} data={item.data} analyserNode={analyserNode} />
              ),
              style: panelStyle,
            },
          ]}
          expandIcon={({ isActive }) => {
            return isActive ? (
              <SlArrowDown className={'rotate-180'} size={18} />
            ) : (
              <SlArrowDown className={'rotate-0'} size={18} />
            )
          }}
        />
      ))}
    </div>
  )
}

export default React.memo(DataViewComponents)
