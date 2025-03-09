// @ts-ignore
import { Badge, Select, Tag, Tabs } from 'antd'
import Empty from 'src/core/components/Empty/Empty'
// @ts-ignore
import React, { useEffect, useState, useRef } from 'react'
import { useAlertAnalyzeContext } from 'src/pro/contexts/AlertAnalyzeContext'
import { convertTime } from 'src/core/utils/time'
// @ts-ignore
import { VariableSizeList as List } from 'react-window'
import { useSelector } from 'react-redux'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { useTranslation } from 'react-i18next'
const i18nKey = 'node'

const AnormalList = () => {
  const { t } = useTranslation('pro/dataViewCom')

  const {
    finalAnormalEvents,
    deltaAnormalEvents,
    selectedNodeInfo,
    anormalTypes,
    anormalTypesOptions,
  } = useAlertAnalyzeContext()
  const { startTime: stepStartTime, endTime: stepEndTime } = useSelector(selectSecondsTimeRange)
  const viewportHeight = window.innerHeight
  const [newAnormalList, setNewAnormalList] = useState([])
  const [updateAnormalList, setUpdateAnormalList] = useState([])
  const [resolvedAnormalList, setResolvedAnormalList] = useState([])
  const [allAnormalList, setAllAnormalList] = useState([])
  const [anormalStatus, setAnormalStatus] = useState('all')
  const [anormalType, setAnormalType] = useState([1, 2, 3, 4, 5])
  const [key, setKey] = useState(0)
  // @ts-ignore
  const [listHeight, setlistHeight] = useState(270)
  const [options, setOptions] = useState([])
  const [scroll, setScroll] = useState(false)
  const [currentList, setCurrentList] = useState([])
  const listRef = useRef({})
  const tipRef = useRef({})
  const rowHeights = useRef({})

  const getList = (anormalStatus) => {
    if (anormalStatus === 'startFiring') return newAnormalList
    if (anormalStatus === 'updatedFiring') return updateAnormalList
    if (anormalStatus === 'resolved') return resolvedAnormalList
    return allAnormalList
  }

  const itemData = getList(anormalStatus)

  const getSelectedList = (anormalStatus) => {
    if (anormalStatus === 'startFiring') return setCurrentList(newAnormalList)
    if (anormalStatus === 'updatedFiring') return setCurrentList(updateAnormalList)
    if (anormalStatus === 'resolved') return setCurrentList(resolvedAnormalList)
    return setCurrentList(allAnormalList)
  }

  const anormalStatusOptions = [
    {
      key: 'startFiring',
      label: (
        <Tag key={'startFiring'} color={'error'}>
          {t('add')} {getList('startFiring').length}
        </Tag>
      ),
    },
    {
      key: 'updatedFiring',
      label: (
        <Tag key={'updatedFiring'} color={'warning'}>
          {t('duplicate')} {getList('updatedFiring').length}
        </Tag>
      ),
    },
    {
      key: 'resolved',
      label: (
        <Tag key={'resolved'} color={'success'}>
          {t('resolved')} {getList('resolved').length}
        </Tag>
      ),
    },
    {
      key: 'all',
      label: (
        <Tag key={'all'} color={'processing'}>
          {t('currentBacklog')} {getList('all').length}
        </Tag>
      ),
    },
  ]

  const useRowChanged = ({ index, setRowHeight }) => {
    const rowRef = useRef({})

    useEffect(() => {
      if (rowRef.current) {
        // @ts-ignore
        setRowHeight(index, rowRef.current.clientHeight)
      }
    }, [rowRef])

    return {
      rowRef,
    }
  }

  const setRowHeight = (index, size) => {
    // @ts-ignore
    listRef.current.resetAfterIndex(0)
    rowHeights.current = { ...rowHeights.current, [index]: size }
  }

  const getRowHeight = (index) => (rowHeights.current[index] || 160) + 20

  const filterByAnormalType = (array) => {
    return array.filter((element) =>
      // @ts-ignore
      anormalType.some((item) => element.anormalType === parseInt(item, 10)),
    )
  }

  const Row = ({ index, style, data }) => {
    delete style.height
    const { rowRef } = useRowChanged({ index, setRowHeight })
    return (
      <div
        className="flex justify-between w-full"
        style={style}
        // @ts-ignore
        ref={rowRef}
      >
        <div className="text-xs flex-grow">
          <div>{data[index].anormalReason}</div>
          <div>
            <span className="text-gray-400">{t('appName')}：</span>
            {data[index].serviceName}
          </div>
          <div className="text-ellipsis text-wrap">
            <span className="text-gray-400">{t('serviceEndpoint')}：</span>
            {data[index].endpoint}
          </div>
          <div className="text-ellipsis text-wrap">
            <span className="text-gray-400">{t('time')}：</span>
            {convertTime(data[index].timestamp, 'yyyy-mm-dd hh:mm:ss.SSS')}
          </div>
          <div className="text-ellipsis text-wrap">
            <span className="text-gray-400">{t('details')}：</span>
            {data[index].anormalMsg}
          </div>
        </div>
        <div className="flex-grow-0">
          <div
            style={{
              display: 'flex',
              justifyContent: 'center',
              alignItems: 'center',
              position: 'relative',
              right: '20%',
            }}
          >
            <Tag>{anormalTypesOptions[data[index].anormalType].label}</Tag>
          </div>
        </div>
      </div>
    )
  }

  useEffect(() => {
    const observer = new ResizeObserver(() => {
      // 每当视口变化时，更新 `key`，强制组件重新加载
      setKey((prevKey) => prevKey + 1)
    })

    // 监听整个 `document.body` 的尺寸变化
    observer.observe(document.body)

    // 清理函数：移除观察器
    return () => observer.disconnect()
  }, [])

  useEffect(() => {
    getSelectedList(anormalStatus)
  }, [anormalStatus, itemData])

  useEffect(() => {
    const config = anormalTypesOptions
      .filter((item) => Object.values(anormalTypes).includes(item['value']))
      .filter((item) => item.key !== 0)
      .map((item) => {
        return {
          value: item.key,
          label: item.label,
          type: item.value,
        }
      })
    setOptions(config)
    const selectItems = anormalTypesOptions
      .filter((item) => Object.values(anormalTypes).includes(item['value']))
      .map((item) => {
        return item.key
      })
    setAnormalType(selectItems)
  }, [anormalTypes])

  useEffect(() => {
    if (listRef.current && currentList.length) {
      // @ts-ignore
      listRef.current.scrollTo(0) // 滚动到顶部
    }
  }, [currentList])

  useEffect(() => {
    let newAnormalList = []
    let updateAnormalList = []
    let resolvedAnormalList = []
    let allAnormalList = []

    deltaAnormalEvents.forEach((event) => {
      if (
        !selectedNodeInfo?.endpoint ||
        (selectedNodeInfo?.endpoint === event.endpoint &&
          selectedNodeInfo?.service === event.serviceName)
      ) {
        if (event.anormalStatus === 'startFiring') {
          newAnormalList.push(event)
        } else if (event.anormalStatus === 'updatedFiring') {
          updateAnormalList.push(event)
        } else {
          resolvedAnormalList.push(event)
        }
      }
    })

    finalAnormalEvents.forEach((event) => {
      if (
        !selectedNodeInfo?.endpoint ||
        (selectedNodeInfo?.endpoint === event.endpoint &&
          selectedNodeInfo?.service === event.serviceName)
      ) {
        allAnormalList.push(event)
      }
    })

    selectedNodeInfo?.endpoint ? setlistHeight(180) : setlistHeight(270)

    setNewAnormalList(filterByAnormalType(newAnormalList))
    setUpdateAnormalList(filterByAnormalType(updateAnormalList))
    setResolvedAnormalList(filterByAnormalType(resolvedAnormalList))
    setAllAnormalList(filterByAnormalType(allAnormalList))
  }, [deltaAnormalEvents, finalAnormalEvents, selectedNodeInfo, anormalType])
  return (
    <div style={{ height: viewportHeight }}>
      <div
        className="flex items-start justify-between"
        style={{ flexDirection: 'column', marginBottom: '0px' }}
      >
        <div>
          <Select
            options={options}
            mode="multiple"
            showSearch={false}
            value={anormalType}
            defaultValue={[1, 2, 3, 4, 5]}
            allowClear
            onChange={(value) => {
              setAnormalType(value)
            }}
            style={{
              width: '100%',
              minWidth: '240px',
              maxWidth: '700px',
            }}
            placeholder={t('selectErrorType')}
          />
        </div>
        <div>
          <Tabs
            items={anormalStatusOptions}
            onChange={(value) => {
              setAnormalStatus(value)
            }}
            style={{ width: '100%', minWidth: '150px' }}
            defaultActiveKey="all"
          />
        </div>
      </div>
      <div
        onClick={() => {
          setScroll(!scroll)
        }}
        key={key}
        style={{ width: '100%' }}
      >
        <div
          style={
            scroll && currentList.length
              ? {
                  border: '2px solid #7B9DFF',
                  paddingLeft: '2px',
                  paddingTop: '2px',
                }
              : {
                  border: '2px solid transparent',
                  pointerEvents: 'none',
                  paddingLeft: '2px',
                  paddingTop: '2px',
                }
          }
        >
          {currentList.length ? (
            <List
              height={viewportHeight - 310}
              width={'100%'}
              itemSize={getRowHeight}
              itemData={getList(anormalStatus)}
              itemCount={getList(anormalStatus).length}
              ref={listRef}
            >
              {Row}
            </List>
          ) : (
            <Empty />
          )}
        </div>
      </div>
    </div>
  )
}

export default AnormalList
