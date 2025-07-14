import React, { useCallback, useEffect, useRef, useState } from 'react'
import { Handle, Position } from 'reactflow'
import { BiCctv } from 'react-icons/bi'
import { Badge, theme, Tooltip } from 'antd'
import { useUpdateEffect } from 'react-use'
import _ from 'lodash'
import { useServiceAnalyzeContext } from '../contexts/ServiceAnalyzeContext'
import { LuShieldOff } from 'react-icons/lu'
import { useTranslation } from 'react-i18next'

const handleStyle = { left: 10 }

const AlertAnalyzeNode = React.memo((props) => {
  const { id, data, isConnectable } = props
  const { t: ct } = useTranslation('common')
  const {
    timeLineChartData,
    topologyNodesMap,
    anormalTypes,
    anormalTypesOptions,
    marklineTimestamp,
    setSelectedNodeInfo,
    selectedNodeInfo,
    isPlaying,
  } = useServiceAnalyzeContext()
  const [prevSumCount, setPrevSumCount] = useState(0)
  const onChange = useCallback((evt) => {}, [])
  const nodeRef = useRef(null)
  const [typeMap, setTypeMap] = useState(null)
  const getCount = (type) => {
    return topologyNodesMap?.[id]?.[type] || 0
  }

  //获取节点错误总数
  const getErrorSumCount = () => {
    return topologyNodesMap?.[id]
      ? Object.values(topologyNodesMap?.[id]).reduce(
          (accumulator, currentValue) => accumulator + currentValue,
          0,
        )
      : 0
  }

  useUpdateEffect(() => {
    const div = nodeRef.current
    const currentSumCount = getErrorSumCount()
    if (div && isPlaying) {
      const animationClass =
        currentSumCount >= prevSumCount ? 'ripple-active' : 'ripple-active-resolve'

      // 添加动画 class
      div.classList.add(animationClass)

      // 动画结束后移除 class，确保可以重复触发
      const handleAnimationEnd = () => div.classList.remove(animationClass)
      div.addEventListener('animationend', handleAnimationEnd, { once: true })
    }
    // 更新 prevSumCount
    setPrevSumCount(currentSumCount)
  }, [typeMap])

  useEffect(() => {
    let nodeErrorType = topologyNodesMap?.[id]
    if (!_.isEqual(nodeErrorType, typeMap)) {
      setTypeMap(nodeErrorType)
    }
  }, [topologyNodesMap?.[id]])
  const clickNode = () => {
    console.log(data)
    if (
      selectedNodeInfo?.service === data.service &&
      selectedNodeInfo?.endpoint === data.endpoint
    ) {
      setSelectedNodeInfo({ service: '', endpoint: '' })
    } else if (!data?.outOfGroup) {
      setSelectedNodeInfo({ service: data.service, endpoint: data.endpoint })
    }
  }
  const { useToken } = theme
  const { token } = useToken()
  return (
    <div
      ref={nodeRef}
      className={
        selectedNodeInfo.service === data.service && selectedNodeInfo.endpoint === data.endpoint
          ? 'selectedAlertAnalyzeNode '
          : 'alertAnalyzeNode'
      }
      onClick={clickNode}
    >
      <Handle
        type="target"
        position={Position.Left}
        isConnectable={isConnectable}
        className="invisible"
      />
      <Tooltip title={data.outOfGroup ? ct('outOfGroup') : ''}>
        <div
          className="nodeContent w-[240px] min-h-[60px] "
          style={{
            backgroundColor: token.colorBgLayout,
            color: data?.outOfGroup ? token.colorTextDisabled : token.colorPrimaryText,
            borderColor: data?.outOfGroup ? token.colorTextDisabled : token.colorPrimaryText,
          }}
        >
          <div className="absolute top-0 left-0">
            {' '}
            <BiCctv size={35} color={data.isTraced ? '#80ce8dff' : '#a1a1a1'} />
          </div>
          <div className="text-center text-lg pt-2 px-2">
            {data.label}
            {data.outOfGroup && (
              <div className="absolute top-0 right-0">
                <LuShieldOff size={35} />
              </div>
            )}
            <div
              className="text-xs  text-left  break-all"
              style={{ color: token.colorTextSecondary }}
            >
              {data.endpoint}
            </div>
            <div className="w-full text-left">
              {anormalTypes?.map((type) => {
                let typeItem = anormalTypesOptions.find((item) => item.value === type)
                return (
                  <div>
                    <span className="text-sm" style={{ color: token.colorText }}>
                      {typeItem.label}
                    </span>
                    <Badge
                      count={getCount(typeItem.key)}
                      showZero
                      color={getCount(typeItem.key) > 0 ? '#dc4446' : 'gray'}
                    />
                  </div>
                )
              })}
            </div>
          </div>
        </div>
      </Tooltip>
      <Handle
        type="source"
        position={Position.Right}
        id="b"
        isConnectable={isConnectable}
        className="invisible"
      />
    </div>
  )
})

export default AlertAnalyzeNode
