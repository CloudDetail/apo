import { Button, Card, Steps } from 'antd'
import { useEffect, useState } from 'react'
import CauseConfirmStep from './CauseConfirmStep'
import DataViewConfirmStep from './DataViewConfirmStep/index'
import { FaChevronRight } from 'react-icons/fa'
import React from 'react'
import { useChatAIContext } from 'src/pro/contexts/ChatAIContext'
import GlassCard from 'src/pro/components/GlassCard'
import styles from './index.module.scss'
const CauseConfirm = () => {
  const causeConfirmvisibility = useChatAIContext((ctx) => ctx.causeConfirmvisibility)
  const setCauseConfirmvisibilit = useChatAIContext((ctx) => ctx.setCauseConfirmvisibilit)
  const suspectedNodeList = useChatAIContext((ctx) => ctx.suspectedNodeList)
  const selectedSuspectedNode = useChatAIContext((ctx) => ctx.selectedSuspectedNode)
  const [current, setCurrent] = useState(0)
  const [stepItems, setStepItems] = useState()
  const steps = [
    {
      title: '选择疑似根因节点',
      content: <CauseConfirmStep nextStep={() => setCurrent(1)} />,
    },
    {
      title: '根据疑似根因节点数据确认故障根因',
      content: <DataViewConfirmStep />,
    },
  ]
  const onChangeStep = (current: number) => {
    setCurrent(current)
    // carouselRef.current?.goTo(current)
  }
  useEffect(() => {
    const items = steps.map((item, index) => {
      let status, description
      if (index === 0) {
        status = selectedSuspectedNode?.service ? 'finish' : 'wait'
      }
      return { key: item.title, title: item.title, status, description }
    })
    setStepItems(items)
  }, [selectedSuspectedNode])
  useEffect(() => {
    setCurrent(current)
  }, [causeConfirmvisibility])
  return (
    <>
      <Card
        title="确认告警故障根因"
        extra={
          <Button
            onClick={() => setCauseConfirmvisibilit(false)}
            // className=" absolute top-0 right-0"
            style={{ zIndex: 100 }}
            type="link"
          >
            收起
            <FaChevronRight size={25} />
          </Button>
        }
        className={`${styles.causeConfirmCard} ${causeConfirmvisibility ? styles.fadeIn : styles.fadeOut}`}
        classNames={{ body: 'py-0' }}
      >
        {suspectedNodeList?.length === 0 && <GlassCard content={<p>未分析出可选择的疑似节点</p>} />}

        <Steps
          current={current}
          items={stepItems}
          onChange={onChangeStep}
          type="navigation"
          className={styles.steps}

          // size="small"
        />
        <div
          className="grid h-full "
          style={{
            gridTemplateColumns: '1fr 30px 1fr', // 左右均分，中间固定宽度
            gap: '16px', // 左中右间距
          }}
        >
          {/* 左侧卡片 */}
          <div className="h-full w-full">
            <Card
              className="h-full bg-[#212631]"
              styles={{
                body: {
                  height: '100%',
                  display: 'flex',
                  flexDirection: 'column',
                  padding: '16px 8px',
                },
              }}
            >
              {steps[0].content}
            </Card>
          </div>

          {/* 中间图标 */}
          <div className="flex items-center justify-center">
            <FaChevronRight size={30} />
          </div>

          {/* 右侧卡片 */}
          <div className="h-full overflow-hidden">
            <Card
              className="h-full bg-[#212631]"
              styles={{
                body: {
                  height: '100%',
                  display: 'flex',
                  flexDirection: 'column',
                  width: '100%',
                  padding: '8px',
                },
              }}
            >
              {steps[1].content}
            </Card>
          </div>
        </div>

        {/* <Carousel
          afterChange={(currentSlide) => setCurrent(currentSlide)}
          ref={carouselRef}
          // adaptiveHeight
        >
          {steps.map((step) => (
            <>{step.content}</>
          ))}
        </Carousel> */}
      </Card>
    </>
  )
}
export default React.memo(CauseConfirm)
