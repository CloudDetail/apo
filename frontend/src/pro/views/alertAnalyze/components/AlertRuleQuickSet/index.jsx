import React, { useEffect, useState } from 'react'
import { Modal, Form, Input, Select, Card, Checkbox, Flex } from 'antd'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { getMetricsApi, addMutationRule } from 'src/pro/api/alertAnalyze'
import { useSelector } from 'react-redux'
import { getStep } from 'src/core/utils/step'
import { showToast } from 'src/core/utils/toast'
const { TextArea } = Input
const { Option } = Select

const AlertRuleQuickSet = ({ modalVisibility, closeModal }) => {
  //获取时间信息
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const step = getStep(startTime, endTime)

  //保存group信息
  const [group, setGroup] = useState('')
  const [groupUpper, setGroupUpper] = useState('')
  const [groupLower, setGroupLower] = useState('')

  //获取表单实例
  const [form] = Form.useForm()

  //告警等待时间单位
  const [unit, setUnit] = useState('s')

  //指标列表
  const [metricsList, setMetricsList] = useState([])

  //metrics数据处理
  const handleMetricsList = (metrics) => {
    const newMetricsList = []

    metrics.forEach((metric) => {
      // 查找是否存在该 group
      const existingGroup = newMetricsList.find((group) => group.title === metric.group)

      if (existingGroup) {
        // 如果存在，直接添加到 options 中
        existingGroup.options.push({
          label: (
            <span>
              {metric.name} <span style={{ color: '#fc6f03' }}>{metric.describe}</span>
            </span>
          ),
          value: `${metric.name}+${metric.describe}+${metric.customMetric}+${metric.group}`,
        })
      } else {
        // 如果不存在，创建新的 group
        newMetricsList.push({
          label: <span>{metric.group}</span>,
          title: metric.group,
          options: [
            {
              label: (
                <span>
                  {metric.name} <span style={{ color: '#fc6f03' }}>{metric.describe}</span>
                </span>
              ),
              value: `${metric.name}+${metric.describe}+${metric.customMetric}+${metric.group}`,
            },
          ],
        })
      }
    })

    return newMetricsList
  }

  //点击保存按钮处理逻辑
  const handleOk = async () => {
    const uploadData = await form.validateFields()
    uploadData.for = uploadData.for + unit
    uploadData.mutationCheckPQL = {
      customMetric: uploadData.mutationCheckPQL,
      group: group,
    }
    uploadData.mutationLowerLimit = uploadData.mutationLowerLimit
      ? {
          customMetric: uploadData.mutationLowerLimit,
          group: groupLower,
        }
      : undefined
    uploadData.mutationUpperLimit = uploadData.mutationUpperLimit
      ? {
          customMetric: uploadData.mutationUpperLimit,
          group: groupUpper,
        }
      : undefined
    uploadData.startTime = startTime
    uploadData.endTime = endTime
    uploadData.step = step
    const result = await addMutationRule(uploadData)
    if (result == 'ok') {
      showToast({
        title: '添加异常检测配置成功',
        color: 'success',
      })
      closeModal()
    }
  }

  //获取快速指标列表
  const getMetrics = async () => {
    try {
      let { metrics } = await getMetricsApi()
      metrics = handleMetricsList(metrics)
      setMetricsList(metrics)
    } catch (err) {
      console.error('Error fetching metrics:', err)
    }
  }

  //处理指标选择改变
  const handleChange = (value) => {
    console.log(value)
    const customMetric = value.split('+')[2]
    setGroup(value.split('+')[3])
    form.setFieldValue('mutationCheckPQL', customMetric)
  }

  //处理上界指标改变
  const handleChangeUpper = (value) => {
    const customMetric = value.split('+')[2]
    setGroupUpper(value.split('+')[3])
    form.setFieldValue('mutationUpperLimit', customMetric)
  }

  //处理下界指标改变
  const handleChangeLower = (value) => {
    const customMetric = value.split('+')[2]
    setGroupLower(value.split('+')[3])
    form.setFieldValue('mutationLowerLimit', customMetric)
  }

  //时间单位选择
  const selectAfter = (
    // <Form.Item
    //     name={"unit"}
    //     style={{
    //         height:"7px"
    //     }}
    // >
    <Select style={{ width: '90px' }} value={unit} onChange={(value) => setUnit(value)}>
      <Option value="s">s (秒)</Option>
      <Option value="m">m (分钟)</Option>
      <Option value="h">h (小时)</Option>
      <Option value="d">d (天)</Option>
    </Select>
    // </Form.Item>
  )
  //初始化
  useEffect(() => {
    form.resetFields()
    getMetrics()
  }, [modalVisibility])

  return (
    <Modal
      open={modalVisibility}
      onCancel={closeModal}
      maskClosable={false}
      onOk={handleOk}
      width={800}
      okText={'保存'}
      cancelText={'取消'}
      title={'异常检测配置'}
      centered
      destroyOnClose
    >
      <div
        style={{
          width: '95%',
          height: '71vh',
          display: 'flex',
          flexDirection: 'column',
          justifyContent: 'start',
          alignItems: 'center',
          marginTop: '30px',
        }}
      >
        <Form
          form={form}
          style={{
            width: '100%',
          }}
        >
          <div
            style={{
              width: '100%',
              display: 'flex',
              justifyContent: 'space-between',
            }}
          >
            <div
              style={{
                width: '60%',
                display: 'flex',
                flexDirection: 'column',
              }}
            >
              <Form.Item
                label="突变检测名"
                name="name"
                rules={[
                  {
                    required: true,
                    message: '请填写突变检测名',
                  },
                ]}
              >
                <Input />
              </Form.Item>
            </div>
            <Form.Item
              label="告警时间"
              name={'for'}
              rules={[
                {
                  required: true,
                  message: '请输入告警时间',
                },
              ]}
              style={{
                width: '35%',
              }}
            >
              <Input addonAfter={selectAfter} />
            </Form.Item>
          </div>
          <Card
            title={'检测语句'}
            style={{
              height: 'auto',
              paddingBottom: '40px',
            }}
          >
            <Flex vertical>
              <p
                style={{
                  marginBottom: '5px',
                }}
              >
                <span style={{ color: 'red' }}>* </span>告警指标
              </p>
              <Select
                // mode="tags"
                maxCount={1}
                placeholder="请选择指标"
                options={metricsList}
                style={{
                  width: '100%',
                  marginBottom: '20px',
                }}
                onSelect={handleChange}
              />
              <Form.Item
                name="mutationCheckPQL"
                layout="vertical"
                rules={[
                  {
                    required: true,
                    message: '请选择快速指标',
                  },
                ]}
              >
                <TextArea />
              </Form.Item>
            </Flex>
            <Flex vertical>
              <p
                style={{
                  marginTop: '50px',
                  marginBottom: '5px',
                }}
              >
                突变上界
                <span style={{ fontSize: '12px', color: '#5286E8' }}>
                  {' '}
                  (指标大于该值时触发报警)
                </span>
              </p>
              <Select
                // mode="tags"
                maxCount={1}
                placeholder="请选择指标"
                options={metricsList}
                style={{
                  width: '100%',
                  marginBottom: '20px',
                }}
                onSelect={handleChangeUpper}
              />
              <Form.Item name="mutationUpperLimit" layout="vertical">
                <TextArea />
              </Form.Item>
            </Flex>
            <Flex vertical>
              <p
                style={{
                  marginTop: '50px',
                  marginBottom: '5px',
                }}
              >
                突变下界
                <span style={{ fontSize: '12px', color: '#5286E8' }}>
                  {' '}
                  (指标小于该值时触发报警)
                </span>
              </p>
              <Select
                // mode="tags"
                maxCount={1}
                placeholder="请选择指标"
                options={metricsList}
                style={{
                  width: '100%',
                  marginBottom: '20px',
                }}
                onSelect={handleChangeLower}
              />
              <Form.Item name="mutationLowerLimit" layout="vertical">
                <TextArea />
              </Form.Item>
            </Flex>
          </Card>
          <Form.Item
            name="synchronizeToAlertRules"
            valuePropName="checked"
            style={{ marginTop: '10px' }}
          >
            <Checkbox>同步创建告警规则</Checkbox>
          </Form.Item>
        </Form>
      </div>
    </Modal>
  )
}

export default AlertRuleQuickSet
